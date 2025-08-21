package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"

	"github.com/mbark/shmake/internal/logger"
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
	GlobalCLI CLI = CLI{
		Command: &Command{
			Command: &cli.Command{},
		},
	}
	WaitGroup sync.WaitGroup
)

func ShmakeCLI(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var name, usage string
	if err := starlark.UnpackArgs("cli", args, kwargs,
		"name", &name,
		"usage?", &usage,
	); err != nil {
		return nil, err
	}
	GlobalCLI.Command.Command.Name = name
	GlobalCLI.Command.Command.Usage = usage
	return starlark.None, nil
}

func ShmakeCommand(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var name, help string
	var action starlark.Callable
	var argsList *starlark.List
	var flagsDict *starlark.Dict
	var category string
	if err := starlark.UnpackArgs("command", args, kwargs,
		"name", &name,
		"help?", &help,
		"action", &action,
		"args?", &argsList,
		"flags?", &flagsDict,
		"category?", &category,
	); err != nil {
		return nil, err
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:     name,
			Usage:    help,
			Action:   createCommandAction(name, thread, action),
			Category: category,
		},
	}

	if err := processArgs(argsList, cmd); err != nil {
		return nil, err
	}

	if err := processFlags(flagsDict, cmd); err != nil {
		return nil, err
	}

	GlobalCLI.Command.Command.Commands = append(GlobalCLI.Command.Command.Commands, cmd.Command)
	return starlark.None, nil
}

func ShmakeSubCommand(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var pathList *starlark.List
	var help string
	var action starlark.Callable
	var argsList *starlark.List
	var flagsDict *starlark.Dict
	var category string
	if err := starlark.UnpackArgs("sub_command", args, kwargs,
		"path", &pathList,
		"help?", &help,
		"action", &action,
		"args?", &argsList,
		"flags?", &flagsDict,
		"category?", &category,
	); err != nil {
		return nil, err
	}

	path, err := parsePath(pathList)
	if err != nil {
		return nil, err
	}

	parentCmd, err := findSubCommand(GlobalCLI.Command.Command, path)
	if err != nil {
		return nil, err
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:     path[len(path)-1],
			Usage:    help,
			Action:   createCommandAction(path[len(path)-1], thread, action),
			Category: category,
		},
	}

	if err := processArgs(argsList, cmd); err != nil {
		return nil, err
	}

	if err := processFlags(flagsDict, cmd); err != nil {
		return nil, err
	}

	parentCmd.Commands = append(parentCmd.Commands, cmd.Command)
	return starlark.None, nil
}

// processFlags handles flag configuration for commands.
func processFlags(flagsDict *starlark.Dict, cmd *Command) error {
	if flagsDict == nil {
		return nil
	}

	for _, item := range flagsDict.Items() {
		key, ok := item[0].(starlark.String)
		if !ok {
			return fmt.Errorf("flag name must be string")
		}
		flagDef, ok := item[1].(*starlark.Dict)
		if !ok {
			return fmt.Errorf("flag value must be dict")
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
						return &cli.BoolFlag{
							Name:  string(key),
							Usage: flagHelp,
							Value: bool(defaultVal.(starlark.Bool)),
						}
					}
				case "string":
					builder = func() cli.Flag {
						return &cli.StringFlag{
							Name:  string(key),
							Usage: flagHelp,
							Value: string(defaultVal.(starlark.String)),
						}
					}
				case "int":
					i, _ := defaultVal.(starlark.Int).Int64()
					builder = func() cli.Flag {
						return &cli.IntFlag{
							Name:  string(key),
							Usage: flagHelp,
							Value: int(i),
						}
					}
				default:
					return fmt.Errorf("unknown flag type: %s", typeName)
				}
			case "default":
				defaultVal = kv[1]
			case "help":
				flagHelp = string(kv[1].(starlark.String))
			}
		}

		cmd.Command.Flags = append(cmd.Command.Flags, builder())
	}

	return nil
}

var (
	actionHeader    = lipgloss.NewStyle().Foreground(lipgloss.ANSIColor(ansi.Magenta)).Bold(true)
	argOrFlagHeader = lipgloss.NewStyle().Padding(0, 2).Foreground(lipgloss.ANSIColor(ansi.Cyan))
	argOrFlagStyle  = lipgloss.NewStyle().Faint(true).Padding(0, 4)
)

