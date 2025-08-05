package main

import (
	"fmt"
	"log/slog"

	lua "github.com/yuin/gopher-lua"
)

func IsUserData[T any](L *lua.LState) T {
	var t T

	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(T); ok {
		return v
	}
	L.ArgError(1, fmt.Sprintf("expected %T, got %T", t, ud.Value))
	return t
}

func NewUserData[T any, LT LuaType](L *lua.LState, t T, lt LT) int {
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, L.GetTypeMetatable(lt.TypeName()))
	L.Push(ud)
	return 1
}

type LuaType interface {
	TypeName() string
	Funcs() map[string]lua.LGFunction
}

type LuaTypeNewer interface {
	New(L *lua.LState) int
}

type LuaTypeGlobal interface {
	GlobalName() string
}

func RegisterLuaTypes(runtime *Runtime, L *lua.LState, types ...LuaType) {
	for _, lType := range types {
		logger := runtime.logger.With(slog.String("type", lType.TypeName()))
		mt := L.NewTypeMetatable(lType.TypeName())
		L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), lType.Funcs()))

		if lType, ok := lType.(LuaTypeGlobal); ok {
			L.SetGlobal(lType.GlobalName(), mt)
			logger.With(slog.String("global", lType.GlobalName())).Debug("registered type with global name")
		}
		if nw, ok := lType.(LuaTypeNewer); ok {
			L.SetField(mt, "new", L.NewFunction(nw.New))
			logger.With(slog.String("new", "new")).Debug("registered type with new function")
		}

		logger.Info("registered new type")
	}
}
