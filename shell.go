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
)

func getShellModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"run":   run(runtime, runtime.addCommand),
			"start": start(runtime.addCommand),
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

func run(runtime *Runtime, addCommand func(cmd Command)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		if lv.Type() != lua.LTString {
			panic("string required.")
		}

		if str, ok := lv.(lua.LString); ok {
			addCommand(Command{
				version: func() *string {
					return nil
				},
				run: func() {
					command := withVariables(runtime, string(str))
					fmt.Printf("running command %s\n", command)
					cmd := exec.Command("bash", "-c", command)
					err := cmd.Run()
					if err != nil {
						panic(err)
					}
				},
			})
		}

		return 0
	}
}

func start(addCommand func(cmd Command)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		if str, ok := lv.(lua.LString); ok {
			addCommand(Command{
				version: func() *string {
					return nil
				},
				run: func() {
					cmd := exec.Command("bash", "-c", fmt.Sprintf("%s", string(str)))
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr
					err := cmd.Start()
					if err != nil {
						panic(err)
					}
				},
			})

			return 0
		} else if tbl, ok := lv.(*lua.LTable); ok {
			startCommands := make(map[string]string)
			tbl.ForEach(func(luaKey lua.LValue, luaVal lua.LValue) {
				var key, value string
				if k, ok := luaKey.(lua.LString); ok {
					key = string(k)
				} else {
					panic(fmt.Sprintf("key must string, got %v", luaKey.Type()))
				}

				if v, ok := luaVal.(lua.LString); ok {
					value = string(v)
				} else {
					panic(fmt.Sprintf("value must string, got %v", luaKey.Type()))
				}

				startCommands[key] = value
			})

			addCommand(Command{
				version: func() *string {
					return nil
				},
				run: func() {
					fmt.Println("start run()")
					var colorIdx uint8 = 0

					wg := sync.WaitGroup{}
					for k, c := range startCommands {
						wg.Add(1)
						colorIdx += 1
						go func(name, command string, colorIndex uint8) {
							defer wg.Done()

							cmd := exec.Command("bash", "-c", fmt.Sprintf("%s", command))

							stdout, err := cmd.StdoutPipe()
							if err != nil {
								panic(err)
							}

							stderr, err := cmd.StderrPipe()
							if err != nil {
								panic(err)
							}

							if err := cmd.Start(); err != nil {
								panic(err)
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

							if err := cmd.Wait(); err != nil {
								panic(err)
							}
						}(k, c, colorIdx)
					}

					wg.Wait()
				},
			})

			return 0
		}

		panic(fmt.Sprintf("unexpected type for start, got %v", lv.Type()))
	}
}
