package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"text/template"

	"github.com/logrusorgru/aurora/v3"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func getShellModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"run":   run(runtime),
			"start": start(runtime),
		},
	}
}

func withVariables(runtime *Runtime, input string) string {
	t := template.Must(template.New("run").Parse(input))
	var buf bytes.Buffer
	err := t.Execute(&buf, runtime.variables)
	if err != nil {
		panic(fmt.Errorf("execute template: %w", err))
	}

	return buf.String()
}

func run(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		str, ok := lv.(lua.LString)
		if !ok {
			L.TypeError(1, lua.LTString)
		}

		command := withVariables(runtime, string(str))
		runtime.logger.Debug("running command", zap.String("command", command))
		cmd := exec.CommandContext(L.Context(), "bash", "-c", command)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			panic(err)
		}

		return 0
	}
}

type startOptions = map[string]struct {
	Cmd   string
	Watch string
}

func start(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		tbl, ok := lv.(*lua.LTable)
		if !ok {
			L.TypeError(1, lua.LTTable)
		}

		var startCommands startOptions
		if err := gluamapper.Map(tbl, &startCommands); err != nil {
			runtime.logger.Fatal("failed to map commands", zap.Error(err))
		}

		for k, c := range startCommands {
			c.Cmd = withVariables(runtime, c.Cmd)
			startCommands[k] = c
		}

		var colorIdx uint8 = 0

		wg := sync.WaitGroup{}
		for k, c := range startCommands {
			runtime.logger.Debug("starting command",
				zap.String("name", k),
				zap.String("command", c.Cmd),
				zap.String("watch", c.Watch))

			wg.Add(1)
			colorIdx += 1
			go func(name, command, watch string, colorIndex uint8) {
				defer wg.Done()

				if watch != "" {
					onChange := make(chan bool)
					close := startWatching(runtime, watch, onChange)
					defer close()

					for {
						Lt, cancel := L.NewThread()
						cmd := exec.CommandContext(Lt.Context(), "bash", "-c", fmt.Sprintf("%s", command))
						err := startCommand(cmd, name, colorIndex)
						if err != nil {
							runtime.logger.Fatal("start command", zap.Error(err))
						}

						runtime.logger.Info("command started", zap.String("command", command))

						_ = <-onChange

						runtime.logger.Info("restarting", zap.String("command", command))
						cancel()
					}
				} else {
					cmd := exec.CommandContext(L.Context(), "bash", "-c", fmt.Sprintf("%s", command))
					err := startCommand(cmd, name, colorIndex)
					if err != nil {
						runtime.logger.Fatal("start command", zap.Error(err))
					}
					runtime.logger.Info("command started", zap.String("command", command))

					if err := cmd.Wait(); err != nil {
						runtime.logger.Fatal("cmd wait", zap.Error(err))
					}
				}
			}(k, c.Cmd, c.Watch, colorIdx)
		}

		wg.Wait()

		return 0
	}
}

func startCommand(cmd *exec.Cmd, name string, colorIndex uint8) error {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd start: %w", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Printf("%s | %s\n", aurora.Index(colorIndex, name).Faint(), m)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Printf("%s | %s\n", aurora.Index(colorIndex, name).Faint(), m)
		}
	}()

	return nil
}
