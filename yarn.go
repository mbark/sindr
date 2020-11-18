package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func getYarnModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"run":     yarnRun(runtime),
			"install": yarnInstall(runtime),
		},
	}
}

func yarnRun(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		str, ok := lv.(lua.LString)
		if !ok {
			panic("argument must be a string")
		}

		runtime.addCommand(Command{
			run: func(ctx context.Context) {
				args := strings.Split(string(str), " ")
				args = append([]string{"run"}, args...)
				cmd := exec.CommandContext(ctx, "yarn", args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					fmt.Println("yarn", "run", string(str))
					panic(err)
				}
			},
		})

		return 0
	}
}

func yarnInstall(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		runtime.addCommand(Command{
			run: func(ctx context.Context) {
				cmd := exec.CommandContext(ctx, "yarn", "install")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err := cmd.Run()
				if err != nil {
					panic(err)
				}
			},
		})

		return 0
	}
}
