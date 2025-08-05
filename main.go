package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/logrusorgru/aurora/v3"
	lua "github.com/yuin/gopher-lua"
)

const version = "0.0.1"

// NoReturnVal is shorthand for an empty array of LValues
var NoReturnVal = []lua.LValue{}

type ModuleFunction = func(runtime *Runtime, L *lua.LState) ([]lua.LValue, error)

type Module struct {
	exports map[string]ModuleFunction
}

func (m Module) withLogging() Module {
	for k, fn := range m.exports {
		m.exports[k] = func(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
			logger := runtime.logger.With(slog.String("fn", k))
			logger.Debug("running")
			res, err := fn(runtime, L)
			logger.With(slog.Any("err", err), slog.Any("res", res)).Debug("done")
			return res, err
		}
	}
	return m
}

func (m Module) loader(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		exports := make(map[string]lua.LGFunction)
		for name, fn := range m.exports {
			f := fn
			exports[name] = func(L *lua.LState) int {
				rets, err := f(runtime, L)
				if err != nil {
					var et ErrBadType
					var ea ErrBadArg
					if ok := errors.As(err, &et); ok {
						L.TypeError(et.Index, et.Typ)
					}
					if ok := errors.As(err, &ea); ok {
						L.ArgError(ea.Index, ea.Message)
					}

					L.RaiseError(err.Error())
				}

				for _, ret := range rets {
					L.Push(ret)
				}

				return len(rets)
			}
		}

		mod := L.SetFuncs(L.NewTable(), exports)

		L.Push(mod)
		return 1
	}
}

type Runtime struct {
	modules map[string]Module

	// Track all async commands being run
	wg sync.WaitGroup

	prevDir string

	cache   Cache
	logger  *slog.Logger
	logFile io.WriteCloser
}

func NewRuntime(logFile io.WriteCloser) (*Runtime, error) {
	cacheDir := cacheHome()
	r := &Runtime{
		modules: make(map[string]Module),
		cache:   NewCache(cacheDir),
		logger:  slog.New(slog.NewJSONHandler(logFile, nil)),
		logFile: logFile,
	}

	return r, nil
}

func getMainModule(r *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"command": newCommand,
			"string":  templateString,

			"async": async,
			"wait":  wait,
			"watch": watch,
			"pool":  pool,

			"shell": shell,

			"diff":         diff,
			"store":        store,
			"with_version": withVersion,
		},
	}
}

func main() {
	L := lua.NewState()
	defer L.Close()

	// Ensure we have a context before the app starts
	ctx := context.Background()
	L.SetContext(ctx)

	checkErr := func(err error) {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "%s\n", aurora.Red(err))
			os.Exit(1)
		}
	}

	logFile, err := getLogFile()
	checkErr(err)
	defer func() { _ = logFile.Close() }()

	r, err := NewRuntime(logFile)
	checkErr(err)

	RegisterLuaTypes(r, L, ShmakeType{Runtime: r}, CommandType{Runtime: r}, PoolType{})

	r.modules["shmake.main"] = getMainModule(r)
	r.modules["shmake.files"] = getFileModule(r)

	for name, module := range r.modules {
		L.PreloadModule(name, module.withLogging().loader(r))
	}

	dir, err := findPathUpdwards("main.lua")
	checkErr(err)

	err = os.Chdir(dir)
	checkErr(err)

	L.SetGlobal("current_dir", lua.LString(dir))
	err = L.DoFile("main.lua")
	checkErr(err)
}
