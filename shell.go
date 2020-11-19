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
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func getShellModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"run":    run,
			"output": output,
			"start":  start,
		},
	}
}

func withVariables(runtime *Runtime, input string) string {
	t := template.Must(template.New("run").Parse(input))
	var buf bytes.Buffer
	err := t.Execute(&buf, runtime.variables)
	if err != nil {
		runtime.logger.With(zap.Error(err)).Fatal("execute template")
	}

	return buf.String()
}

func run(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)
	c, err := MapString(1, lv)
	if err != nil {
		return nil, err
	}

	command := withVariables(runtime, c)

	runtime.logger.With(zap.String("command", command)).Debug("running shell command")

	cmd := exec.CommandContext(L.Context(), "bash", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("running shell cmd failed: %w", err)
	}

	return NoReturnVal, nil
}

func output(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	c, err := MapString(1, lv)
	if err != nil {
		return nil, err
	}

	command := withVariables(runtime, c)

	runtime.logger.With(zap.String("command", command)).Debug("running shell command and returning output")

	cmd := exec.CommandContext(L.Context(), "bash", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("running shell cmd failed: %w", err)
	}

	return []lua.LValue{lua.LString(string(output))}, nil
}

type startOptions = map[string]struct {
	Cmd   string
	Watch string
}

func start(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var startCommands startOptions
	err := MapTable(1, lv, &startCommands)
	if err != nil {
		return nil, err
	}

	for k, c := range startCommands {
		c.Cmd = withVariables(runtime, c.Cmd)
		startCommands[k] = c
	}

	var colorIdx uint8 = 0

	wg := sync.WaitGroup{}
	for k, c := range startCommands {
		log := runtime.logger.
			With(zap.String("name", k)).
			With(zap.String("command", c.Cmd)).
			With(zap.String("watch", c.Watch))

		wg.Add(1)
		colorIdx += 1
		go func(name, command, watch string, colorIndex uint8) {
			defer wg.Done()

			if watch == "" {
				cmd := exec.CommandContext(L.Context(), "bash", "-c", fmt.Sprintf("%s", command))
				err := startShellCmd(cmd, name, colorIndex)
				if err != nil {
					runtime.logger.With(zap.Error(err)).Fatal("start command")
				}

				log.Debug("shell command started")

				if err := cmd.Wait(); err != nil {
					log.With(zap.Error(err)).Fatal("shell command failed")
				}
			} else {
				onChange := make(chan bool)
				close, err := startWatching(runtime, watch, onChange)
				defer close()
				if err != nil {
					log.With(zap.Error(err)).Panic("failed to start watcher")
				}

				for {
					Lt, cancel := L.NewThread()
					cmd := exec.CommandContext(Lt.Context(), "bash", "-c", fmt.Sprintf("%s", command))
					err := startShellCmd(cmd, name, colorIndex)
					if err != nil {
						log.With(zap.Error(err)).Fatal("start shell command failed")
					}
					log.Debug("shell command started")

					_ = <-onChange

					log.Debug("shell command restarting")
					cancel()
				}
			}
		}(k, c.Cmd, c.Watch, colorIdx)
	}
	wg.Wait()

	return NoReturnVal, nil
}

func startShellCmd(cmd *exec.Cmd, name string, colorIndex uint8) error {
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
