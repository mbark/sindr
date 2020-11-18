package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/iancoleman/strcase"
	"github.com/logrusorgru/aurora/v3"
	"github.com/urfave/cli/v2"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type task struct {
	Name string
	Fn   *lua.LFunction
	Env  string
	Args map[string]string
}

type env struct {
	Name    string
	Default bool
}

// Command ...
type Command struct {
	run func(ctx context.Context)
}

// Runtime ...
type Runtime struct {
	tasks        map[string]task
	environments map[string]env
	variables    map[string]string
	varOrder     []string // varOrder keeps track of which order the variables are registered

	defaultEnv *env
	modules    map[string]Module

	// Track all async commands being run
	wg sync.WaitGroup

	prevDir string

	cache  Cache
	logger *zap.SugaredLogger
}

// Module ...
type Module struct {
	exports map[string]lua.LGFunction
}

func getRuntime() *Runtime {
	home := os.Getenv("HOME")
	cacheHome := xdgPath("CACHE_HOME", path.Join(home, ".cache"))
	shmakeCache := path.Join(cacheHome, "shmake")
	cacheDir := path.Join(shmakeCache, "cache")

	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		panic(fmt.Errorf("creating cache directory: %w", err))
	}

	logPath := path.Join(shmakeCache, "shmake.log")
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{logPath}
	cfg.ErrorOutputPaths = []string{logPath}
	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Errorf("creating logger: %w", err))
	}

	r := &Runtime{
		tasks:        make(map[string]task),
		environments: make(map[string]env),
		variables:    make(map[string]string),
		modules:      make(map[string]Module),
		cache:        NewCache(cacheDir),
		logger:       logger.Sugar(),
	}

	mainModule := Module{
		exports: map[string]lua.LGFunction{
			"task": registerTask(func(t task) {
				logger.With(zap.String("name", t.Name)).Debug("registered task")

				r.tasks[t.Name] = t
			}),
			"env": registerEnv(func(e env) {
				logger.With(zap.String("name", e.Name)).Debug("registered env")

				r.environments[e.Name] = e
				if e.Default {
					r.defaultEnv = &e
				}
			}),
			"var": registerVar(func(name, value string) {
				logger.With(zap.String("name", name), zap.String("value", value)).Debug("registered var")

				r.variables[name] = value
				r.varOrder = append(r.varOrder, name)
			}),
		},
	}

	r.modules["shmake.main"] = mainModule
	return r
}

func (module Module) loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), module.exports)

	L.Push(mod)
	return 1
}

func registerTask(register func(t task)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		mapper := gluamapper.NewMapper(gluamapper.Option{NameFunc: func(n string) string { return n }})

		var t task
		if err := mapper.Map(lv.(*lua.LTable), &t); err != nil {
			panic(err)
		}

		register(t)
		return 0
	}
}

func registerEnv(register func(e env)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var e env
		if err := gluamapper.Map(lv.(*lua.LTable), &e); err != nil {
			panic(err)
		}

		register(e)
		L.Push(lua.LString(e.Name))
		return 1
	}
}

func registerVar(register func(name, value string)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var v struct {
			Name  string
			Value string
		}
		if err := gluamapper.Map(lv.(*lua.LTable), &v); err != nil {
			panic(err)
		}

		register(v.Name, v.Value)
		return 0
	}
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

			panic(err)
		}

		return filepath.Abs(dir)
	}

	return "", errors.New("path not found")
}

func main() {
	L := lua.NewState()
	defer L.Close()

	r := getRuntime()

	r.modules["shmake.files"] = getFileModule(r)
	r.modules["shmake.shell"] = getShellModule(r)
	r.modules["shmake.cache"] = getCacheModule(r)
	r.modules["shmake.git"] = getGitModule(r)
	r.modules["shmake.yarn"] = getYarnModule(r)
	r.modules["shmake.run"] = getRunModule(r)

	for name, module := range r.modules {
		L.PreloadModule(name, module.loader)
	}

	dir, err := findPathUpdwards("main.lua")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", aurora.Red(err))
		os.Exit(1)
	}
	err = os.Chdir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", aurora.Red(err))
		os.Exit(1)
	}

	if err := L.DoFile("main.lua"); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", aurora.Red(err))
		os.Exit(1)
	}

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
			Aliases:     []string{"v"},
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
		Flags: cliFlags,
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	if err := envApp.Run(os.Args); err != nil {
		panic(err)
	}

	if environment == "" && r.defaultEnv != nil {
		environment = r.defaultEnv.Name
	}

	r.logger.Debug("environment set", zap.String("environment", environment))
	if environment != "" {
		if _, ok := r.environments[environment]; !ok {
			panic(fmt.Sprintf("no environment %s", environment))
		}
	}

	r.cache.ForceOutOfDate = noCache

	r.logger.Debug("commands available", zap.Any("commands", r.tasks))

	var commands []*cli.Command
	for nameV, tV := range r.tasks {
		name := nameV
		t := tV

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
					if verbose {
						cfg := zap.NewDevelopmentConfig()
						cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
						cfg.OutputPaths = []string{"stdout"}
						cfg.ErrorOutputPaths = []string{"stdout"}
						logger, err := cfg.Build()
						if err != nil {
							panic(fmt.Errorf("building zap config: %w", err))
						}

						r.logger = logger.Sugar()
					}

					defer r.logger.Sync()
					r.logger.Debug("running command", zap.String("command", name))

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
					}

					L.SetContext(c.Context)
					err := L.CallByParam(lua.P{Fn: t.Fn, NRet: 1, Protect: true}, args)
					if err != nil {
						return err
					}

					return nil
				},
			})
	}

	app := &cli.App{
		Name:     "shmake",
		Usage:    "make shmake",
		Flags:    cliFlags,
		Commands: commands,
	}

	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
