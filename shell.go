package shmake

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/muesli/termenv"
	"go.starlark.net/starlark"
)

func shmakeShell(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var command, prefix string
	if err := starlark.UnpackArgs("shell", args, kwargs,
		"command", &command,
		"prefix?", &prefix,
	); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	slog.With(slog.String("command", command)).Debug("running shell command")

	cmd := exec.CommandContext(ctx, "bash", "-c", command) // #nosec G204
	out, err := startShellCmd(cmd, prefix)
	if err != nil {
		return nil, fmt.Errorf("start shell cmd failed: %w", err)
	}

	out = strings.TrimSpace(out)
	return starlark.String(out), nil
}

func startShellCmd(cmd *exec.Cmd, name string) (string, error) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("cmd start: %w", err)
	}

	prefix := func(s string) string {
		return termenv.String(s).Foreground(termenv.ANSIBrightBlue).Faint().String()
	}
	var out strings.Builder
	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			out.WriteString(m + "\n")
			if name != "" {
				fmt.Printf("%s | %s\n", prefix(name), m)
			} else {
				fmt.Println(m)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			if name != "" {
				fmt.Printf("%s | %s\n", prefix(name), m)
			} else {
				fmt.Println(m)
			}
		}
	}()

	err = cmd.Wait()
	return out.String(), err
}
