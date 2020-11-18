package main

import (
	"fmt"

	"github.com/peterbourgon/diskv/v3"
	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

func getCacheModule(runtime *Runtime) Module {
	return Module{
		exports: map[string]lua.LGFunction{
			"diff":  diff(runtime),
			"store": store(runtime),
		},
	}
}

type Cache struct {
	diskv *diskv.Diskv

	ForceOutOfDate bool // ForceOutOfDate makes all gets return nil
}

func NewCache(file string) Cache {
	return Cache{
		diskv: diskv.New(diskv.Options{
			BasePath:     file,
			Transform:    func(s string) []string { return []string{} },
			CacheSizeMax: 1024 * 1024,
		}),
	}
}

func (c Cache) StoreVersion(name, version string) error {
	return c.diskv.Write(name, []byte(version))
}

func (c Cache) GetVersion(name string) (*string, error) {
	if c.ForceOutOfDate {
		return nil, nil
	}

	if !c.diskv.Has(name) {
		return nil, nil
	}

	value, err := c.diskv.Read(name)
	if err != nil {
		return nil, fmt.Errorf("cache read: %w", err)
	}

	val := string(value)
	return &val, nil
}

type cacheDiffOptions struct {
	Name    string
	Version string
}

func diff(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		runtime.logger.Info("running cache_diff command")

		lv := L.Get(-1)

		var options cacheDiffOptions
		if tbl, ok := lv.(*lua.LTable); ok {
			if err := gluamapper.Map(tbl, &options); err != nil {
				panic(fmt.Errorf("failed to map table %+v: %w", tbl, err))
			}
		} else {
			panic(fmt.Sprintf("table expected, got %s", lv.Type()))
		}

		currentVersion, err := runtime.cache.GetVersion(options.Name)
		if err != nil {
			panic(err)
		}

		isDiff := currentVersion == nil || *currentVersion != options.Version

		L.Push(lua.LBool(isDiff))
		return 1
	}
}

func store(runtime *Runtime) lua.LGFunction {
	return func(L *lua.LState) int {
		lv := L.Get(-1)

		var options cacheDiffOptions
		if tbl, ok := lv.(*lua.LTable); ok {
			if err := gluamapper.Map(tbl, &options); err != nil {
				panic(fmt.Errorf("failed to map table %+v: %w", tbl, err))
			}
		} else {
			panic(fmt.Sprintf("table expected, got %s", lv.Type()))
		}

		runtime.logger.Info("storing cache version",
			zap.String("name", options.Name),
			zap.String("version", options.Version))
		runtime.cache.StoreVersion(options.Name, options.Version)

		return 0
	}
}
