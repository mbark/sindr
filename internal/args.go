package internal

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
)

func SindrStringArg(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrArg[string](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue string) (cli.Argument, error) {
			return &cli.StringArg{
				Name:      name,
				UsageText: usage,
				Value:     defaultValue,
			}, nil
		},
	)
}

func SindrIntArg(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrArg[int](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue int) (cli.Argument, error) {
			return &cli.IntArg{
				Name:      name,
				UsageText: usage,
				Value:     defaultValue,
			}, nil
		},
	)
}

func sindrArg[T any](
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
	newArg func(name, usage string, defaultValue T) (cli.Argument, error),
) (starlark.Value, error) {
	var name, usage string
	var defaultValue T
	if err := starlark.UnpackArgs("string_flag", args, kwargs,
		"name", &name,
		"usage?", &usage,
		"default?", &defaultValue,
	); err != nil {
		return nil, err
	}

	f, err := newArg(name, usage, defaultValue)
	if err != nil {
		return nil, err
	}
	return NewArg(name, f), nil
}

func processArgs(argsList *starlark.List, cmd *Command) error {
	if argsList == nil {
		return nil
	}

	for e := range argsList.Elements() {
		arg, ok := e.(*Arg)
		if !ok {
			return fmt.Errorf("expected Arg, got: %T", e)
		}

		cmd.Command.Arguments = append(cmd.Command.Arguments, arg.arg)
		cmd.Args = append(cmd.Args, arg.name)
	}

	return nil
}

var _ starlark.Value = (*Arg)(nil)

type Arg struct {
	name string
	arg  cli.Argument
}

func NewArg(name string, flag cli.Argument) *Arg {
	return &Arg{name: name, arg: flag}
}

func (c *Arg) String() string        { return "<ctx>" }
func (c *Arg) Type() string          { return "Arg" }
func (c *Arg) Freeze()               {}
func (c *Arg) Truth() starlark.Bool  { return starlark.True }
func (c *Arg) Hash() (uint32, error) { return 0, errors.New("unhashable") }
