package shmake

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

func write(_ *Runtime, l *lua.LState) ([]lua.LValue, error) {
	fileName, err := MapString(l, 1)
	if err != nil {
		return nil, err
	}
	content, err := MapString(l, 2)
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

func delete(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	lv := l.Get(-1)

	var config deleteConfig
	if glob, ok := lv.(lua.LString); ok {
		config.Files = string(glob)
		config.OnlyDirectories = false
	} else if _, ok := lv.(*lua.LTable); ok {
		cfg, err := MapTable[deleteConfig](l, 1)
		if err != nil {
			return nil, err
		}
		config = cfg
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

func copy(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	opts, err := MapTable[copyOptions](l, 1)
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

func mkdir(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	dir, err := MapString(l, 1)
	if err != nil {
		return nil, err
	}
	options, err := MapOptionalTable[mkdirOptions](l, 2)
	if err != nil {
		return nil, err
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
	return func(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
		lv := l.Get(-1)

		var config deleteConfig
		if glob, ok := lv.(lua.LString); ok {
			config.Files = string(glob)
			config.OnlyDirectories = false
		} else if _, ok := lv.(*lua.LTable); ok {
			cfg, err := MapTable[deleteConfig](l, 1)
			if err != nil {
				return nil, err
			}
			config = cfg
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

func chdir(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
	dir, err := MapString(l, 1)
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

func popdir(runtime *Runtime, l *lua.LState) ([]lua.LValue, error) {
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