// createCommandAction creates the action function for a command.
func createCommandAction(
	name string,
	thread *starlark.Thread,
	action starlark.Callable,
) func(context.Context, *cli.Command) error {
	return func(ctx context.Context, command *cli.Command) error {
		logger.Log(actionHeader.Render(name))
		if action == nil {
			return nil
		}

		// The help flag is always defined
		if len(command.Flags) > 1 {
			logger.Log(argOrFlagHeader.Render("Flags"))
		}

		flags := make(starlark.StringDict)
		for _, flag := range command.Flags {
			var sval starlark.Value
			var lval string
			switch val := flag.Get().(type) {
			case string:
				sval = starlark.String(val)
				lval = fmt.Sprintf("'%s'", val)
			case int:
				sval = starlark.MakeInt(val)
				lval = strconv.Itoa(val)
			case bool:
				sval = starlark.Bool(val)
				lval = strconv.FormatBool(val)
			default:
				return fmt.Errorf("unknown flag value type: %v", flag)
			}

			for _, f := range flag.Names() {
				flags[f] = sval
			}
			if flag.Names()[0] != "help" {
				logger.Log(argOrFlagStyle.Render(fmt.Sprintf("%s: %s", strings.Join(flag.Names(), ","), lval)))
			}
		}

		if len(command.Arguments) > 0 {
			logger.Log(argOrFlagHeader.Render("Named arguments"))
		}

		argsDict := make(starlark.StringDict)
		for _, arg := range command.Arguments {
			var argName, lval string
			switch a := arg.(type) {
			case *cli.StringArg:
				argsDict[a.Name] = starlark.String(command.StringArg(a.Name))
				argName, lval = a.Name, fmt.Sprintf("'%s'", command.StringArg(a.Name))
			case *cli.IntArg:
				argsDict[a.Name] = starlark.MakeInt(command.IntArg(a.Name))
				argName, lval = a.Name, strconv.Itoa(command.IntArg(a.Name))
			}
			logger.Log(argOrFlagStyle.Render(fmt.Sprintf("%s: %s", argName, lval)))
		}

		slice := command.Args().Slice()
		list := make([]starlark.Value, len(slice))
		if len(slice) > 0 {
			logger.Log(argOrFlagHeader.Render("Positional arguments"))
		}
		for i, a := range slice {
			list[i] = starlark.String(a)
			logger.Log(argOrFlagStyle.Render(fmt.Sprintf("%d: %s", i, a)))
		}

		_, err := starlark.Call(thread, action, starlark.Tuple{&Context{
			Flags:     flags,
			Args:      argsDict,
			ArgsSlice: starlark.NewList(list),
		}}, nil)
		return err
	}
}

// processArgs handles argument configuration for commands.
func processArgs(argsList *starlark.List, cmd *Command) error {
	if argsList == nil {
		return nil
	}

	for i := 0; i < argsList.Len(); i++ {
		str, ok := argsList.Index(i).(starlark.String)
		if !ok {
			return fmt.Errorf("args must be list of strings")
		}
		cmd.Command.Arguments = append(cmd.Command.Arguments, &cli.StringArg{Name: string(str)})
		cmd.Args = append(cmd.Args, string(str))
	}
	return nil
}

// parsePath converts a Starlark list to a string slice.
func parsePath(pathList *starlark.List) ([]string, error) {
	var path []string
	for i := 0; i < pathList.Len(); i++ {
		str, ok := pathList.Index(i).(starlark.String)
		if !ok {
			return nil, fmt.Errorf("args must be list of strings")
		}
		path = append(path, string(str))
	}
	return path, nil
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

var (
	_ starlark.Value    = (*Context)(nil)
	_ starlark.HasAttrs = (*Context)(nil)
)

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
		return NewFlagMap(c.Flags), nil
	case "args":
		return NewFlagMap(c.Args), nil
	case "args_list":
		return c.ArgsSlice, nil
	}
	return nil, nil
}

func (c *Context) AttrNames() []string {
	return []string{"flags", "args"}
}

var (
	_ starlark.Value   = (*FlagMap)(nil)
	_ starlark.Mapping = (*FlagMap)(nil)
)

// FlagMap allows both dot-access and index-access to flags:
//
//	ctx.flags["some-flag"]  --> works
//	ctx.flags.some_flag      --> also works (mapped from key "some-flag")
type FlagMap struct {
	data      starlark.StringDict
	aliasKeys map[string]string // maps snake_case -> actual key (e.g. some_flag -> some-flag)
}

func (m *FlagMap) Get(value starlark.Value) (v starlark.Value, found bool, err error) {
	key, ok := value.(starlark.String)
	if !ok {
		return starlark.None, false, fmt.Errorf("flag key must be string")
	}

	v, ok = m.data[string(key)]
	return v, ok, nil
}

func NewFlagMap(d starlark.StringDict) *FlagMap {
	alias := map[string]string{}
	for k := range d {
		if isValidIdentifier(k) {
			alias[k] = k
		} else {
			// Convert dash-case to snake_case as fallback
			kSnake := strings.ReplaceAll(k, "-", "_")
			if isValidIdentifier(kSnake) {
				alias[kSnake] = k
			}
		}
	}
	return &FlagMap{
		data:      d,
		aliasKeys: alias,
	}
}

func (m *FlagMap) String() string        { return "<flags>" }
func (m *FlagMap) Type() string          { return "FlagMap" }
func (m *FlagMap) Freeze()               {} // assume values are immutable
func (m *FlagMap) Truth() starlark.Bool  { return starlark.True }
func (m *FlagMap) Hash() (uint32, error) { return 0, errors.New("unhashable") }

// Attr supports ctx.flags.some_flag.
func (m *FlagMap) Attr(name string) (starlark.Value, error) {
	if realKey, ok := m.aliasKeys[name]; ok {
		return m.data[realKey], nil
	}
	return nil, nil
}

func (m *FlagMap) AttrNames() []string {
	names := make([]string, 0, len(m.aliasKeys))
	for k := range m.aliasKeys {
		names = append(names, k)
	}
	return names
}

func (m *FlagMap) Index(i starlark.Value) (starlark.Value, error) {
	key, ok := i.(starlark.String)
	if !ok {
		return nil, fmt.Errorf("flag key must be string")
	}
	val, ok := m.data[string(key)]
	if !ok {
		return starlark.None, nil
	}
	return val, nil
}

func isValidIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}
