package shell

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// Loader ...
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)

	L.Push(mod)
	return 1
}

var exports = map[string]lua.LGFunction{
	"run": run,
}

func run(L *lua.LState) int {
	fmt.Println("called shell.run")
	return 0
}
