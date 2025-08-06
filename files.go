package main

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	lua "github.com/yuin/gopher-lua"
)

func getFileModule(_ *Runtime) Module {
	return Module{
		exports: map[string]ModuleFunction{
			"write":     write,
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

func write(_ *Runtime, L *lua.LState) ([]lua.LValue, error) {
	fileName, err := MapString(1, L.Get(1))
	if err != nil {
		return nil, err
	}
	content, err := MapString(2, L.Get(2))
	if err != nil {
		return nil, err
	}

	err = os.WriteFile(fileName, []byte(content), 0600)
	if err != nil {
		return nil, err
	}

	return NoReturnVal, nil
}

type deleteConfig struct {
	Files           string
	OnlyDirectories bool
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
	dir, err := MapString(1, L.Get(-1))
	if err != nil {
		return nil, err
	}
	var options mkdirOptions
	if L.GetTop() >= 2 {
		err := MapTable(2, L.Get(2), &options)
		if err != nil {
			return nil, err
		}
	}

	if options.All {
		err = os.MkdirAll(dir, 0700)
	} else {
		err = os.Mkdir(dir, 0700)
	}
	if errors.Is(err, os.ErrExist) {
		return NoReturnVal, nil
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
