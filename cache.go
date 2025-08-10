package shmake

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/peterbourgon/diskv/v3"
	"go.starlark.net/starlark"
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

func mapCacheDiffOptions(kwargs []starlark.Tuple) (*cacheDiffOptions, error) {
	options := &cacheDiffOptions{}

	for _, kv := range kwargs {
		key, ok := kv[0].(starlark.String)
		if !ok {
			continue
		}

		switch string(key) {
		case "name":
			if val, ok := kv[1].(starlark.String); ok {
				options.Name = string(val)
			}
		case "version":
			if val, ok := kv[1].(starlark.String); ok {
				options.Version = string(val)
			}
		case "int_version":
			if val, ok := kv[1].(starlark.Int); ok {
				if i, ok := val.Int64(); ok {
					options.IntVersion = int(i)
				}
			}
		}
	}

	if options.Version == "" && options.IntVersion != 0 {
		options.Version = strconv.Itoa(options.IntVersion)
	}

	return options, nil
}

func shmakeDiff(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options, err := mapCacheDiffOptions(kwargs)
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(cache, *options)
	if err != nil {
		return nil, err
	}
	return starlark.Bool(isDiff), nil
}

func shmakeGetVersion(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, fmt.Errorf("get_version() requires exactly 1 argument")
	}

	name, ok := args.Index(0).(starlark.String)
	if !ok {
		return nil, fmt.Errorf("name must be a string")
	}

	v, err := cache.GetVersion(string(name))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return starlark.None, nil
	}

	return starlark.String(*v), nil
}

func shmakeStore(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	options, err := mapCacheDiffOptions(kwargs)
	if err != nil {
		return nil, err
	}

	slog.
		With(slog.String("version", options.Version)).
		With(slog.String("name", options.Name)).
		Debug("storing cache version")

	if err := cache.StoreVersion(options.Name, options.Version); err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func shmakeWithVersion(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, fmt.Errorf("with_version() requires exactly 1 positional argument (the function)")
	}

	fnVal, ok := args.Index(0).(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("first argument must be callable")
	}

	options, err := mapCacheDiffOptions(kwargs)
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(cache, *options)
	if err != nil {
		return nil, err
	}
	if !isDiff {
		return starlark.Bool(false), nil
	}

	res, err := starlark.Call(thread, fnVal, nil, nil)
	if err != nil {
		return nil, err
	}
	slog.With(
		slog.String("name", options.Name),
		slog.Any("response", res),
	).Debug("with_version function returned")

	if err := cache.StoreVersion(options.Name, options.Version); err != nil {
		return nil, err
	}
	return starlark.Bool(true), nil
}

func checkIfDiff(cache Cache, options cacheDiffOptions) (bool, error) {
	if options.Version == "" {
		options.Version = strconv.Itoa(options.IntVersion)
	}

	currentVersion, err := cache.GetVersion(options.Name)
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
