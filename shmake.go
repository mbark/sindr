package shmake

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

type ModuleFunction = func(runtime *Runtime, l *lua.LState) ([]lua.LValue, error)

type Module struct {
	exports map[string]ModuleFunction
}

func (m Module) withLogging() Module {
	for k, fn := range m.exports {
		m.exports[k] = func(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
			logger := slog.With(slog.String("fn", k))
			logger.Debug("running")
			res, err := fn(runtime, l)
			logger.With(slog.Any("err", err), slog.Any("res", res)).Debug("done")
			return res, err
		}
	}
	return m
}

func (m Module) loader(runtime *Runtime) lua.LGFunction {
	return func(l *lua.LState) int {
		exports := make(map[string]lua.LGFunction)
		for name, fn := range m.exports {
			f := fn
			exports[name] = func(l *lua.LState) int {
				rets, err := f(runtime, l)
				if err != nil {
					var et ErrBadType
					var ea ErrBadArg
					if ok := errors.As(err, &et); ok {
						l.TypeError(et.Index, et.Typ)
					}
					if ok := errors.As(err, &ea); ok {
						l.ArgError(ea.Index, ea.Message)
					}

					l.RaiseError("%s", err.Error())
				}

				for _, ret := range rets {
					l.Push(ret)
				}

				return len(rets)
			}
		}

		mod := l.SetFuncs(l.NewTable(), exports)

		l.Push(mod)
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

func getMainModule() Module {
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

func Run() {
	cli.RootCommandHelpTemplate = helpTemplate

	l := lua.NewState()
	defer l.Close()

	// Ensure we have a context before the app starts
	ctx := context.Background()
	l.SetContext(ctx)

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

	RegisterLuaTypes(r, l, ShmakeType{Runtime: r}, CommandType{Runtime: r}, PoolType{})

	r.modules["shmake.main"] = getMainModule()
	r.modules["shmake.files"] = getFileModule(r)

	for name, module := range r.modules {
		l.PreloadModule(name, module.withLogging().loader(r))
	}

	dir, err := findPathUpdwards("main.lua")
	checkErr(err)

	err = os.Chdir(dir)
	checkErr(err)

	l.SetGlobal("print", l.NewFunction(func(l *lua.LState) int {
		var buf bytes.Buffer
		top := l.GetTop()
		for i := 1; i <= top; i++ {
			buf.WriteString(l.Get(i).String())
			if i != top {
				buf.WriteString("\t")
			}
		}
		buf.WriteString("\n")

		_, _ = os.Stdout.Write(buf.Bytes())
		return 0
	}))

	l.SetGlobal("current_dir", lua.LString(dir))
	err = l.DoFile("main.lua")
	checkErr(err)
}
