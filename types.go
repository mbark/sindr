package shmake

import (
	"fmt"
	"log/slog"

	lua "github.com/yuin/gopher-lua"
)

func IsUserData[T any](l *lua.LState) T {
	var t T

	ud := l.CheckUserData(1)
	if v, ok := ud.Value.(T); ok {
		return v
	}
	l.ArgError(1, fmt.Sprintf("expected %T, got %T", t, ud.Value))
	return t
}

func NewUserData[T any, LT LuaType](l *lua.LState, t T, lt LT) int {
	ud := l.NewUserData()
	ud.Value = t
	l.SetMetatable(ud, l.GetTypeMetatable(lt.TypeName()))
	l.Push(ud)
	return 1
}

type LuaType interface {
	TypeName() string
	Funcs() map[string]lua.LGFunction
}

type LuaTypeNewer interface {
	New(l *lua.LState) int
}

type LuaTypeGlobal interface {
	GlobalName() string
}

func RegisterLuaTypes(runtime *Runtime, l *lua.LState, types ...LuaType) {
	for _, lType := range types {
		logger := slog.With(slog.String("type", lType.TypeName()))
		mt := l.NewTypeMetatable(lType.TypeName())
		l.SetField(mt, "__index", l.SetFuncs(l.NewTable(), lType.Funcs()))

		if lType, ok := lType.(LuaTypeGlobal); ok {
			l.SetGlobal(lType.GlobalName(), mt)
			logger.With(slog.String("global", lType.GlobalName())).Debug("registered type with global name")
		}
		if nw, ok := lType.(LuaTypeNewer); ok {
			l.SetField(mt, "new", l.NewFunction(nw.New))
			logger.With(slog.String("new", "new")).Debug("registered type with new function")
		}

		logger.Debug("registered new type")
	}
}
