package internal

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"

	"github.com/mbark/sindr/internal/logger"
)

type CLI struct {
	Command *Command
}

type Command struct {
	Command *cli.Command
	Args    []string
	ArgType map[string]func(string) (starlark.Value, error)
}

func SindrCLI(
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

	sindrCLI, err := getSindrCLI(thread)
	if err != nil {
		return nil, err
	}

	sindrCLI.Command.Command.Name = name
	sindrCLI.Command.Command.Usage = usage
	return starlark.None, nil
}

func SindrCommand(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var name, usage string
	var action starlark.Callable
	var argsList *starlark.List
	var flagsList *starlark.List
	var category string
	if err := starlark.UnpackArgs("command", args, kwargs,
		"name", &name,
		"usage?", &usage,
		"action", &action,
		"args?", &argsList,
		"flags?", &flagsList,
		"category?", &category,
	); err != nil {
		return nil, err
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:     name,
			Usage:    usage,
			Action:   createCommandAction(name, thread, action),
			Category: category,
		},
	}

	if err := processArgs(argsList, cmd); err != nil {
		return nil, err
	}

	if err := processFlags(flagsList, cmd); err != nil {
		return nil, err
	}

	sindrCLI, err := getSindrCLI(thread)
	if err != nil {
		return nil, err
	}

	sindrCLI.Command.Command.Commands = append(sindrCLI.Command.Command.Commands, cmd.Command)
	return starlark.None, nil
}

func SindrSubCommand(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var pathList *starlark.List
	var usage string
	var action starlark.Callable
	var argsList *starlark.List
	var flagsList *starlark.List
	var category string
	if err := starlark.UnpackArgs("sub_command", args, kwargs,
		"path", &pathList,
		"usage?", &usage,
		"action", &action,
		"args?", &argsList,
		"flags?", &flagsList,
		"category?", &category,
	); err != nil {
		return nil, err
	}

	path, err := fromList[string](
		pathList,
		func(v starlark.Value) (string, error) { return castString(v) },
	)
	if err != nil {
		return nil, err
	}

	sindrCLI, err := getSindrCLI(thread)
	if err != nil {
		return nil, err
	}

	parentCmd, err := findSubCommand(sindrCLI.Command.Command, path)
	if err != nil {
		return nil, err
	}

	cmd := &Command{
		Command: &cli.Command{
			Name:     path[len(path)-1],
			Usage:    usage,
			Action:   createCommandAction(path[len(path)-1], thread, action),
			Category: category,
		},
	}

	if err := processArgs(argsList, cmd); err != nil {
		return nil, err
	}

	if err := processFlags(flagsList, cmd); err != nil {
		return nil, err
	}

	parentCmd.Commands = append(parentCmd.Commands, cmd.Command)
	return starlark.None, nil
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
		logger.WithStack(thread.CallStack()).Log(actionHeader.Render(name))
		if action == nil {
			return nil
		}

		// The help flag is always defined
		if len(command.Flags) > 1 {
			logger.WithStack(thread.CallStack()).Log(argOrFlagHeader.Render("Flags"))
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
			case []string:
				sval = toList(val, func(s string) starlark.Value { return starlark.String(s) })
				lval = "[" + strings.Join(val, ", ") + "]"
			case []int:
				sval = toList(val, func(i int) starlark.Value { return starlark.MakeInt(i) })
				lval = "[" + strings.Join(mapList(val, func(i int) string {
					return strconv.Itoa(i)
				}), ", ") + "]"

			default:
				return fmt.Errorf("unknown flag value type for '%s': %T", flag.Names()[0], flag.Get())
			}

			for _, f := range flag.Names() {
				flags[f] = sval
			}
			if flag.Names()[0] != "help" {
				logger.WithStack(thread.CallStack()).Log(
					argOrFlagStyle.Render(
						fmt.Sprintf("%s: %s", strings.Join(flag.Names(), ","), lval),
					),
				)
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
			logger.WithStack(thread.CallStack()).
				Log(argOrFlagStyle.Render(fmt.Sprintf("%s: %s", argName, lval)))
		}

		slice := command.Args().Slice()
		list := make([]starlark.Value, len(slice))
		if len(slice) > 0 {
			logger.WithStack(thread.CallStack()).Log(argOrFlagHeader.Render("Positional arguments"))
		}
		for i, a := range slice {
			list[i] = starlark.String(a)
			logger.WithStack(thread.CallStack()).
				Log(argOrFlagStyle.Render(fmt.Sprintf("%d: %s", i, a)))
		}

		c := NewContext(flags, argsDict, starlark.NewList(list))
		thread.SetLocal("ctx", c)

		_, err := starlark.Call(thread, action, starlark.Tuple{c}, nil)
		return err
	}
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
	Flags        *FlagMap
	Args         *FlagMap
	FlagsAndArgs *FlagMap
	ArgsSlice    *starlark.List
}

func NewContext(flags, args starlark.StringDict, argsSlice *starlark.List) *Context {
	return &Context{
		Flags:        NewFlagMap(flags),
		Args:         NewFlagMap(args),
		FlagsAndArgs: NewFlagMap(union(flags, args)),
		ArgsSlice:    argsSlice,
	}
}

func (c *Context) String() string        { return "<ctx>" }
func (c *Context) Type() string          { return "Context" }
func (c *Context) Freeze()               {}
func (c *Context) Truth() starlark.Bool  { return starlark.True }
func (c *Context) Hash() (uint32, error) { return 0, errors.New("unhashable") }
func (c *Context) Attr(name string) (starlark.Value, error) {
	value, err := c.FlagsAndArgs.Attr(name)
	if err != nil {
		return nil, err
	}
	if value != nil {
		return value, nil
	}

	switch name {
	case "flags":
		return c.Flags, nil
	case "args":
		return c.Args, nil
	case "args_list":
		return c.ArgsSlice, nil
	default:
		return nil, nil
	}
}

func (c *Context) AttrNames() []string {
	return []string{"flags", "args", "args_list"}
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

func (m *FlagMap) Get(value starlark.Value) (v starlark.Value, found bool, err error) {
	key, ok := value.(starlark.String)
	if !ok {
		return starlark.None, false, fmt.Errorf("flag key must be string")
	}

	v, ok = m.data[string(key)]
	return v, ok, nil
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
