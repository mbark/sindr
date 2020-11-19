package main

import (
	"fmt"
	"sync"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func getRunModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"async": async,
			"await": await,
			"watch": watch,
		},
	}
}

func async(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-2)
	fn, ok := lv.(*lua.LFunction)
	if !ok {
		L.TypeError(1, lua.LTFunction)
	}

	lv = L.Get(-1)

	runtime.wg.Add(1)
	Lt, _ := L.NewThread()
	go func() {
		defer runtime.wg.Done()

		err := Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true}, lv)
		if err != nil {
			L.RaiseError(err.Error())
		}
	}()

	return NoReturnVal, nil
}

func await(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	runtime.wg.Wait()
	return NoReturnVal, nil
}

type watchOpts = map[string]struct {
	Fn    *lua.LFunction
	Args  interface{}
	Watch string
}

type watchCommand struct {
	Run   func(*lua.LState)
	Watch string
}

func watch(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	tbl, ok := lv.(*lua.LTable)
	if !ok {
		L.TypeError(1, lua.LTTable)
	}

	var opts watchOpts
	if err := gluamapper.Map(tbl, &opts); err != nil {
		L.ArgError(1, fmt.Errorf("invalid config: %w", err).Error())
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

		run := func(L *lua.LState) {
			err := L.CallByParam(lua.P{Fn: c.Fn, NRet: 1, Protect: true}, largs)
			if err != nil {
				L.RaiseError(err.Error())
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
			close, err := startWatching(runtime, cmd.Watch, onChange)
			defer close()
			if err != nil {
				panic(fmt.Errorf("start watching %s: %w", cmd.Watch, err))
			}

			for {
				Lt, cancel := L.NewThread()

				cmd.Run(Lt)
				runtime.logger.Info("command started", zap.String("name", name))

				_ = <-onChange

				runtime.logger.Info("restarting", zap.String("name", name))
				cancel()
			}
		}(k, c, colorIdx)
	}

	wg.Wait()

	return NoReturnVal, nil
}
