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
	GlobalName() string
	New(L *lua.LState) int
	Funcs() map[string]lua.LGFunction
}

func RegisterLuaTypes(runtime *Runtime, L *lua.LState, types ...LuaType) {
	for _, lType := range types {
		mt := L.NewTypeMetatable(lType.TypeName())
		L.SetGlobal(lType.GlobalName(), mt)
		L.SetField(mt, "new", L.NewFunction(lType.New))
		L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), lType.Funcs()))
		runtime.logger.Info("registered new type", slog.String("type", lType.TypeName()))
	}
}
