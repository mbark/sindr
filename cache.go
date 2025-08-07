package shmake

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/peterbourgon/diskv/v3"
	lua "github.com/yuin/gopher-lua"
)

type Cache struct {
	diskv          *diskv.Diskv
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
	Name       string
	Version    string
	IntVersion int
}

func mapCacheDiffOptions(L *lua.LState, idx int, stackIdx int) (*cacheDiffOptions, error) {
	var options cacheDiffOptions
	err := MapTable(idx, L.Get(stackIdx), &options)
	if err != nil {
		return nil, err
	}
	if options.Version == "" {
		options.Version = strconv.Itoa(options.IntVersion)
	}

	return &options, nil
}

func diff(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	options, err := mapCacheDiffOptions(L, 1, -1)
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(runtime, *options)
	if err != nil {
		return nil, err
	}
	return []lua.LValue{lua.LBool(isDiff)}, nil
}

func store(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	options, err := mapCacheDiffOptions(L, 1, -1)
	if err != nil {
		return nil, err
	}

	slog.
		With(slog.String("version", options.Version)).
		With(slog.String("name", options.Name)).
		Debug("storing cache version")

	err = runtime.cache.StoreVersion(options.Name, options.Version)
	return NoReturnVal, nil
}

func withVersion(runtime *Runtime, L *lua.LState) ([]lua.LValue, error) {
	options, err := mapCacheDiffOptions(L, 1, 1)
	if err != nil {
		return nil, err
	}

	fn, err := MapFunction(2, L.Get(2))
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(runtime, *options)
	if err != nil {
		return nil, err
	}
	if !isDiff {
		return NoReturnVal, nil
	}

	err = L.CallByParam(lua.P{Fn: fn, NRet: 1, Protect: true})
	if err != nil {
		L.RaiseError(err.Error())
	}

	err = runtime.cache.StoreVersion(options.Name, options.Version)
	return []lua.LValue{lua.LBool(isDiff)}, nil
}

func checkIfDiff(runtime *Runtime, options cacheDiffOptions) (bool, error) {
	if options.Version == "" {
		options.Version = strconv.Itoa(options.IntVersion)
	}

	currentVersion, err := runtime.cache.GetVersion(options.Name)
	if err != nil {
		return false, err
	}

	isDiff := currentVersion == nil || *currentVersion != options.Version
	slog.With(
		slog.String("version", options.Version),
		slog.Any("current_version", currentVersion),
		slog.String("name", options.Name),
		slog.Bool("is_diff", isDiff),
	).Debug("diffing cache versions")
	return isDiff, nil
}
