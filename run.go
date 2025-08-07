package shmake

import (
	"errors"
	"log/slog"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
)

func pool(_ *Runtime, l *lua.LState) ([]lua.LValue, error) {
	ud := l.NewUserData()
	ud.Value = &Pool{wg: sync.WaitGroup{}}
	l.SetMetatable(ud, l.GetTypeMetatable(PoolType{}.TypeName()))
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

func (PoolType) Run(l *lua.LState) int {
	p := IsUserData[*Pool](l)
	fn, err := MapFunction(l, 2)
	if err != nil {
		l.RaiseError(err.Error())
	}

	p.wg.Add(1)
	go func() {
		Lt, cancel := l.NewThread()
		defer cancel()
		defer p.wg.Done()

		err = Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
		if err != nil {
			l.RaiseError(err.Error())
		}
	}()

	return 0
}

func (PoolType) Wait(l *lua.LState) int {
	p := IsUserData[*Pool](l)
	p.wg.Wait()
	return 0
}

type Pool struct {
	wg sync.WaitGroup
}

func async(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	fn, err := MapFunction(l, 1)
	if err != nil {
		return nil, err
	}

	runtime.wg.Add(1)
	Lt, _ := l.NewThread()
	go func() {
		defer runtime.wg.Done()

		err := Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
		if err != nil {
			l.RaiseError(err.Error())
		}
	}()

	return NoReturnVal, nil
}

func wait(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	runtime.wg.Wait()
	return NoReturnVal, nil
}

func watch(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	glob, err := MapString(l, 1)
	if err != nil {
		return nil, err
	}

	fn, err := MapFunction(l, 2)
	if err != nil {
		return nil, err
	}

	runtime.wg.Add(1)
	go func() {
		defer runtime.wg.Done()

		log := slog.With(slog.String("glob", glob))

		onChange := make(chan bool)
		close, err := startWatching(runtime, glob, onChange)
		defer close()
		if err != nil {
			log.With(slog.Any("err", err)).Error("starting watcher failed")
			return
		}

		for {
			done := runWatchFnOnce(runtime, l, fn, onChange)
			if done {
				return
			}

			log.Info("changes detected, running function")
		}
	}()

	return NoReturnVal, nil
}

func runWatchFnOnce(runtime *Runtime, l *lua.LState, fn *lua.LFunction, onChange chan bool) bool {
	Lt, cancel := l.NewThread()
	defer cancel()

	done := make(chan bool)
	go func() {
		err := Lt.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
		var lerr *lua.ApiError
		if errors.As(err, &lerr) && strings.HasSuffix(lerr.Object.String(), "signal: killed") {
			slog.With(slog.Any("err", err)).Debug("function killed")
		} else if err != nil {
			l.RaiseError("%s", err.Error())
		}
		<-done
	}()

	slog.Debug("waiting for changes")
	select {
	case <-Lt.Context().Done():
		return true
	case <-done:
		return false
	case <-onChange:
		return false
	}
}
