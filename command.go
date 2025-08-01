package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/mbark/devslog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v3"
	lua "github.com/yuin/gopher-lua"
)

// Shmake global to create the cli instance.
type Shmake struct {
	Command *cli.Command
	Runtime *Runtime
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
	shmake := &Shmake{Runtime: s.Runtime}
	name := L.CheckString(1)

	var options commandOptions
	if L.GetTop() >= 2 {
		err := MapTable(2, L.Get(2), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	shmake.Command = &cli.Command{
		Name:  name,
		Usage: options.Usage,
	}
	return NewUserData(L, shmake, ShmakeType{})
}

func (s ShmakeType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command": s.Command,
		"run":     s.Run,
	}
}

func (s ShmakeType) Command(L *lua.LState) int {
	parent := IsUserData[*Shmake](L)
	name := L.CheckString(2)

	var options commandOptions
	if L.GetTop() > 2 {
		err := MapTable(3, L.Get(3), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	cmd := &cli.Command{
		Name:  name,
		Usage: options.Usage,
	}
	parent.Command.Commands = append(parent.Command.Commands, cmd)
	return NewUserData(L, cmd, CommandType{})
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

	cmd := shmake.Command
	cmd.Version = version
	cmd.Flags = append(cmd.Flags, cliFlags...)
	cmd.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
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
		return ctx, nil
	}

	err := cmd.Run(L.Context(), os.Args)
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
	return NewUserData(L, &cli.Command{}, CommandType{})
}

func (c CommandType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command":     c.Command,
		"action":      c.Action,
		"flag":        c.StringFlag,
		"string_flag": c.StringFlag,
		"int_flag":    c.IntFlag,
		"bool_flag":   c.BoolFlag,
	}
}

type flagOptions[T any] struct {
	Default  T
	Usage    string
	Required bool
}

func mapFlagOptions[T any](L *lua.LState) flagOptions[T] {
	var flag flagOptions[T]
	if L.GetTop() < 3 {
		return flag
	}

	err := MapTable(3, L.Get(3), &flag)
	if err != nil {
		L.RaiseError("invalid options: %v", err)
	}

	return flag
}

func (c CommandType) StringFlag(L *lua.LState) int {
	cmd := IsUserData[*cli.Command](L)
	name := L.CheckString(2)

	flag := mapFlagOptions[string](L)
	cmd.Flags = append(cmd.Flags, &cli.StringFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) IntFlag(L *lua.LState) int {
	cmd := IsUserData[*cli.Command](L)
	name := L.CheckString(2)

	flag := mapFlagOptions[int](L)
	cmd.Flags = append(cmd.Flags, &cli.IntFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) BoolFlag(L *lua.LState) int {
	cmd := IsUserData[*cli.Command](L)
	name := L.CheckString(2)

	flag := mapFlagOptions[bool](L)
	cmd.Flags = append(cmd.Flags, &cli.BoolFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) Action(L *lua.LState) int {
	cmd := IsUserData[*cli.Command](L)
	action := L.CheckFunction(2)

	cmd.Action = func(ctx context.Context, command *cli.Command) error {
		L.SetContext(ctx)

		tbl := L.NewTable()
		for _, flag := range command.Flags {
			var lval lua.LValue
			switch val := flag.Get().(type) {
			case string:
				lval = lua.LString(val)
			case int:
				lval = lua.LNumber(val)
			case bool:
				lval = lua.LBool(val)
			default:
				L.RaiseError("unknown flag value type: %v", flag)
			}
			L.SetField(tbl, flag.Names()[0], lval)
		}
		return L.CallByParam(lua.P{Fn: action, NRet: 1, Protect: true}, tbl)
	}
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) Command(L *lua.LState) int {
	cmd := IsUserData[*cli.Command](L)
	name := L.CheckString(2)

	var options commandOptions
	if L.GetTop() > 2 {
		err := MapTable(3, L.Get(3), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	cmd.Name = name
	cmd.Usage = options.Usage
	return NewUserData(L, cmd, CommandType{})
}

type commandOptions struct {
	Usage string
}
