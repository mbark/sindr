package main

import (
	"context"
	"fmt"
	"io"
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
			"copy":      copy(runtime.addCommand),
			"mkdir":     mkdir(runtime.addCommand),
			"newest_ts": timestamp(true),
			"oldest_ts": timestamp(false),
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
			run: func(ctx context.Context) {
				fmt.Println("deleting files", config)
				files := findGlobMatches(config)
				removeFiles(files)
			},
		})

		return 0
	}
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

type copyOptions struct {
	From string
	To   string
}

func copy(addCommand func(cmd Command)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var opts copyOptions
		if tbl, ok := lv.(*lua.LTable); ok {
			if err := gluamapper.Map(tbl, &opts); err != nil {
				panic(err)
			}
		} else {
			panic("table expected")
		}

		addCommand(Command{
			run: func(ctx context.Context) {
				err := CopyFile(opts.From, opts.To)
				if err != nil {
					panic(err)
				}
			},
		})

		return 0
	}
}

type mkdirOptions struct {
	Dir string
	All bool
}

func mkdir(addCommand func(cmd Command)) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var opts mkdirOptions
		if str, ok := lv.(lua.LString); ok {
			opts.Dir = string(str)
			opts.All = false
		} else if tbl, ok := lv.(*lua.LTable); ok {
			if err := gluamapper.Map(tbl, &opts); err != nil {
				panic(err)
			}
		} else {
			panic("table expected")
		}

		addCommand(Command{
			run: func(ctx context.Context) {
				var err error
				if opts.All {
					err = os.MkdirAll(opts.Dir, 0700)
				} else {
					err = os.Mkdir(opts.Dir, 0700)
				}
				if err != nil {
					panic(err)
				}
			},
		})

		return 0
	}
}

func timestamp(useNewest bool) lua.LGFunction {
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

		var modTime *time.Time = nil
		for _, f := range files {
			stat, err := os.Stat(f)
			if err != nil {
				panic(err)
			}
			if modTime == nil {
				mt := stat.ModTime()
				modTime = &mt
			}

			if useNewest {
				if stat.ModTime().After(*modTime) {
					mt := stat.ModTime()
					modTime = &mt
				}
			} else {
				if stat.ModTime().Before(*modTime) {
					mt := stat.ModTime()
					modTime = &mt
				}
			}
		}

		L.Push(lua.LNumber(modTime.Unix()))
		return 1
	}
}
