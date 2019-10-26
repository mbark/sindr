package main

import (
	"fmt"
	"os/exec"

	lua "github.com/yuin/gopher-lua"
)

func getShellModule() Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"run": run,
		},
	}
}

func run(L *lua.LState) int {
	lv := L.Get(-1)
	if str, ok := lv.(lua.LString); ok {
		fmt.Printf("running command %s", string(str))
		cmd := exec.Command("bash", "-c", fmt.Sprintf("%s", string(str)))
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	}
	if lv.Type() != lua.LTString {
		panic("string required.")
	}

	return 0
}
