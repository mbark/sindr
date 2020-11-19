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
		exports: map[string]ModuleFunction{
			"diff":  diff,
			"store": store,
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

func diff(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var options cacheDiffOptions
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		L.TypeError(1, lua.LTTable)
	}

	if err := gluamapper.Map(tbl, &options); err != nil {
		L.ArgError(1, fmt.Errorf("invalid config: %w", err).Error())
	}

	currentVersion, err := runtime.cache.GetVersion(options.Name)
	if err != nil {
		return nil, err
	}

	isDiff := currentVersion == nil || *currentVersion != options.Version

	return []lua.LValue{lua.LBool(isDiff)}, nil
}

func store(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	lv := L.Get(-1)

	var options cacheDiffOptions
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		L.TypeError(1, lua.LTTable)
	}

	if err := gluamapper.Map(tbl, &options); err != nil {
		L.ArgError(1, fmt.Errorf("invalid config: %w", err).Error())
	}

	runtime.logger.
		With(zap.String("version", options.Version)).
		With(zap.String("name", options.Name)).
		Info("storing cache version")

	runtime.cache.StoreVersion(options.Name, options.Version)

	return NoReturnVal, nil
}
