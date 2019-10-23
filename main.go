package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mbark/shmake/files"
	"github.com/mbark/shmake/shell"
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

type runtime struct {
	tasks        map[string]task
	environments map[string]env
	defaultEnv   *env
	exports      map[string]lua.LGFunction
}

func getRuntime() *runtime {
	r := &runtime{
		tasks:        make(map[string]task),
		environments: make(map[string]env),
		exports:      make(map[string]lua.LGFunction),
	}

	r.exports["register_task"] = registerTask(func(t task) {
		fmt.Printf("registered task %s\n", t.Name)
		r.tasks[t.Name] = t
	})

	r.exports["register_env"] = registerEnv(func(e env) {
		fmt.Printf("registered env %s\n", e.Name)
		r.environments[e.Name] = e
		if e.Default {
			r.defaultEnv = &e
		}
	})

	return r
}

func (runtime runtime) loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), runtime.exports)

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

func main() {
	L := lua.NewState()
	defer L.Close()

	r := getRuntime()

	L.PreloadModule("shmake.main", r.loader)
	L.PreloadModule("shmake.shell", shell.Loader)
	L.PreloadModule("shmake.files", files.Loader)
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
		fmt.Printf("adding task %s with environment %s\n", t.Name, t.Env)
		if t.Env != "" && t.Env != *environment {
			fmt.Printf("skip adding %s\n", t.Name)
			continue
		}

		app.Commands = append(app.Commands,
			cli.Command{
				Name: name,
				Action: func(c *cli.Context) error {
					fmt.Printf("Running command %s\n", name)
					return L.CallByParam(lua.P{
						Fn:      t.Fn,
						NRet:    1,
						Protect: true,
					})
				},
			})
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
