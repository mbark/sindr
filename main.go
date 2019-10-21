package main

import (
	"fmt"
	"os"

	"github.com/mbark/shmake/shell"
	"github.com/urfave/cli"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type task struct {
	Name string
	Fn   *lua.LFunction
}

type runtime struct {
	tasks   map[string]task
	exports map[string]lua.LGFunction
}

func getRuntime() *runtime {
	r := &runtime{
		tasks:   make(map[string]task),
		exports: make(map[string]lua.LGFunction),
	}

	r.exports["register_task"] = registerTask(func(t task) {
		fmt.Printf("registered task %s\n", t.Name)
		r.tasks[t.Name] = t
	})

	return r
}

func (runtime runtime) loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), runtime.exports)

	L.Push(mod)
	return 1
}

func registerTask(register func(t task)) func(L *lua.LState) int {
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

func main() {
	L := lua.NewState()
	defer L.Close()

	r := getRuntime()

	L.PreloadModule("shmake.main", r.loader)
	L.PreloadModule("shmake.shell", shell.Loader)
	if err := L.DoFile("main.lua"); err != nil {
		panic(err)
	}

	app := cli.NewApp()

	app.Name = "shmake"
	app.Usage = "make shmake"
	for name, t := range r.tasks {
		app.Commands = []cli.Command{
			{
				Name: name,
				Action: func(c *cli.Context) error {
					return L.CallByParam(lua.P{
						Fn:      t.Fn,
						NRet:    1,
						Protect: true,
					})
				},
			},
		}
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
