package main

import (
	"errors"
	"flag"
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

// NoReturnVal is a shorthand for an empty array of LValues
var NoReturnVal = []lua.LValue{}

type Cmd struct {
	Name string
	Fn   *lua.LFunction
	Env  string
	Args map[string]string
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

type Runtime struct {
	commands     map[string]Cmd
	environments map[string]Env
	variables    map[string]Var
	varOrder     []string // varOrder keeps track of which order the variables are registered

	defaultEnv *Env
	modules    map[string]Module

	// Track all async commands being run
	wg sync.WaitGroup

	prevDir string

	cache  Cache
	logger *zap.SugaredLogger
}

func cacheHome() string {
	home := os.Getenv("HOME")
	cacheDir := xdgPath("CACHE_HOME", path.Join(home, ".cache"))
	return path.Join(cacheDir, "shmake")
}

func logPath(cacheDir string) string {
	return path.Join(cacheDir, "shmake.log")
}

func NewRuntime() (*Runtime, error) {
	cacheDir := cacheHome()
	logPath := logPath(cacheDir)

	err := os.MkdirAll(cacheDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{logPath}
	cfg.ErrorOutputPaths = []string{logPath}

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("creating logger: %w", err)
	}

	r := &Runtime{
		commands:     make(map[string]Cmd),
		environments: make(map[string]Env),
		variables:    make(map[string]Var),
		modules:      make(map[string]Module),
		cache:        NewCache(cacheDir),
		logger:       logger.Sugar(),
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

func (module Module) loader(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		exports := make(map[string]lua.LGFunction)
		for name, fn := range module.exports {
			f := fn
			exports[name] = func(L *lua.LState) int {
				rets, err := f(runtime, L)
				if err != nil {
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

func registerCmd(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	mapper := gluamapper.NewMapper(gluamapper.Option{NameFunc: func(n string) string { return n }})

	var t Cmd
	if err := mapper.Map(lv.(*lua.LTable), &t); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	runtime.commands[t.Name] = t
	return NoReturnVal, nil
}

func registerEnv(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var e Env
	if err := gluamapper.Map(lv.(*lua.LTable), &e); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
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
	if err := gluamapper.Map(lv.(*lua.LTable), &v); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
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

	checkErr := func(err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", aurora.Red(err))
			os.Exit(1)
		}
	}

	r, err := NewRuntime()
	checkErr(err)

	r.modules["shmake.files"] = getFileModule(r)
	r.modules["shmake.shell"] = getShellModule(r)
	r.modules["shmake.cache"] = getCacheModule(r)
	r.modules["shmake.git"] = getGitModule(r)
	r.modules["shmake.yarn"] = getYarnModule(r)
	r.modules["shmake.run"] = getRunModule(r)

	for name, module := range r.modules {
		L.PreloadModule(name, module.loader(r))
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
		HideHelp: true,
		HideHelpCommand: true,
		OnUsageError: func(context *cli.Context, err error, isSubcommand bool) error {
			// Ignore errors when trying to get help, let the other app handle that
			if errors.Is(err, flag.ErrHelp) {
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
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.OutputPaths = []string{"stdout", logPath(cacheHome())}
		cfg.ErrorOutputPaths = []string{"stderr", logPath(cacheHome())}

		logger, err := cfg.Build()
		checkErr(err)

		r.logger = logger.Sugar()
	}

	if environment == "" && r.defaultEnv != nil {
		environment = r.defaultEnv.Name
	}

	r.logger.With(zap.String("env", environment)).Debug("environment set")
	if environment != "" {
		if _, ok := r.environments[environment]; !ok {
			checkErr(errors.New(fmt.Sprintf("no environment with name %s", environment)))
		}
	}

	var commands []*cli.Command
	for nameV, tV := range r.commands {
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
					defer r.logger.Sync()
					log := r.logger.With(zap.String("cmd", name))

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

						log = log.With(zap.String("arg."+k, v))
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

	app := &cli.App{
		Name:     "shmake",
		Usage:    "make shmake",
		Flags:    cliFlags,
		Commands: commands,
	}

	app.Run(os.Args)
	checkErr(err)
}
