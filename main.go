package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/logrusorgru/aurora/v3"
	"github.com/mbark/devslog"
	slogmulti "github.com/samber/slog-multi"
	"github.com/urfave/cli/v2"
	lua "github.com/yuin/gopher-lua"
)

const version = "0.0.1"

// NoReturnVal is a shorthand for an empty array of LValues
var NoReturnVal = []lua.LValue{}

type Cmd struct {
	Name       string
	SubCmdPath []string
	Fn         *lua.LFunction
	Env        string
	Args       map[string]string
}

type Env struct {
	Name    string
	Default bool
}

type Var = string

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
	commands     map[string]Cmd
	subCommands  map[string]Cmd
	environments map[string]Env
	variables    map[string]Var
	varOrder     []string // varOrder keeps track of which order the variables are registered

	defaultEnv *Env
	modules    map[string]Module

	// Track all async commands being run
	wg sync.WaitGroup

	prevDir string

	cache  Cache
	logger *slog.Logger
}

func cacheHome() string {
	home := os.Getenv("HOME")
	cacheDir := xdgPath("CACHE_HOME", path.Join(home, ".cache"))
	return path.Join(cacheDir, "shmake")
}

func logPath(cacheDir string) string {
	return path.Join(cacheDir, "shmake.log")
}

func getLogFile() (io.WriteCloser, error) {
	cacheDir := cacheHome()
	logFile := logPath(cacheDir)

	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	_, err = os.Stat(logFile)
	var buf io.WriteCloser
	if errors.Is(err, os.ErrNotExist) {
		buf, err = os.Create(logFile)
	} else if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	} else {
		buf, err = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0600)
	}
	return buf, err
}

func NewRuntime(logFile io.WriteCloser) (*Runtime, error) {
	cacheDir := cacheHome()
	r := &Runtime{
		commands:     make(map[string]Cmd),
		environments: make(map[string]Env),
		variables:    make(map[string]Var),
		modules:      make(map[string]Module),
		cache:        NewCache(cacheDir),
		logger:       slog.New(slog.NewJSONHandler(logFile, nil)),
	}

	mainModule := Module{
		exports: map[string]ModuleFunction{
			"cmd": registerCmd,
			"env": registerEnv,
			"var": registerVar,
		},
	}

	r.modules["shmake.main"] = mainModule
	return r, nil
}

type cmdOpts struct {
	Env  string
	Args map[string]string
}

// registerCmd fn = shmake.cmd('name', function(args) end, { opts... })
func registerCmd(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv1 := L.Get(-1) // table if three arguments have been passed
	lv2 := L.Get(-2) // table if three arguments have been passed
	lv3 := L.Get(-3) // table if three arguments have been passed

	var err error
	var t cmdOpts
	var name string
	var fn *lua.LFunction

	// Three arguments passed, should be string, fn, table
	if lv1.Type() != lua.LTNil && lv2.Type() != lua.LTNil && lv3.Type() != lua.LTNil {
		name, err = MapString(1, lv3)
		if err != nil {
			return nil, err
		}

		fn, err = MapFunction(2, lv2)
		if err != nil {
			return nil, err
		}

		err = MapTable(3, lv1, &t)
		if err != nil {
			return nil, ErrBadArg{Index: 3, Message: fmt.Errorf("invalid config: %w", err).Error()}
		}
	} else if lv1.Type() != lua.LTNil && lv2.Type() != lua.LTNil && lv3.Type() == lua.LTNil {
		name, err = MapString(1, lv2)
		if err != nil {
			return nil, err
		}

		fn, err = MapFunction(2, lv1)
		if err != nil {
			return nil, err
		}
	} else {
		L.RaiseError("wrong number of arguments passed to cmd")
	}
	if strings.Contains(name, " ") {
		L.RaiseError("name can't contain spaces")
	}

	runtime.commands[name] = Cmd{
		Name: name,
		Fn:   fn,
		Env:  t.Env,
		Args: t.Args,
	}

	return []lua.LValue{fn}, nil
}

