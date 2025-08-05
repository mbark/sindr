package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/mbark/devslog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v3"
	lua "github.com/yuin/gopher-lua"
)

// Shmake global to create the cli instance.
type Shmake struct {
	Command *Command
	Runtime *Runtime
}

// Command is the struct we use when building a command; it's essentially just a cli.Command wrapper but we use it for
// now in case we want to add more data in the future.
type Command struct {
	Command *cli.Command
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

func newCommand(r *Runtime, L *lua.LState) ([]lua.LValue, error) {
	shmake := &Shmake{Runtime: r}
	name := L.CheckString(1)

	var options commandOptions
	if L.GetTop() >= 2 {
		err := MapTable(2, L.Get(2), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	shmake.Command = &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: options.Usage,
		},
	}

	ud := L.NewUserData()
	ud.Value = shmake
	L.SetMetatable(ud, L.GetTypeMetatable(ShmakeType{}.TypeName()))
	return []lua.LValue{ud}, nil
}

func (s ShmakeType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command":     s.Command,
		"sub_command": s.SubCommand,
		"run":         s.Run,
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

	cmd := &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: options.Usage,
		},
	}
	parent.Command.Command.Commands = append(parent.Command.Command.Commands, cmd.Command)
	return NewUserData(L, cmd, CommandType{})
}

func (s ShmakeType) SubCommand(L *lua.LState) int {
	root := IsUserData[*Shmake](L)
	name, err := MapArray[string](2, L.Get(2))
	if err != nil {
		L.RaiseError("invalid sub command path: %v", err)
	}

	var options commandOptions
	if L.GetTop() > 2 {
		err := MapTable(3, L.Get(3), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	parent, err := findSubCommand(root.Command.Command, name)
	if err != nil {
		L.RaiseError("unable to find sub command for path [%s]: %v", strings.Join(name, ","), err)
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:  name[len(name)-1],
			Usage: options.Usage,
		},
	}
	parent.Commands = append(parent.Commands, cmd.Command)
	return NewUserData(L, cmd, CommandType{})
}

func findSubCommand(cmd *cli.Command, path []string) (*cli.Command, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("empty path")
	}
	// the last one is the name of the new command
	if len(path) == 1 {
		return cmd, nil
	}

	for _, c := range cmd.Commands {
		if c.Name == path[0] {
			return findSubCommand(c, path[1:])
		}
	}

	return nil, fmt.Errorf("no command with name %s found", path[0])
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
	cmd.Command.Version = version
	cmd.Command.Flags = append(cmd.Command.Flags, cliFlags...)
	cmd.Command.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
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

	err := cmd.Command.Run(L.Context(), os.Args)
	if err != nil {
		L.RaiseError("failed to shell shmake: %v", err)
	}

	s.Runtime.wg.Wait()
	return 0
}

func (c CommandType) TypeName() string {
	return "command"
}

func (c CommandType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command":     c.Command,
		"action":      c.Action,
		"flag":        c.StringFlag,
		"string_flag": c.StringFlag,
		"int_flag":    c.IntFlag,
		"bool_flag":   c.BoolFlag,
		"arg":         c.StringArg,
		"string_arg":  c.StringArg,
		"int_arg":     c.IntArg,
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
	cmd := IsUserData[*Command](L)
	name := L.CheckString(2)

	flag := mapFlagOptions[string](L)
	cmd.Command.Flags = append(cmd.Command.Flags, &cli.StringFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) IntFlag(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	name := L.CheckString(2)

	flag := mapFlagOptions[int](L)
	cmd.Command.Flags = append(cmd.Command.Flags, &cli.IntFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) BoolFlag(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	name := L.CheckString(2)

	flag := mapFlagOptions[bool](L)
	cmd.Command.Flags = append(cmd.Command.Flags, &cli.BoolFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(L, cmd, CommandType{})
}

type argOptions[T any] struct {
	Default T
	Usage   string
}

func mapArgOptions[T any](L *lua.LState) argOptions[T] {
	var arg argOptions[T]
	if L.GetTop() < 3 {
		return arg
	}

	err := MapTable(3, L.Get(3), &arg)
	if err != nil {
		L.RaiseError("invalid options: %v", err)
	}

	return arg
}

func (c CommandType) StringArg(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	name := L.CheckString(2)

	flag := mapArgOptions[string](L)
	cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.StringArg{
		Name:      name,
		Value:     flag.Default,
		UsageText: flag.Usage,
	})
	return NewUserData(L, cmd, CommandType{})
}
func (c CommandType) IntArg(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	name := L.CheckString(2)

	flag := mapArgOptions[int](L)
	cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.IntArg{
		Name:      name,
		Value:     flag.Default,
		UsageText: flag.Usage,
	})
	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) Action(L *lua.LState) int {
	cmd := IsUserData[*Command](L)
	action := L.CheckFunction(2)

	cmd.Command.Action = func(ctx context.Context, command *cli.Command) error {
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
			L.RawSet(tbl, lua.LString(flag.Names()[0]), lval)
		}

		args := L.NewTable()
		for i, arg := range command.Arguments {
			var lval lua.LValue
			switch val := arg.Get().(type) {
			case string:
				lval = lua.LString(val)
			case int:
				lval = lua.LNumber(val)
			case bool:
				lval = lua.LBool(val)
			default:
				L.RaiseError("unknown flag value type: %v", arg)
			}
			L.RawSetInt(args, i+1, lval)
		}

		return L.CallByParam(lua.P{Fn: action, NRet: 1, Protect: true}, tbl, args)
	}

	return NewUserData(L, cmd, CommandType{})
}

func (c CommandType) Command(L *lua.LState) int {
	parent := IsUserData[*Command](L)
	name := L.CheckString(2)

	var options commandOptions
	if L.GetTop() > 2 {
		err := MapTable(3, L.Get(3), &options)
		if err != nil {
			L.RaiseError("invalid options: %v", err)
		}
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: options.Usage,
		},
	}
	parent.Command.Commands = append(parent.Command.Commands, cmd.Command)
	return NewUserData(L, cmd, CommandType{})
}

type commandOptions struct {
	Usage string
}
