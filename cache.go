package shmake

import (
	"fmt"
	"hash/fnv"
	"log/slog"
	"strconv"

	"github.com/peterbourgon/diskv/v3"
	"go.starlark.net/starlark"
)

func shmakeDiff(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	options, err := unpackCacheOptions(fn, args, kwargs)
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(cache, *options)
	if err != nil {
		return nil, err
	}
	return starlark.Bool(isDiff), nil
}

func shmakeGetVersion(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
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

func shmakeStore(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	options, err := unpackCacheOptions(fn, args, kwargs)
	if err != nil {
		return nil, err
	}

	slog.
		With(slog.String("version", options.version)).
		With(slog.String("name", options.name)).
		Debug("storing cache version")

	if err := cache.StoreVersion(options.name, options.version); err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func shmakeWithVersion(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, fmt.Errorf(
			"with_version() requires exactly 1 positional argument (the function)",
		)
	}

	fnVal, ok := args.Index(0).(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("first argument must be callable")
	}

	options, err := unpackCacheOptions(fn, args, kwargs)
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
		slog.String("name", options.name),
		slog.Any("response", res),
	).Debug("with_version function returned")

	if err := cache.StoreVersion(options.name, options.version); err != nil {
		return nil, err
	}
	return starlark.Bool(true), nil
}

func checkIfDiff(cache diskCache, options cacheDiffOptions) (bool, error) {
	currentVersion, err := cache.GetVersion(options.name)
	if err != nil {
		return false, err
	}

	isDiff := currentVersion == nil || *currentVersion != options.version
	slog.With(
		slog.String("version", options.version),
		slog.Any("current_version", currentVersion),
		slog.String("name", options.name),
		slog.Bool("is_diff", isDiff),
	).Debug("diffing cache versions")
	return isDiff, nil
}

type cacheDiffOptions struct {
	name    string
	version string
}

func unpackCacheOptions(fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (*cacheDiffOptions, error) {
	var name string
	var version stringOrInt
	err := starlark.UnpackArgs(fn.Name(), args, kwargs,
		"name", &name,
		"version", &version,
	)
	if err != nil {
		return nil, err
	}
	return &cacheDiffOptions{
		name:    name,
		version: version.String(),
	}, nil
}

type diskCache struct {
	diskv          *diskv.Diskv
	ForceOutOfDate bool // ForceOutOfDate makes all gets return nil
}

func NewCache(file string) diskCache {
	return diskCache{
		diskv: diskv.New(diskv.Options{
			BasePath:     file,
			Transform:    func(s string) []string { return []string{} },
			CacheSizeMax: 1024 * 1024,
		}),
	}
}

func (c diskCache) StoreVersion(name, version string) error {
	return c.diskv.Write(name, []byte(version))
}

func (c diskCache) GetVersion(name string) (*string, error) {
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

var (
	_ starlark.Unpacker = new(stringOrInt)
	_ starlark.Value    = new(stringOrInt)
)

type stringOrInt struct {
	s      *string
	i      *int
	frozen bool
}

func (si *stringOrInt) String() string {
	if si == nil {
		return ""
	}
	if si.s != nil {
		return *si.s
	}
	if si.i != nil {
		return strconv.Itoa(*si.i)
	}

	return ""
}

func (si *stringOrInt) Type() string {
	return "stringOrInt"
}

func (si *stringOrInt) Freeze() {
	si.frozen = true
}

func (si *stringOrInt) Truth() starlark.Bool {
	return starlark.True
}

func (si *stringOrInt) Hash() (uint32, error) {
	h := fnv.New32a()
	_, _ = h.Write([]byte(si.String()))
	return h.Sum32(), nil
}

func (si *stringOrInt) Unpack(v starlark.Value) error {
	switch v := v.(type) {
	case starlark.String:
		s := string(v)
		si.s = &s
		return nil

	case starlark.Int:
		i, ok := v.Int64()
		if !ok {
			return fmt.Errorf("integer not int: %v", v)
		}
		ii := int(i)
		si.i = &ii
		return nil

	default:
		return fmt.Errorf("unsupported type: %T", v)
	}
}
