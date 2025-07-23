package main

import (
	"log/slog"
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

	c, err := MapString(1, lv)
	if err != nil {
		return nil, err
	}

	args := strings.Split(c, " ")
	args = append([]string{"run"}, args...)

	runtime.logger.With(slog.Any("args", args)).Debug("yarn run")

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
