package main

import (
	"log/slog"
	"os"

	"github.com/mbark/devslog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"
	lua "github.com/yuin/gopher-lua"
)

// Shmake global to create the cli instance.
type Shmake struct {
	Commands []*cli.Command
	Runtime  *Runtime
}

var (
	_ LuaType = ShmakeType{}
	_ LuaType = CommandType{}
)

type ShmakeType struct {
	Runtime *Runtime
}
type CommandType struct {
	Runtime *Runtime
}

func (s ShmakeType) TypeName() string {
	return "shmake"
}

func (s ShmakeType) GlobalName() string {
	return "shmake"
}

func (s ShmakeType) New(L *lua.LState) int {
	return NewUserData(L, &Shmake{Runtime: s.Runtime}, ShmakeType{})
}

func (s ShmakeType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command": s.Command,
		"run":     s.Run,
	}
}

func (s ShmakeType) Command(L *lua.LState) int {
	shmake := IsUserData[*Shmake](L) // keep for now
	name := L.CheckString(2)

	var options commandOptions
	if L.GetTop() > 2 {
		err := MapTable(3, L.Get(3), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	return NewUserData(L, &Command{
		FlagValues: map[string]string{},
		Name:       name,
		Options:    options,
		Shmake:     shmake,
	}, CommandType{})
}

func (s ShmakeType) Run(L *lua.LState) int {
	shmake := IsUserData[*Shmake](L)

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
		Commands: shmake.Commands,
		Before: func(ctx *cli.Context) error {
			if verbose {
				slog.SetLogLoggerLevel(slog.LevelDebug)
				opts := &slog.HandlerOptions{Level: slog.LevelDebug}
				shmake.Runtime.logger = slog.New(slogmulti.Fanout(
					slog.NewJSONHandler(shmake.Runtime.logFile, opts),
					devslog.NewHandler(os.Stdout, &devslog.Options{HandlerOptions: opts}),
				))
			}
			if noCache {
				shmake.Runtime.cache.ForceOutOfDate = noCache
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

func (c CommandType) TypeName() string {
	return "command"
}

func (c CommandType) GlobalName() string {
	return "command"
}

func (c CommandType) New(L *lua.LState) int {
	return NewUserData(L, &Command{}, CommandType{})
}

func (c CommandType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"action": c.Action,
		"flag":   c.StringFlag,
	}
}

type stringFlag struct {
	Default string
	Usage   string
}

func (c CommandType) StringFlag(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	name := L.CheckString(2)

	var flag stringFlag
	err := MapTable(3, L.Get(3), &flag)
	if err != nil {
		L.RaiseError("invalid options: %v", err)
	}

	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:        name,
		DefaultText: flag.Default,
		Usage:       flag.Usage,
		Action: func(ctx *cli.Context, s string) error {
			cmd.FlagValues[name] = s
			return nil
		},
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) Action(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	action := L.CheckFunction(2)

	cmd.Shmake.Commands = append(cmd.Shmake.Commands, &cli.Command{
		Name:  cmd.Name,
		Usage: cmd.Options.Usage,
		Flags: cmd.Flags,
		Action: func(ctx *cli.Context) error {
			L.SetContext(ctx.Context)

			tbl := L.NewTable()
			for name, value := range cmd.FlagValues {
				L.SetField(tbl, name, lua.LString(value))
			}
			return L.CallByParam(lua.P{Fn: action, NRet: 1, Protect: true}, tbl)
		},
	})

	return 0
}

type commandOptions struct {
	Usage string
}

type Command struct {
	Flags      []cli.Flag
	FlagValues map[string]string
	Name       string
	Action     lua.LGFunction

	Options commandOptions

	Shmake *Shmake
}
