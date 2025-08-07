package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/muesli/termenv"
	lua "github.com/yuin/gopher-lua"
)

type runOptions struct {
	Prefix string
}

func shell(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(1)
	c, err := MapString(1, lv)
	if err != nil {
		return nil, err
	}

	var options runOptions
	if L.GetTop() > 1 {
		err := MapTable(2, L.Get(2), &options)
		if err != nil {
			return nil, err
		}
	}

	slog.With(slog.String("command", c)).Debug("running shell command")

	cmd := exec.CommandContext(L.Context(), "bash", "-c", c)
	out, err := startShellCmd(cmd, options.Prefix)
	if err != nil {
		return nil, fmt.Errorf("start shell cmd failed: %w", err)
	}

	out = strings.TrimSpace(out)
	return []lua.LValue{lua.LString(out)}, nil
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
			out.WriteString(m)
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
