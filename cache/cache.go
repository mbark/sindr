package cache

import (
	"fmt"
	"hash/fnv"
	"strconv"

	"github.com/charmbracelet/lipgloss"
	"github.com/peterbourgon/diskv/v3"
	"go.starlark.net/starlark"

	"github.com/mbark/shmake/internal/logger"
)

var GlobalCache diskCache

func SetCache(file string) {
	GlobalCache = NewCache(file)
}

func NewCacheValue(
	_ *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var cacheDir string
	err := starlark.UnpackArgs(fn.Name(), args, kwargs, "cache_dir?", &cacheDir)
	if err != nil {
		return nil, err
	}

	c := &Cache{cacheDir: cacheDir}
	if cacheDir != "" {
		c.diskCache = NewCache(cacheDir)
	} else {
		c.diskCache = GlobalCache // Use global cache
	}

	return c, nil
}

var (
	_ starlark.Value    = (*Cache)(nil)
	_ starlark.HasAttrs = (*Cache)(nil)
)

type Cache struct {
	cacheDir  string
	diskCache diskCache
}

func (c Cache) String() string {
	return "cache"
}

func (c Cache) Type() string {
	return "cache"
}

func (c Cache) Freeze() {
	// Cache is immutable, no-op
}

func (c Cache) Truth() starlark.Bool {
	return starlark.True
}

func (c Cache) Hash() (uint32, error) {
	h := fnv.New32a()
	if c.cacheDir != "" {
		_, _ = h.Write([]byte(c.cacheDir))
	}
	return h.Sum32(), nil
}

func (c Cache) Attr(name string) (starlark.Value, error) {
	switch name {
	case "diff":
		return starlark.NewBuiltin("diff", c.diff), nil
	case "get_version":
		return starlark.NewBuiltin("get_version", c.getVersion), nil
	case "set_version":
		return starlark.NewBuiltin("set_version", c.setVersion), nil
	case "with_version":
		return starlark.NewBuiltin("with_version", c.withVersion), nil
	default:
		return nil, nil
	}
}

func (c Cache) AttrNames() []string {
	return []string{"diff", "get_version", "set_version", "with_version"}
}

// Method wrappers for Cache to expose the shmake functions.
func (c *Cache) diff(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	options, err := unpackCacheOptions(fn, kwargs)
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(c.diskCache, *options)
	if err != nil {
		return nil, err
	}
	return starlark.Bool(isDiff), nil
}

func (c *Cache) getVersion(
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

	v, err := c.diskCache.GetVersion(string(name))
	if err != nil {
		return nil, err
	}
	if v == nil {
		return starlark.None, nil
	}

	return starlark.String(*v), nil
}

var (
	cachePrefixStyle = lipgloss.NewStyle().Faint(true)
	cacheNameStyle   = lipgloss.NewStyle().Bold(true)
)

func (c *Cache) setVersion(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	options, err := unpackCacheOptions(fn, kwargs)
	if err != nil {
		return nil, err
	}

	logger.LogVerbose(
		cachePrefixStyle.Render("cache:"),
		cacheNameStyle.Render(options.name),
		cachePrefixStyle.Render("version=")+cacheNameStyle.Render(options.version),
	)

	if err := c.diskCache.StoreVersion(options.name, options.version); err != nil {
		return nil, err
	}
	return starlark.None, nil
}

func (c *Cache) withVersion(
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

	options, err := unpackCacheOptions(fn, kwargs)
	if err != nil {
		return nil, err
	}

	isDiff, err := checkIfDiff(c.diskCache, *options)
	if err != nil {
		return nil, err
	}
	if !isDiff {
		return starlark.Bool(false), nil
	}

	_, err = starlark.Call(thread, fnVal, nil, nil)
	if err != nil {
		return nil, err
	}

	if err := c.diskCache.StoreVersion(options.name, options.version); err != nil {
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
	if currentVersion == nil {
		logger.LogVerbose(
			cachePrefixStyle.Render("cache:"),
			cacheNameStyle.Render(options.name),
			"current not set",
			cachePrefixStyle.Render("version=")+cacheNameStyle.Render(options.version),
		)
	} else {
		logger.LogVerbose(
			cachePrefixStyle.Render("cache:"),
			cacheNameStyle.Render(options.name),
			cachePrefixStyle.Render("current=")+cacheNameStyle.Render(*currentVersion),
			cachePrefixStyle.Render("version=")+cacheNameStyle.Render(options.version),
		)
	}

	return isDiff, nil
}

type cacheDiffOptions struct {
	name    string
	version string
}

func unpackCacheOptions(
	fn *starlark.Builtin,
	kwargs []starlark.Tuple,
) (*cacheDiffOptions, error) {
	var name string
	var version stringOrInt
	err := starlark.UnpackArgs(fn.Name(), nil, kwargs,
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

// stringOrInt is used to allow the version to be passed in both as a version and an int_version.
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
