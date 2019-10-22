package files

import (
	"os"
	"path/filepath"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type deleteConfig struct {
	FileGlob        string
	OnlyDirectories bool
}

// Loader ...
func Loader(L *lua.LState) int {
	mod := L.SetFuncs(L.NewTable(), exports)

	L.Push(mod)
	return 1
}

var exports = map[string]lua.LGFunction{
	"delete": delete,
}

func removeGlob(glob string, onlyDirectories bool) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		panic(err)
	}

	for _, file := range matches {
		stat, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		if onlyDirectories && !stat.Mode().IsDir() {
			continue
		}

		err = os.RemoveAll(file)
		if err != nil {
			panic(err)
		}
	}
}

func delete(L *lua.LState) int {
	lv := L.Get(-1)
	if glob, ok := lv.(lua.LString); ok {
		removeGlob(string(glob), false)
		return 0
	} else if tbl, ok := lv.(*lua.LTable); ok {
		var config deleteConfig
		if err := gluamapper.Map(tbl, &config); err != nil {
			panic(err)
		}

		removeGlob(config.FileGlob, config.OnlyDirectories)

		return 0
	}

	panic("string required.")
}
