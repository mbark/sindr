package internal

import (
	"errors"
	"fmt"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
)

func SindrStringFlag(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrFlag[starlark.String](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue starlark.String) (cli.Flag, error) {
			return &cli.StringFlag{
				Name:  name,
				Usage: usage,
				Value: string(defaultValue),
			}, nil
		},
	)
}

func SindrBoolFlag(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrFlag[starlark.Bool](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue starlark.Bool) (cli.Flag, error) {
			return &cli.BoolFlag{
				Name:  name,
				Usage: usage,
				Value: bool(defaultValue),
			}, nil
		},
	)
}

func SindrIntFlag(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrFlag[starlark.Int](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue starlark.Int) (cli.Flag, error) {
			i, err := castInt(defaultValue)
			if err != nil {
				return nil, err
			}

			return &cli.IntFlag{
				Name:  name,
				Usage: usage,
				Value: i,
			}, nil
		},
	)
}

func SindrStringSliceFlag(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrFlag[*starlark.List](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue *starlark.List) (cli.Flag, error) {
			value, err := fromList(
				defaultValue,
				func(value starlark.Value) (string, error) { return castString(value) },
			)
			if err != nil {
				return nil, err
			}

			return &cli.StringSliceFlag{
				Name:  name,
				Usage: usage,
				Value: value,
			}, nil
		},
	)
}

func SindrIntSliceFlag(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return sindrFlag[*starlark.List](
		thread,
		fn,
		args,
		kwargs,
		func(name, usage string, defaultValue *starlark.List) (cli.Flag, error) {
			value, err := fromList(
				defaultValue,
				func(value starlark.Value) (int, error) { return castInt(value) },
			)
			if err != nil {
				return nil, err
			}

			return &cli.IntSliceFlag{
				Name:  name,
				Usage: usage,
				Value: value,
			}, nil
		},
	)
}

func sindrFlag[T starlark.Value](
	_ *starlark.Thread,
	_ *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
	newFlag func(name, usage string, defaultValue T) (cli.Flag, error),
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

	f, err := newFlag(name, usage, defaultValue)
	if err != nil {
		return nil, err
	}
	return NewFlag(f), nil
}

func processFlags(flagsList *starlark.List, cmd *Command) error {
	if flagsList == nil {
		return nil
	}

	for item := range flagsList.Elements() {
		flag, ok := item.(*Flag)
		if !ok {
			return fmt.Errorf("expected flag, got: %T", item)
		}

		cmd.Command.Flags = append(cmd.Command.Flags, flag.flag)
	}

	return nil
}

var _ starlark.Value = (*Flag)(nil)

type Flag struct {
	flag cli.Flag
}

func NewFlag(flag cli.Flag) *Flag {
	return &Flag{flag: flag}
}

func (c *Flag) String() string        { return "<ctx>" }
func (c *Flag) Type() string          { return "Flag" }
func (c *Flag) Freeze()               {}
func (c *Flag) Truth() starlark.Bool  { return starlark.True }
func (c *Flag) Hash() (uint32, error) { return 0, errors.New("unhashable") }

var _ starlark.Value = (*Arg)(nil)
