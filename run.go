package main

import (
	"log/slog"

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
	fn, err := MapFunction(1, L.Get(1))
	if err != nil {
		return nil, err
	}

	runtime.wg.Add(1)
	Lt, _ := L.NewThread()
	go func() {
		defer runtime.wg.Done()

		err := Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
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

func watch(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	glob, err := MapString(1, L.Get(1))
	if err != nil {
		return nil, err
	}

	fn, err := MapFunction(2, L.Get(2))
	if err != nil {
		return nil, err
	}

	runtime.wg.Add(1)
	go func() {
		defer runtime.wg.Done()

		log := runtime.logger.With(slog.String("glob", glob))

		onChange := make(chan bool)
		close, err := startWatching(runtime, glob, onChange)
		defer close()
		if err != nil {
			log.With(slog.Any("err", err)).Error("starting watcher failed")
			return
		}

		Lt, cancel := L.NewThread()
		defer cancel()
		for {
			log.Info("waiting for changes")
			_ = <-onChange

			log.Info("changes detected, running function")
			err := Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
			if err != nil {
				L.RaiseError(err.Error())
			}
		}
	}()

	return NoReturnVal, nil
}