// registerSubCmd fn = shmake.sub_cmd('name', function(args) end, { opts... })
func registerSubCmd(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv1 := L.Get(-1) // table if three arguments have been passed
	lv2 := L.Get(-2) // table if three arguments have been passed
	lv3 := L.Get(-3) // table if three arguments have been passed

	var err error
	var t cmdOpts
	var path []string
	var fn *lua.LFunction

	// Three arguments passed, should be table, fn, table
	if lv1.Type() != lua.LTNil && lv2.Type() != lua.LTNil && lv3.Type() != lua.LTNil {
		path, err = MapArray[string](1, lv3)
		if err != nil {
			return nil, err
		}

		fn, err = MapFunction(2, lv2)
		if err != nil {
			return nil, err
		}

		err = MapTable(3, lv1, &t)
		if err != nil {
			return nil, ErrBadArg{Index: 3, Message: fmt.Errorf("invalid config: %w", err).Error()}
		}
	} else if lv1.Type() != lua.LTNil && lv2.Type() != lua.LTNil && lv3.Type() == lua.LTNil {
		path, err = MapArray[string](1, lv2)
		if err != nil {
			return nil, err
		}

		fn, err = MapFunction(2, lv1)
		if err != nil {
			return nil, err
		}
	} else {
		L.RaiseError("wrong number of arguments passed to cmd")
	}
	for _, name := range path {
		if strings.Contains(name, " ") {
			L.RaiseError("command name can't contain spaces")
		}
	}

	runtime.commands[strings.Join(path, " ")] = Cmd{
		SubCmdPath: path,
		Fn:         fn,
		Env:        t.Env,
		Args:       t.Args,
	}

	return []lua.LValue{fn}, nil
}

func registerEnv(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var e Env
	err := MapTable(1, lv, &e)
	if err != nil {
		return nil, err
	}

	runtime.environments[e.Name] = e
	if e.Default {
		runtime.defaultEnv = &e
	}

	return []lua.LValue{lua.LString(e.Name)}, nil
}

func registerVar(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var v struct {
		Name  string
		Value string
	}
	err := MapTable(1, lv, &v)
	if err != nil {
		return nil, err
	}

	runtime.variables[v.Name] = v.Value
	runtime.varOrder = append(runtime.varOrder, v.Name)
	return NoReturnVal, nil
}

func xdgPath(name, defaultPath string) string {
	dir := os.Getenv(fmt.Sprintf("XDG_%s", name))
	if dir != "" && path.IsAbs(dir) {
		return dir
	}

	return defaultPath
}

