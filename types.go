package main

import lua "github.com/yuin/gopher-lua"

func IsUserData[T any](L *lua.LState) T {
	var t T

	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(T); ok {
		return v
	}
	L.ArgError(1, "shmake expected")
	return t
}

type LuaType interface {
	TypeName() string
	GlobalName() string
	New(L *lua.LState) int
	Funcs() map[string]lua.LGFunction
}

func RegisterLuaTypes(L *lua.LState, types ...LuaType) {
	for _, lType := range types {
		mt := L.NewTypeMetatable(lType.TypeName())
		L.SetGlobal(lType.GlobalName(), mt)
		L.SetField(mt, "new", L.NewFunction(lType.New))
		L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), lType.Funcs()))
	}
}
