package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type deleteConfig struct {
	Files           string
	OnlyDirectories bool
}

func getFileModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"delete":    delete,
			"copy":      copy,
			"mkdir":     mkdir,
			"chdir":     chdir,
			"popdir":    popdir,
			"newest_ts": timestamp(true),
			"oldest_ts": timestamp(false),
		},
	}
}

func findGlobMatches(config deleteConfig) ([]string, error) {
	matches, err := filepath.Glob(config.Files)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, file := range matches {
		stat, err := os.Stat(file)
		if err != nil {
			return nil, err
		}

		if config.OnlyDirectories && !stat.Mode().IsDir() {
			continue
		}

		files = append(files, file)
	}

	return files, nil
}

func delete(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var config deleteConfig
	if glob, ok := lv.(lua.LString); ok {
		config.Files = string(glob)
		config.OnlyDirectories = false
	} else if _, ok := lv.(*lua.LTable); ok {
		err := MapTable(1, lv, &config)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, ErrBadArg{Index: 1, Message: "string or table expected"}
	}

	files, err := findGlobMatches(config)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		err := os.RemoveAll(file)
		if err != nil {
			return nil, err
		}
	}

	return NoReturnVal, nil
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

func copy(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var opts copyOptions
	err := MapTable(1, lv, &opts)
	if err != nil {
		return nil, err
	}

	err = CopyFile(opts.From, opts.To)
	if err != nil {
		return nil, err
	}

	return NoReturnVal, nil
}

type mkdirOptions struct {
	Dir string
	All bool
}

func mkdir(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var opts mkdirOptions
	if str, ok := lv.(lua.LString); ok {
		opts.Dir = string(str)
		opts.All = false
	} else if _, ok := lv.(*lua.LTable); ok {
		err := MapTable(1, lv, &opts)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, ErrBadArg{Index: 1, Message: "string or table expected"}
	}

	var err error
	if opts.All {
		err = os.MkdirAll(opts.Dir, 0700)
	} else {
		err = os.Mkdir(opts.Dir, 0700)
	}
	if err != nil {
		return nil, err
	}

	return NoReturnVal, nil
}

func timestamp(useNewest bool) ModuleFunction {
	return func(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
		lv := L.Get(-1)

		var config deleteConfig
		if glob, ok := lv.(lua.LString); ok {
			config.Files = string(glob)
			config.OnlyDirectories = false
		} else if _, ok := lv.(*lua.LTable); ok {
			err := MapTable(1, lv, &config)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, ErrBadArg{Index: 1, Message: "string or table expected"}
		}

		files, err := findGlobMatches(config)
		if err != nil {
			return nil, err
		}

		if len(files) == 0 {
			return []lua.LValue{lua.LNumber(0)}, nil
		}

		var modTime *time.Time = nil
		for _, f := range files {
			stat, err := os.Stat(f)
			if err != nil {
				return nil, err
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

		return []lua.LValue{lua.LNumber(modTime.Unix())}, nil
	}
}

func chdir(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)
	dir, err := MapString(1, lv)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	runtime.prevDir = cwd
	err = os.Chdir(dir)
	if err != nil {
		return nil, err
	}

	return NoReturnVal, nil
}

func popdir(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	if runtime.prevDir == "" {
		return nil, errors.New("no previous directory stored")
	}

	err := os.Chdir(runtime.prevDir)
	if err != nil {
		return nil, err
	}
	runtime.prevDir = ""

	return NoReturnVal, nil
}
