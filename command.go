package shmake

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/charmbracelet/log"
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

func newCommand(r *Runtime, l *lua.LState) ([]lua.LValue, error) {
	shmake := &Shmake{Runtime: r}
	name := l.CheckString(1)

	options, err := MapOptionalTable[commandOptions](l, 2)
	if err != nil {
		return nil, err
	}

	shmake.Command = &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: options.Usage,
		},
	}

	ud := l.NewUserData()
	ud.Value = shmake
	l.SetMetatable(ud, l.GetTypeMetatable(ShmakeType{}.TypeName()))
	return []lua.LValue{ud}, nil
}

func (s ShmakeType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"command":     s.Command,
		"sub_command": s.SubCommand,
		"run":         s.Run,
	}
}

func (s ShmakeType) Command(l *lua.LState) int {
	parent := IsUserData[*Shmake](l)
	name := l.CheckString(2)

	options, err := MapOptionalTable[commandOptions](l, 3)
	if err != nil {
		l.RaiseError("invalid options: %v", err)
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: options.Usage,
		},
	}
	parent.Command.Command.Commands = append(parent.Command.Command.Commands, cmd.Command)
	return NewUserData(l, cmd, CommandType{})
}

func (s ShmakeType) SubCommand(l *lua.LState) int {
	root := IsUserData[*Shmake](l)
	name, err := MapArray[string](l, 2)
	if err != nil {
		l.RaiseError("invalid sub command path: %v", err)
	}

	options, err := MapOptionalTable[commandOptions](l, 3)
	if err != nil {
		l.RaiseError("invalid options: %v", err)
	}

	parent, err := findSubCommand(root.Command.Command, name)
	if err != nil {
		l.RaiseError("unable to find sub command for path [%s]: %v", strings.Join(name, ","), err)
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:  name[len(name)-1],
			Usage: options.Usage,
		},
	}
	parent.Commands = append(parent.Commands, cmd.Command)
	return NewUserData(l, cmd, CommandType{})
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

func (s ShmakeType) Run(l *lua.LState) int {
	shmake := IsUserData[*Shmake](l)

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
			slog.SetDefault(slog.New(slogmulti.Fanout(
				slog.NewJSONHandler(
					shmake.Runtime.logFile,
					&slog.HandlerOptions{Level: slog.LevelDebug},
				),
				log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel}),
			)))
		}
		if noCache {
			shmake.Runtime.Cache.ForceOutOfDate = noCache
		}
		return ctx, nil
	}

	err := cmd.Command.Run(l.Context(), s.Runtime.Args)
	if err != nil {
		l.RaiseError("%s", err.Error())
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

func mapFlagOptions[T any](l *lua.LState) flagOptions[T] {
	flag, err := MapOptionalTable[flagOptions[T]](l, 3)
	if err != nil {
		l.RaiseError("invalid options: %v", err)
	}

	return flag
}

func (c CommandType) StringFlag(l *lua.LState) int {
	cmd := IsUserData[*Command](l)
	name := l.CheckString(2)

	flag := mapFlagOptions[string](l)
	cmd.Command.Flags = append(cmd.Command.Flags, &cli.StringFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(l, cmd, CommandType{})
}

func (c CommandType) IntFlag(l *lua.LState) int {
	cmd := IsUserData[*Command](l)
	name := l.CheckString(2)

	flag := mapFlagOptions[int](l)
	cmd.Command.Flags = append(cmd.Command.Flags, &cli.IntFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(l, cmd, CommandType{})
}

func (c CommandType) BoolFlag(l *lua.LState) int {
	cmd := IsUserData[*Command](l)
	name := l.CheckString(2)

	flag := mapFlagOptions[bool](l)
	cmd.Command.Flags = append(cmd.Command.Flags, &cli.BoolFlag{
		Name:     name,
		Usage:    flag.Usage,
		Value:    flag.Default,
		Required: flag.Required,
	})
	return NewUserData(l, cmd, CommandType{})
}

type argOptions[T any] struct {
	Default T
	Usage   string
}

func mapArgOptions[T any](l *lua.LState) argOptions[T] {
	arg, err := MapOptionalTable[argOptions[T]](l, 3)
	if err != nil {
		l.RaiseError("invalid options: %v", err)
	}

	return arg
}

func (c CommandType) StringArg(l *lua.LState) int {
	cmd := IsUserData[*Command](l)
	name := l.CheckString(2)

	flag := mapArgOptions[string](l)
	cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.StringArg{
		Name:      name,
		Value:     flag.Default,
		UsageText: flag.Usage,
	})
	return NewUserData(l, cmd, CommandType{})
}

func (c CommandType) IntArg(l *lua.LState) int {
	cmd := IsUserData[*Command](l)
	name := l.CheckString(2)

	flag := mapArgOptions[int](l)
	cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.IntArg{
		Name:      name,
		Value:     flag.Default,
		UsageText: flag.Usage,
	})
	return NewUserData(l, cmd, CommandType{})
}

func (c CommandType) Action(l *lua.LState) int {
	cmd := IsUserData[*Command](l)
	action := l.CheckFunction(2)

	cmd.Command.Action = func(ctx context.Context, command *cli.Command) error {
		l.SetContext(ctx)

		tbl := l.NewTable()
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
				l.RaiseError("unknown flag value type: %v", flag)
			}
			l.RawSet(tbl, lua.LString(flag.Names()[0]), lval)
		}

		args := l.NewTable()
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
				l.RaiseError("unknown flag value type: %v", arg)
			}
			l.RawSetInt(args, i+1, lval)
		}

		return l.CallByParam(lua.P{Fn: action, NRet: 1, Protect: true}, tbl, args)
	}

	return NewUserData(l, cmd, CommandType{})
}

func (c CommandType) Command(l *lua.LState) int {
	parent := IsUserData[*Command](l)
	name := l.CheckString(2)

	options, err := MapOptionalTable[commandOptions](l, 3)
	if err != nil {
		l.RaiseError("invalid options: %v", err)
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: options.Usage,
		},
	}
	parent.Command.Commands = append(parent.Command.Commands, cmd.Command)
	return NewUserData(l, cmd, CommandType{})
}

type commandOptions struct {
	Usage string
}
