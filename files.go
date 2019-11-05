package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

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
			"delete": delete(runtime, runtime.addCommand),
		},
	}
}

func removeGlob(glob string, onlyDirectories bool) (int64, bool) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		panic(err)
	}

	fmt.Printf("glob %s matches %s\n", glob, matches)

	removedFile := false
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

		removedFile = true
	}

	if removedFile {
		return time.Now().Unix(), true
	}

	return 0, false
}

func delete(runtime *Runtime, addCommand func(cmd func() int64)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		if glob, ok := lv.(lua.LString); ok {
			addCommand(func() int64 {
				lastRun, hasTimestamp := removeGlob(string(glob), false)
				if !hasTimestamp {
					return runtime.getLastTimestamp("removeGlob:" + string(glob))
				}

				return lastRun
			})

			return 0
		} else if tbl, ok := lv.(*lua.LTable); ok {
			addCommand(func() int64 {
				var config deleteConfig
				if err := gluamapper.Map(tbl, &config); err != nil {
					panic(err)
				}

				lastRun, hasTimestamp := removeGlob(config.Files, config.OnlyDirectories)
				if !hasTimestamp {
					return runtime.getLastTimestamp("removeGlob:" + config.Files)
				}

				return lastRun
			})

			return 0
		}

		panic("string required.")
	}
}
