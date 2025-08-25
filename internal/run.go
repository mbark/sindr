package internal

import (
	"errors"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"

	"github.com/mbark/sindr/internal/logger"
)

func SindrStart(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, errors.New("start() requires exactly 1 argument (a function)")
	}

	callable, ok := args.Index(0).(*starlark.Function)
	if !ok {
		return nil, errors.New("start() argument must be a callable function")
	}

	WaitGroup.Add(1)
	go func() {
		defer WaitGroup.Done()

		newThread := &starlark.Thread{Name: "async"}
		_, err := starlark.Call(newThread, callable, starlark.Tuple{}, nil)
		if err != nil {
			logger.LogErr("started function failed", err)
		}
	}()

	return starlark.None, nil
}

func SindrWait(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	WaitGroup.Wait()
	return starlark.None, nil
}

func SindrPool(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	pool := &Pool{wg: sync.WaitGroup{}}

	poolMethods := starlark.StringDict{
		"run":  starlark.NewBuiltin("pool.run", MakePoolRun(pool)),
		"wait": starlark.NewBuiltin("pool.wait", MakePoolWait(pool)),
	}

	return starlarkstruct.FromStringDict(starlark.String("pool"), poolMethods), nil
}

func MakePoolRun(pool *Pool) func(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return func(
		thread *starlark.Thread,
		fn *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		if args.Len() != 1 {
			return nil, errors.New("pool.run() requires exactly 1 argument (a function)")
		}

		callable, ok := args.Index(0).(*starlark.Function)
		if !ok {
			return nil, errors.New("pool.run() argument must be a callable function")
		}

		pool.wg.Add(1)
		go func() {
			defer pool.wg.Done()

			newThread := &starlark.Thread{Name: "pool"}
			_, err := starlark.Call(newThread, callable, starlark.Tuple{}, nil)
			if err != nil {
				logger.LogErr("pool function failed", err)
			}
		}()

		return starlark.None, nil
	}
}

func MakePoolWait(pool *Pool) func(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return func(
		thread *starlark.Thread,
		fn *starlark.Builtin,
		args starlark.Tuple,
		kwargs []starlark.Tuple,
	) (starlark.Value, error) {
		pool.wg.Wait()
		return starlark.None, nil
	}
}

type Pool struct {
	wg sync.WaitGroup
}
