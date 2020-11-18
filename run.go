package main

import (
	"context"
	"sync"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func getRunModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"async": async(runtime),
			"await": await(runtime),
			"watch": watch(runtime),
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

		err := L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lv)
		if err != nil {
			panic(err)
		}

		return 0
	}
}

func await(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		runtime.wg.Wait()
		return 0
	}
}

type watchOpts = map[string]struct {
	Fn    *lua.LFunction
	Args  interface{}
	Watch string
}

type watchCommand struct {
	Run   func()
	Watch string
}

func watch(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		tbl, ok := lv.(*lua.LTable)
		if !ok {
			panic("argument must be a table")
		}

		var opts watchOpts
		if err := gluamapper.Map(tbl, &opts); err != nil {
			runtime.logger.Fatal("failed to map commands", zap.Error(err))
		}

		cmds := make(map[string]watchCommand)

		for k, c := range opts {
			var largs lua.LValue
			switch a := c.Args.(type) {
			case string:
				largs = lua.LString(a)

			case []interface{}:
				tbl = &lua.LTable{}

			case map[string]string:
				largs = &lua.LTable{}

			default:
				largs = lua.LNil
			}

			run := func() {
				err := L.CallByParam(lua.P{Fn: c.Fn, NRet: 1, Protect: true}, largs)
				if err != nil {
					panic(err)
				}
			}

			cmds[k] = watchCommand{Run: run, Watch: c.Watch}
		}

		var colorIdx uint8 = 0

		var wg sync.WaitGroup
		for k, c := range cmds {
			runtime.logger.Debug("starting",
				zap.String("name", k),
				zap.String("watch", c.Watch))

			wg.Add(1)
			colorIdx += 1
			go func(name string, cmd watchCommand, colorIndex uint8) {
				defer wg.Done()

				onChange := make(chan bool)
				close := startWatching(runtime, cmd.Watch, onChange)
				defer close()

				for {
					ctx, cancel := context.WithCancel(L.Context())
					L.SetContext(ctx)

					cmd.Run()
					runtime.logger.Info("command started", zap.String("name", name))

					_ = <-onChange

					runtime.logger.Info("restarting", zap.String("name", name))
					cancel()
				}
			}(k, c, colorIdx)
		}

		wg.Wait()

		return 0
	}
}
