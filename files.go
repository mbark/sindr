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
			"delete": delete(runtime, runtime.addCommand),
		},
	}
}

func findGlobMatches(config deleteConfig) []string {
	matches, err := filepath.Glob(config.Files)
	if err != nil {
		panic(err)
	}

	fmt.Printf("glob %s matches %s\n", config.Files, matches)

	var files []string
	for _, file := range matches {
		fmt.Printf("found matching file %s\n", file)
		stat, err := os.Stat(file)
		if err != nil {
			panic(err)
		}

		if config.OnlyDirectories && !stat.Mode().IsDir() {
			continue
		}

		fmt.Printf("removing %s\n", file)
		files = append(files, file)
	}

	return files
}

func removeFiles(files []string) {
	for _, file := range files {
		fmt.Printf("removing file %s\n", file)
		err := os.RemoveAll(file)
		if err != nil {
			panic(err)
		}
	}
}

func delete(runtime *Runtime, addCommand func(cmd Command)) lua.LGFunction {
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