func findPathUpdwards(search string) (string, error) {
	dir := "."

	for {
		// If we hit the root, we're done
		if rel, _ := filepath.Rel("/", search); rel == "." {
			break
		}

		_, err := os.Stat(filepath.Join(dir, search))
		if err != nil {
			if os.IsNotExist(err) {
				dir += "/.."
				continue
			}

			return "", err
		}

		return filepath.Abs(dir)
	}

	return "", errors.New("path not found")
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

	r.modules["shmake.files"] = getFileModule(r)
	r.modules["shmake.shell"] = getShellModule()
	r.modules["shmake.cache"] = getCacheModule(r)
	r.modules["shmake.git"] = getGitModule(r)
	r.modules["shmake.yarn"] = getYarnModule(r)
	r.modules["shmake.run"] = getRunModule()

	for name, module := range r.modules {
		L.PreloadModule(name, module.withLogging().loader(r))
	}

	dir, err := findPathUpdwards("main.lua")
	checkErr(err)

	err = os.Chdir(dir)
	checkErr(err)

	err = L.DoFile("main.lua")
	checkErr(err)

	var environment string
	var verbose, noCache bool

	cliFlags := []cli.Flag{
		&cli.StringFlag{
			Name:        "env",
			Usage:       "set which environment to show commands for",
			Destination: &environment,
		},
		&cli.BoolFlag{
			Name:        "verbose",
			Usage:       "print logs to stdout",
			Destination: &verbose,
		},
		&cli.BoolFlag{
			Name:        "no-cache",
			Usage:       "ignore stored values in the cache",
			Destination: &noCache,
		},
	}

	for name, value := range r.variables {
		cliFlags = append(cliFlags, &cli.StringFlag{
			Name:    strcase.ToKebab(name),
			EnvVars: []string{strcase.ToScreamingSnake(name)},
			Value:   value,
		})
	}

	envApp := &cli.App{
		Flags:           cliFlags,
		HideVersion:     true,
		HideHelp:        true,
		HideHelpCommand: true,
		OnUsageError: func(context *cli.Context, err error, isSubcommand bool) error {
			// Ignore errors when trying to get help, let the other app handle that
			if errors.Is(err, flag.ErrHelp) {
				return nil
			}

			// ignore unrecognized flags here
			if strings.Contains(err.Error(), "flag provided but not defined") {
				return nil
			}

			return err
		},
		Action: func(c *cli.Context) error { return nil },
	}
	err = envApp.Run(os.Args)
	checkErr(err)

	r.cache.ForceOutOfDate = noCache
	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		opts := &slog.HandlerOptions{Level: slog.LevelDebug}
		r.logger = slog.New(slogmulti.Fanout(
			slog.NewJSONHandler(logFile, opts),
			devslog.NewHandler(os.Stdout, &devslog.Options{HandlerOptions: opts}),
		))

	}

	if environment == "" && r.defaultEnv != nil {
		environment = r.defaultEnv.Name
	}

	r.logger.With(slog.String("env", environment)).Debug("environment set")
	if environment != "" {
		if _, ok := r.environments[environment]; !ok {
			checkErr(errors.New(fmt.Sprintf("no environment with name %s", environment)))
		}
	}

	var commands []*cli.Command
	for name, t := range r.commands {
		// TODO: allow handling of sub commands here by moving out this logic
		if t.Env != "" && t.Env != environment {
			continue
		}

		var cmdFlags []cli.Flag
		for name, value := range t.Args {
			cmdFlags = append(cmdFlags, &cli.StringFlag{
				Name:    strcase.ToKebab(name),
				EnvVars: []string{strcase.ToScreamingSnake(name)},
				Value:   value,
			})
		}

		commands = append(commands,
			&cli.Command{
				Name:  name,
				Flags: cmdFlags,
				Action: func(c *cli.Context) error {
					log := r.logger.With(slog.String("cmd", name))

					for _, name := range r.varOrder {
						value := r.variables[name]

						v := c.String(strcase.ToKebab(name))
						if v != "" {
							value = v
						}

						value = withVariables(r, value)
						r.variables[name] = value

						L.SetGlobal(name, lua.LString(value))
					}

					args := &lua.LTable{Metatable: lua.LNil}
					for k, v := range t.Args {
						v = withVariables(r, v)
						args.RawSetString(k, lua.LString(v))

						log = log.With(slog.String("arg."+k, v))
					}

					for k, v := range t.Args {
						k = strcase.ToCamel(k)
						r.variables[k] = withVariables(r, v)
						r.varOrder = append(r.varOrder, k)
					}

					log.Debug("running cmd")

					L.SetContext(c.Context)
					err := L.CallByParam(lua.P{Fn: t.Fn, NRet: 1, Protect: true}, args)
					if err != nil {
						return err
					}

					return nil
				},
			})
	}

	r.logger.Debug("starting app")
	app := &cli.App{
		Name:     "shmake",
		Usage:    "make shmake",
		Version:  version,
		Flags:    cliFlags,
		Commands: commands,
	}
	err = app.Run(os.Args)
	checkErr(err)
}
