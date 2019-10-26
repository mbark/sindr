package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

type deleteConfig struct {
	Files           string
	OnlyDirectories bool
}

func getFileModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"delete": delete(runtime.addCommand),
		},
	}
}

func removeGlob(glob string, onlyDirectories bool) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		panic(err)
	}

	fmt.Printf("glob %s matches %s\n", glob, matches)

	for _, file := range matches {
		fmt.Printf("found matching file %s\n", file)
		stat, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		if onlyDirectories && !stat.Mode().IsDir() {
			continue
		}

		fmt.Printf("removing %s\n", file)
		err = os.RemoveAll(file)
		if err != nil {
			panic(err)
		}
	}
}

func delete(addCommand func(cmd func())) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		if glob, ok := lv.(lua.LString); ok {
			addCommand(func() {
				removeGlob(string(glob), false)
			})

			return 0
		} else if tbl, ok := lv.(*lua.LTable); ok {
			addCommand(func() {
				var config deleteConfig
				if err := gluamapper.Map(tbl, &config); err != nil {
					panic(err)
				}

				removeGlob(config.Files, config.OnlyDirectories)
			})

			return 0
		}

		panic("string required.")
	}
}
