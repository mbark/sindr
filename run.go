package main

import (
	"log/slog"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func getRunModule() Module {
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
	fn, err := MapFunction(1, lv)
	if err != nil {
		return nil, err
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

type watchedFn struct {
	Run   func(*lua.LState)
	Watch string
}

func watch(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var opts watchOpts
	err := MapTable(1, lv, &opts)
	if err != nil {
		return nil, err
	}

	cmds := make(map[string]watchedFn)

	for k, c := range opts {
		var largs lua.LValue
		switch a := c.Args.(type) {
		case string:
			largs = lua.LString(a)

		case []interface{}:

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

		cmds[k] = watchedFn{Run: run, Watch: c.Watch}
	}

	var colorIdx uint8 = 0

	var wg sync.WaitGroup
	for k, c := range cmds {
		wg.Add(1)
		colorIdx += 1
		go func(name string, cmd watchedFn, colorIndex uint8) {
			defer wg.Done()

			log := runtime.logger.With(slog.String("watch", cmd.Watch)).With(slog.String("name", name))

			onChange := make(chan bool)
			close, err := startWatching(runtime, cmd.Watch, onChange)
			defer close()
			if err != nil {
				log.With(slog.Any("err", err)).Error("starting watcher failed")
				return
			}

			for {
				Lt, cancel := L.NewThread()

				log.Debug("running fn")
				cmd.Run(Lt)

				log.Debug("waiting for change")
				_ = <-onChange

				log.Info("restarting")
				cancel()
			}
		}(k, c, colorIdx)
	}

	wg.Wait()
	return NoReturnVal, nil
}
