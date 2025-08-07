package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/log"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v3"
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
			logger := slog.With(slog.String("fn", k))
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
	logFile io.WriteCloser
}

func NewRuntime(logFile io.WriteCloser) (*Runtime, error) {
	cacheDir := cacheHome()

	logger := slog.New(slogmulti.Fanout(
		slog.NewJSONHandler(logFile, &slog.HandlerOptions{Level: slog.LevelDebug}),
		log.NewWithOptions(os.Stderr, log.Options{}),
	))
	slog.SetDefault(logger)

	r := &Runtime{
		modules: make(map[string]Module),
		cache:   NewCache(cacheDir),
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
	cli.RootCommandHelpTemplate = helpTemplate

	L := lua.NewState()
	defer L.Close()

	// Ensure we have a context before the app starts
	ctx := context.Background()
	L.SetContext(ctx)

	checkErr := func(err error) {
		if err == nil {
			return
		}

		var lerr *lua.ApiError
		if errors.As(err, &lerr) {
			slog.With(slog.String("stack_trace", strings.ReplaceAll(lerr.StackTrace, "\t", "  "))).Error(lerr.Object.String())
		} else if err != nil {
			slog.Error(lerr.Object.String())
		}

		os.Exit(1)
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

	L.SetGlobal("print", L.NewFunction(func(L *lua.LState) int {
		var buf bytes.Buffer
		top := L.GetTop()
		for i := 1; i <= top; i++ {
			buf.WriteString(L.Get(i).String())
			if i != top {
				buf.WriteString("\t")
			}
		}
		buf.WriteString("\n")

		_, _ = os.Stdout.Write(buf.Bytes())
		return 0
	}))

	L.SetGlobal("current_dir", lua.LString(dir))
	err = L.DoFile("main.lua")
	checkErr(err)
}
