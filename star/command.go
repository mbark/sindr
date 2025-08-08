package star

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type CLI struct {
	Command *Command
}

type Command struct {
	Command *cli.Command
	Args    []string
	ArgType map[string]func(string) (starlark.Value, error)
}

var (
	gCLI           CLI
	forceOutOfDate bool
	wg             sync.WaitGroup
)

func shmakeCLI(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, usage string
	if err := starlark.UnpackArgs("cli", args, kwargs,
		"name", &name,
		"usage?", &usage,
	); err != nil {
		return nil, err
	}
	gCLI = CLI{
		Command: &Command{
			Command: &cli.Command{
				Name:  name,
				Usage: usage,
			},
		},
	}
	return starlark.None, nil
}

func shmakeCommand(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name, help string
	var action starlark.Callable
	var argsList *starlark.List
	var flagsDict *starlark.Dict
	if err := starlark.UnpackArgs("command", args, kwargs,
		"name", &name,
		"help?", &help,
		"action", &action,
		"args?", &argsList,
		"flags?", &flagsDict,
	); err != nil {
		return nil, err
	}
	cmd := &Command{
		Command: &cli.Command{
			Name:  name,
			Usage: help,
			Action: func(ctx context.Context, command *cli.Command) error {
				flags := make(starlark.StringDict)
				for _, flag := range command.Flags {
					var sval starlark.Value
					switch val := flag.Get().(type) {
					case string:
						sval = starlark.String(val)
					case int:
						sval = starlark.MakeInt(val)
					case bool:
						sval = starlark.Bool(val)
					default:
						return fmt.Errorf("unknown flag value type: %v", flag)
					}
					for _, f := range flag.Names() {
						flags[f] = sval
					}
				}

				argsDict := make(starlark.StringDict)
				for _, arg := range command.Arguments {
					switch a := arg.(type) {
					case *cli.StringArg:
						argsDict[a.Name] = starlark.String(command.StringArg(a.Name))
					case *cli.IntArg:
						argsDict[a.Name] = starlark.MakeInt(command.IntArg(a.Name))
					}
				}

				slice := command.Args().Slice()
				list := make([]starlark.Value, len(slice))
				for i, a := range slice {
					list[i] = starlark.String(a)
				}

				_, err := starlark.Call(thread, action, starlark.Tuple{&Context{
					Flags:     flags,
					Args:      argsDict,
					ArgsSlice: starlark.NewList(list),
				}}, nil)
				return err
			},
		},
	}

	if argsList != nil {
		for i := 0; i < argsList.Len(); i++ {
			str, ok := argsList.Index(i).(starlark.String)
			if !ok {
				return nil, fmt.Errorf("args must be list of strings")
			}
			cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.StringArg{Name: string(str)})
			cmd.Args = append(cmd.Args, string(str))
		}
	}

	if flagsDict != nil {
		for _, item := range flagsDict.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("flag name must be string")
			}
			flagDef, ok := item[1].(*starlark.Dict)
			if !ok {
				return nil, fmt.Errorf("flag value must be dict")
			}

			var defaultVal starlark.Value
			var flagHelp string
			var builder func() cli.Flag
			for _, kv := range flagDef.Items() {
				k := string(kv[0].(starlark.String))
				switch k {
				case "type":
					typeName := string(kv[1].(starlark.String))
					switch typeName {
					case "bool":
						builder = func() cli.Flag {
							return &cli.BoolFlag{Name: string(key), Usage: flagHelp, Value: bool(defaultVal.(starlark.Bool))}
						}
					case "string":
						builder = func() cli.Flag {
							return &cli.StringFlag{Name: string(key), Usage: flagHelp, Value: string(defaultVal.(starlark.String))}
						}
					case "int":
						i, _ := defaultVal.(starlark.Int).Int64()
						builder = func() cli.Flag { return &cli.IntFlag{Name: string(key), Usage: flagHelp, Value: int(i)} }

					default:
						return nil, fmt.Errorf("unknown flag type: %s", typeName)
					}

				case "default":
					defaultVal = kv[1]

				case "help":
					flagHelp = string(kv[1].(starlark.String))
				}
			}

			cmd.Command.Flags = append(cmd.Command.Flags, builder())
		}
	}

	gCLI.Command.Command.Commands = append(gCLI.Command.Command.Commands, cmd.Command)
	return starlark.None, nil
}

