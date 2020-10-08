package main

import (
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
			"delete":    delete(runtime.addCommand),
			"timestamp": timestamp(),
		},
	}
}

func findGlobMatches(config deleteConfig) []string {
	matches, err := filepath.Glob(config.Files)
	if err != nil {
		panic(err)
	}

	var files []string
	for _, file := range matches {
		stat, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		if config.OnlyDirectories && !stat.Mode().IsDir() {
			continue
		}

		files = append(files, file)
	}

	return files
}

func removeFiles(files []string) {
	for _, file := range files {
		err := os.RemoveAll(file)
		if err != nil {
			panic(err)
		}
	}
}

func delete(addCommand func(cmd Command)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var config deleteConfig
		if glob, ok := lv.(lua.LString); ok {
			config.Files = string(glob)
			config.OnlyDirectories = false
		} else if tbl, ok := lv.(*lua.LTable); ok {
			if err := gluamapper.Map(tbl, &config); err != nil {
				panic(err)
			}
		} else {
			panic("string or table expected")
		}

		addCommand(Command{
			version: func() *string {
				files := findGlobMatches(config)
				if len(files) > 0 {
					return nil
				}

				return &AlwaysUpToDateVersion
			},
			run: func() {
				files := findGlobMatches(config)
				removeFiles(files)
			},
		})

		return 0
	}
}

func timestamp() lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var config deleteConfig
		if glob, ok := lv.(lua.LString); ok {
			config.Files = string(glob)
			config.OnlyDirectories = false
		} else if tbl, ok := lv.(*lua.LTable); ok {
			if err := gluamapper.Map(tbl, &config); err != nil {
				panic(err)
			}
		} else {
			panic("string or table expected")
		}

		files := findGlobMatches(config)
		if len(files) == 0 {
			L.Push(lua.LNumber(0))
			return 1
		}

		var modTime time.Time
		for _, f := range files {
			stat, err := os.Stat(f)
			if err != nil {
				panic(err)
			}

			if stat.ModTime().After(modTime) {
				modTime = stat.ModTime()
			}
		}

		L.Push(lua.LNumber(modTime.Unix()))
		return 1
	}
}
