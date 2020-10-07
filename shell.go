package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"github.com/logrusorgru/aurora/v3"
	lua "github.com/yuin/gopher-lua"
)

func getShellModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"run":   run(runtime.addCommand),
			"start": start(runtime.addCommand),
		},
	}
}

func run(addCommand func(cmd Command)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		if lv.Type() != lua.LTString {
			panic("string required.")
		}

		if str, ok := lv.(lua.LString); ok {
			addCommand(Command{
				pre: func() int64 {
					return -1
				},
				run: func() {
					fmt.Printf("running command %s\n", string(str))
					cmd := exec.Command("bash", "-c", fmt.Sprintf("%s", string(str)))
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
				pre: func() int64 {
					return -1
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
		}

		if tbl, ok := lv.(*lua.LTable); ok {
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
		}

		panic(fmt.Sprintf("unexpected type for start, got %v", lv.Type()))
	}
}