func shmakeSubCommand(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var pathList *starlark.List
	var help string
	var action starlark.Callable
	var argsList *starlark.List
	var flagsDict *starlark.Dict
	if err := starlark.UnpackArgs("sub_command", args, kwargs,
		"path", &pathList,
		"help?", &help,
		"action", &action,
		"args?", &argsList,
		"flags?", &flagsDict,
	); err != nil {
		return nil, err
	}

	var path []string
	for i := 0; i < pathList.Len(); i++ {
		str, ok := pathList.Index(i).(starlark.String)
		if !ok {
			return nil, fmt.Errorf("args must be list of strings")
		}
		path = append(path, string(str))
	}

	parentCmd, err := findSubCommand(gCLI.Command.Command, path)
	if err != nil {
		return nil, err
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:  path[len(path)-1],
			Usage: help,
			Action: func(ctx context.Context, command *cli.Command) error {
				flags := make(starlark.StringDict)
				for _, flag := range command.Flags {
					var sval starlark.Value
					switch val := flag.Get().(type) {
					case string:
						sval = starlark.String(val)
					case int:
						sval = starlark.MakeInt(val)
					case bool:
						sval = starlark.Bool(val)
					default:
						return fmt.Errorf("unknown flag value type: %v", flag)
					}
					for _, f := range flag.Names() {
						flags[f] = sval
					}
				}

				argsDict := make(starlark.StringDict)
				for _, arg := range command.Arguments {
					switch a := arg.(type) {
					case *cli.StringArg:
						argsDict[a.Name] = starlark.String(command.StringArg(a.Name))
					case *cli.IntArg:
						argsDict[a.Name] = starlark.MakeInt(command.IntArg(a.Name))
					}
				}

				slice := command.Args().Slice()
				list := make([]starlark.Value, len(slice))
				for i, a := range slice {
					list[i] = starlark.String(a)
				}

				_, err := starlark.Call(thread, action, starlark.Tuple{&Context{
					Flags:     flags,
					Args:      argsDict,
					ArgsSlice: starlark.NewList(list),
				}}, nil)
				return err
			},
		},
	}

	if argsList != nil {
		for i := 0; i < argsList.Len(); i++ {
			str, ok := argsList.Index(i).(starlark.String)
			if !ok {
				return nil, fmt.Errorf("args must be list of strings")
			}
			cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.StringArg{Name: string(str)})
			cmd.Args = append(cmd.Args, string(str))
		}
	}

	if flagsDict != nil {
		for _, item := range flagsDict.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("flag name must be string")
			}
			flagDef, ok := item[1].(*starlark.Dict)
			if !ok {
				return nil, fmt.Errorf("flag value must be dict")
			}

			var defaultVal starlark.Value
			var flagHelp string
			var builder func() cli.Flag
			for _, kv := range flagDef.Items() {
				k := string(kv[0].(starlark.String))
				switch k {
				case "type":
					typeName := string(kv[1].(starlark.String))
					switch typeName {
					case "bool":
						builder = func() cli.Flag {
							return &cli.BoolFlag{Name: string(key), Usage: flagHelp, Value: bool(defaultVal.(starlark.Bool))}
						}
					case "string":
						builder = func() cli.Flag {
							return &cli.StringFlag{Name: string(key), Usage: flagHelp, Value: string(defaultVal.(starlark.String))}
						}
					case "int":
						i, _ := defaultVal.(starlark.Int).Int64()
						builder = func() cli.Flag { return &cli.IntFlag{Name: string(key), Usage: flagHelp, Value: int(i)} }

					default:
						return nil, fmt.Errorf("unknown flag type: %s", typeName)
					}

				case "default":
					defaultVal = kv[1]

				case "help":
					flagHelp = string(kv[1].(starlark.String))
				}
			}

			cmd.Command.Flags = append(cmd.Command.Flags, builder())
		}
	}

	parentCmd.Commands = append(parentCmd.Commands, cmd.Command)
	return starlark.None, nil
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

type Context struct {
	Flags     starlark.StringDict
	Args      starlark.StringDict
	ArgsSlice *starlark.List
}

func (c *Context) String() string        { return "<ctx>" }
func (c *Context) Type() string          { return "Context" }
func (c *Context) Freeze()               {}
func (c *Context) Truth() starlark.Bool  { return starlark.True }
func (c *Context) Hash() (uint32, error) { return 0, errors.New("unhashable") }
func (c *Context) Attr(name string) (starlark.Value, error) {
	switch name {
	case "flags":
		return starlarkstruct.FromStringDict(starlark.String("flags"), c.Flags), nil
	case "args":
		return starlarkstruct.FromStringDict(starlark.String("args"), c.Args), nil
	case "args_list":
		return c.ArgsSlice, nil
	}
	return nil, nil
}

func (c *Context) AttrNames() []string {
	return []string{"flags", "args"}
}
