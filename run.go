package main

import (
	lua "github.com/yuin/gopher-lua"
)

func getRunModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"async": async(runtime),
			"await": await(runtime),
		},
	}
}

func async(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-2)
		fn, ok := lv.(*lua.LFunction)
		if !ok {
			panic("first argument must be a function")
		}

		lv = L.Get(-1)

		runtime.addCommand(Command{
			run: func() {
				runtime.runAsync = true
				defer func() {
					runtime.runAsync = false
				}()

				err := L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lv)
				if err != nil {
					panic(err)
				}
			},
		})

		return 0
	}
}

func await(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		runtime.addCommand(Command{
			run: func() {
				runtime.wg.Wait()
			},
		})

		return 0
	}
}
