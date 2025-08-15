package internal

import (
	"errors"
	"os"
	"path/filepath"

	"go.starlark.net/starlark"
)

// ShmakeNewestTS finds the newest modification time among files matching the given globs.
func ShmakeNewestTS(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return findExtremeTimestamp(args, "newest_ts()", true)
}

// ShmakeOldestTS finds the oldest modification time among files matching the given globs.
func ShmakeOldestTS(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	return findExtremeTimestamp(args, "oldest_ts()", false)
}

// findExtremeTimestamp finds either the newest or oldest timestamp based on the findNewest flag.
func findExtremeTimestamp(args starlark.Tuple, fnName string, findNewest bool) (starlark.Value, error) {
	if args.Len() != 1 {
		return nil, errors.New(fnName + " requires exactly 1 argument (a glob pattern or list of patterns)")
	}

	globs, err := parseGlobArg(args.Index(0))
	if err != nil {
		return nil, err
	}

	result := int64(0)
	found := false

	for _, glob := range globs {
		matches, err := filepath.Glob(glob)
		if err != nil {
			return nil, err
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil {
				continue // skip files that can't be stat'd
			}

			if info.IsDir() {
				continue // skip directories
			}

			modTime := info.ModTime().Unix()
			if !found || (findNewest && modTime > result) || (!findNewest && modTime < result) {
				result = modTime
				found = true
			}
		}
	}

	if !found {
		return nil, errors.New("no files found matching the given patterns")
	}

	return starlark.MakeInt64(result), nil
}

// parseGlobArg parses the argument which can be either a string glob or a list of string globs.
func parseGlobArg(arg starlark.Value) ([]string, error) {
	switch v := arg.(type) {
	case starlark.String:
		return []string{string(v)}, nil
	case *starlark.List:
		globs := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if str, ok := item.(starlark.String); ok {
				globs[i] = string(str)
			} else {
				return nil, errors.New("list items must be strings")
			}
		}
		return globs, nil
	default:
		return nil, errors.New("argument must be a string or list of strings")
	}
}
