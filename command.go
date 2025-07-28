package main

import (
	"log/slog"
	"os"

	"github.com/mbark/devslog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"
	lua "github.com/yuin/gopher-lua"
)

type Shmake struct {
	Commands []*cli.Command
	Runtime  *Runtime
}

const luaShmakeTypeName = "shmake"

func registerShmakeType(L *lua.LState, runtime *Runtime) {
	mt := L.NewTypeMetatable(luaShmakeTypeName)
	L.SetGlobal("shmake", mt)
	L.SetField(mt, "new", L.NewFunction(newShmake(runtime)))
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), map[string]lua.LGFunction{
		"command": shmakeAddCommand,
		"run":     shmakeRunCommand,
	}))
}

func newShmake(runtime *Runtime) func(L *lua.LState) int {
	return func(L *lua.LState) int {
		shmake := &Shmake{Runtime: runtime}
		ud := L.NewUserData()
		ud.Value = shmake
		L.SetMetatable(ud, L.GetTypeMetatable(luaShmakeTypeName))
		L.Push(ud)
		return 1
	}
}

func isShmake(L *lua.LState) *Shmake {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Shmake); ok {
		return v
	}
	L.ArgError(1, "shmake expected")
	return nil
}

type commandOptions struct {
	Usage string
}

func shmakeAddCommand(L *lua.LState) int {
	p := isShmake(L)
	name := L.CheckString(2)
	action := L.CheckFunction(3)

	var options commandOptions
	err := MapTable(4, L.Get(4), &options)
	if err != nil {
		L.RaiseError("invalid options: %v", err)
	}

	p.Commands = append(p.Commands, &cli.Command{
		Name:  name,
		Usage: options.Usage,
		Action: func(ctx *cli.Context) error {
			L.SetContext(ctx.Context)
			return L.CallByParam(lua.P{Fn: action, NRet: 1, Protect: true})
		},
	})

	return 0
}

func shmakeRunCommand(L *lua.LState) int {
	p := isShmake(L)

	var verbose, noCache bool
	cliFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "print logs to stdout",
			Destination: &verbose,
		},
		&cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "ignore stored values in the cache",
			Destination: &noCache,
		},
	}

	app := &cli.App{
		Name:     "shmake",
		Usage:    "make shmake",
		Version:  version,
		Flags:    cliFlags,
		Commands: p.Commands,
		Before: func(ctx *cli.Context) error {
			if verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
				opts := &slog.HandlerOptions{Level: slog.LevelDebug}
				p.Runtime.logger = slog.New(slogmulti.Fanout(
					slog.NewJSONHandler(p.Runtime.logFile, opts),
					devslog.NewHandler(os.Stdout, &devslog.Options{HandlerOptions: opts}),
				))
			}
			if noCache {
				p.Runtime.cache.ForceOutOfDate = noCache
			}
			return nil
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		L.RaiseError("failed to run shmake: %v", err)
	}

	return 0
}
