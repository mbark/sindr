package shmake

import (
	"errors"
	"log/slog"
	"sync"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func shmakeStart(
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

	wg.Add(1)
	go func() {
		defer wg.Done()

		newThread := &starlark.Thread{Name: "async"}
		_, err := starlark.Call(newThread, callable, starlark.Tuple{}, nil)
		if err != nil {
			slog.Error("async function failed", "error", err)
		}
	}()

	return starlark.None, nil
}

func shmakeWait(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	wg.Wait()
	return starlark.None, nil
}

func shmakePool(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	pool := &Pool{wg: sync.WaitGroup{}}

	poolMethods := starlark.StringDict{
		"run":  starlark.NewBuiltin("pool.run", makePoolRun(pool)),
		"wait": starlark.NewBuiltin("pool.wait", makePoolWait(pool)),
	}

	return starlarkstruct.FromStringDict(starlark.String("pool"), poolMethods), nil
}

func makePoolRun(pool *Pool) func(
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
				slog.Error("pool function failed", "error", err)
			}
		}()

		return starlark.None, nil
	}
}

func makePoolWait(pool *Pool) func(
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
