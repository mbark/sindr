package main

import (
	"fmt"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

func getShellModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"run": run(runtime.addCommand),
		},
	}
}

func run(addCommand func(cmd func() int64)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		if lv.Type() != lua.LTString {
			panic("string required.")
		}

		if str, ok := lv.(lua.LString); ok {
			addCommand(func() int64 {
				fmt.Printf("running command %s", string(str))
				cmd := exec.Command("bash", "-c", fmt.Sprintf("%s", string(str)))
				err := cmd.Run()
				if err != nil {
					panic(err)
				}

				return -1
			})
		}

		return 0
	}
}
