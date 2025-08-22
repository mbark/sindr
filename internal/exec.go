package internal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"go.starlark.net/starlark"

	"github.com/mbark/shmake/internal/logger"
)

func ShmakeExec(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var bin, command, prefix string
	var binArgs *starlark.List
	var noOutput bool
	if err := starlark.UnpackArgs("exec", args, kwargs,
		"bin", &bin,
		"command", &command,
		"args?", &binArgs,
		"no_output?", &noOutput,
		"prefix?", &prefix,
	); err != nil {
		return nil, err
	}
	if binArgs == nil {
		binArgs = new(starlark.List)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tmpdir, err := os.MkdirTemp("", "shmake")
	if err != nil {
		return nil, err
	}
	defer func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			logger.LogErr(fmt.Sprintf("failed to remove tmpdir %s", tmpdir), err)
		}
	}()

	file := filepath.Join(tmpdir, "exec")
	err = os.WriteFile(file, []byte(command), 0o644)
	if err != nil {
		return nil, fmt.Errorf("create file %s to exec: %w", file, err)
	}

	cmd := exec.CommandContext(ctx, "/usr/bin/env", bin, file) // #nosec G204
	if prefix != "" {
		logger.LogVerbose(prefixStyle.Render(prefix), commandStyle.Render("$ "+cmd.String()))
	} else {
		logger.LogVerbose(commandStyle.Render("$ " + cmd.String()))
	}

	res, err := StartShellCmd(cmd, prefix, noOutput)
	if err != nil {
		return nil, fmt.Errorf("start shell cmd failed: %w", err)
	}

	return res, nil
}
