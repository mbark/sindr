package main

import (
	"flag"
	"fmt"
	"strconv"
	"io/ioutil"
	"os"
	"path"

	"github.com/urfave/cli"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type task struct {
	Name string
	Fn   *lua.LFunction
	Env  string
}

type env struct {
	Name    string
	Default bool
}

// Runtime ...
type Runtime struct {
	tasks        map[string]task
	environments map[string]env
	defaultEnv   *env
	modules      map[string]Module
	commands     []func() int64
	cache        string
}

// Register a command to run, the command should return an int64 representing the
// unix timestamp for when it was last run.
// There are special values you can return:
// * -1 is always out of date, meaning that it will always be rerun (and
//   thus all commands after it as well)
// *  0 is always up to date, meaning it will never rerun unless a previous command
//    was out of date
func (runtime *Runtime) addCommand(fn func() int64) {
	runtime.commands = append(runtime.commands, fn)
}

func (runtime *Runtime) getLastTimestamp(hash string) int64 {
	file, err := os.Open(path.Join(runtime.cache, hash))
	if err != nil {
		return -1
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	if err != nil {
		return -1
	}

	timestamp, err := strconv.ParseInt(string(b), 10, 64)
	if err != nil {
		return -1
	}

	return timestamp
}

// Module ...
type Module struct {
	exports map[string]lua.LGFunction
}

func getRuntime() *Runtime {
	home := os.Getenv("HOME")
	cacheHome := xdgPath("CACHE_HOME", path.Join(home, ".cache"))

	var commands []func() int64
	r := &Runtime{
		tasks:        make(map[string]task),
		environments: make(map[string]env),
		modules:      make(map[string]Module),
		commands:     commands,
		cache:        path.Join(cacheHome, "shmake"),
	}

	mainModule := Module{
		exports: map[string]lua.LGFunction{
			"register_task": registerTask(func(t task) {
				fmt.Printf("registered task %s\n", t.Name)
				r.tasks[t.Name] = t
			}),
			"register_env": registerEnv(func(e env) {
				fmt.Printf("registered env %s\n", e.Name)
				r.environments[e.Name] = e
				if e.Default {
					r.defaultEnv = &e
				}
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

		var t task
		if err := gluamapper.Map(lv.(*lua.LTable), &t); err != nil {
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

func xdgPath(name, defaultPath string) string {
	dir := os.Getenv(fmt.Sprintf("XDG_%s", name))
	if dir != "" && path.IsAbs(dir) {
		return dir
	}

	return defaultPath
}

func main() {
	L := lua.NewState()
	defer L.Close()

	r := getRuntime()

	r.modules["shmake.files"] = getFileModule(r)
	r.modules["shmake.shell"] = getShellModule(r)

	for name, module := range r.modules {
		L.PreloadModule(name, module.loader)
	}

	if err := L.DoFile("main.lua"); err != nil {
		panic(err)
	}

	flags := flag.NewFlagSet("", flag.ContinueOnError)
	var environment = flags.String("env", "", "")
	flags.Parse(os.Args[1:])

	if *environment == "" && r.defaultEnv != nil {
		environment = &r.defaultEnv.Name
	}

	fmt.Printf("environment %s\n", *environment)

	app := cli.NewApp()
	app.Name = "shmake"
	app.Usage = "make shmake"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "env",
			Usage: "set which environment to show commands for",
		},
	}

	if *environment != "" {
		if _, ok := r.environments[*environment]; !ok {
			panic(fmt.Sprintf("no environment %s", *environment))
		}
	}

	for name, t := range r.tasks {
		if t.Env != "" && t.Env != *environment {
			continue
		}

		app.Commands = append(app.Commands,
			cli.Command{
				Name: name,
				Action: func(c *cli.Context) error {
					fmt.Printf("Running command %s\n", name)
					err := L.CallByParam(lua.P{
						Fn:      t.Fn,
						NRet:    1,
						Protect: true,
					})

					if err != nil {
						return err
					}

					for _, cmd := range r.commands {
						timestamp := cmd()
						fmt.Printf("timestamp: %s", timestamp)
					}

					return nil
				},
			})
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
