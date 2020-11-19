package main

import (
	"os"
	"os/exec"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func getYarnModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"run":     yarnRun,
			"install": yarnInstall,
		},
	}
}

func yarnRun(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	str, ok := lv.(lua.LString)
	if !ok {
		L.TypeError(1, lua.LTString)
	}

	args := strings.Split(string(str), " ")
	args = append([]string{"run"}, args...)
	cmd := exec.CommandContext(L.Context(), "yarn", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return NoReturnVal, nil
}

func yarnInstall(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	cmd := exec.CommandContext(L.Context(), "yarn", "install")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	return NoReturnVal, nil
}
