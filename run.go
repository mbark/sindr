package main

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func getRunModule() Module {
	return Module{
		exports: map[string]ModuleFunction{
			"async": async,
			"await": await,
			"watch": watch,
			"pool":  pool,
		},
	}
}

func pool(_ *Runtime, L *lua.LState) ([]lua.LValue, error) {
	ud := L.NewUserData()
	ud.Value = &Pool{wg: sync.WaitGroup{}}
	L.SetMetatable(ud, L.GetTypeMetatable(PoolType{}.TypeName()))
	return []lua.LValue{ud}, nil
}

var _ LuaType = new(PoolType)

type PoolType struct{}

func (p PoolType) TypeName() string {
	return "pool"
}

func (p PoolType) Funcs() map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"run":  p.Run,
		"wait": p.Wait,
	}
}

func (PoolType) Run(L *lua.LState) int {
	p := IsUserData[*Pool](L)
	fn, err := MapFunction(2, L.Get(2))
	if err != nil {
		L.RaiseError(err.Error())
	}

	p.wg.Add(1)
	go func() {
		Lt, cancel := L.NewThread()
		defer cancel()
		defer p.wg.Done()

		err = Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
		if err != nil {
			L.RaiseError(err.Error())
		}
	}()

	return 0
}

func (PoolType) Wait(L *lua.LState) int {
	p := IsUserData[*Pool](L)
	p.wg.Wait()
	return 0
}

type Pool struct {
	wg sync.WaitGroup
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

		for {
			done := runWatchFnOnce(runtime, L, fn, onChange)
			if done {
				return
			}

			log.Info("changes detected, running function")
		}
	}()

	return NoReturnVal, nil
}

func runWatchFnOnce(runtime *Runtime, L *lua.LState, fn *lua.LFunction, onChange chan bool) bool {
	Lt, cancel := L.NewThread()
	defer cancel()

	done := make(chan bool)
	go func() {
		err := Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
		var lerr *lua.ApiError
		if errors.As(err, &lerr) && strings.HasSuffix(lerr.Object.String(), "signal: killed") {
			runtime.logger.With(slog.Any("err", err)).Info("function killed")
		} else if err != nil {
			runtime.logger.With(slog.Any("err", err)).Error("function error")
		}
		<-done
	}()

	runtime.logger.Info("waiting for changes")
	select {
	case <-Lt.Context().Done():
		return true
	case <-done:
		return false
	case <-onChange:
		return false
	}
}
