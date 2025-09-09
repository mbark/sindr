package internal

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
)

func cast[T any](f starlark.Value) (T, error) {
	v, ok := f.(T)
	if ok {
		return v, nil
	}

	var t T
	if f == nil {
		return t, nil
	}
	return v, fmt.Errorf("expected %T, got %T", t, f)
}

func castString(f starlark.Value) (string, error) {
	v, err := cast[starlark.String](f)
	if err != nil {
		return "", err
	}

	return string(v), nil
}

func castInt(f starlark.Value) (int, error) {
	si, ok := f.(starlark.Int)
	if !ok {
		return 0, fmt.Errorf("expected starlark.Int, got %T", f)
	}

	i, ok := si.Int64()
	if ok {
		return int(i), nil
	}

	return 0, fmt.Errorf("int is not representable as int64: %s", si)
}

func errorAs[T error](err error) (T, bool) {
	var e T
	if errors.As(err, &e) {
		return e, true
	}
	return e, false
}

func toList[T any](l []T, fn func(T) starlark.Value) *starlark.List {
	list := make([]starlark.Value, len(l))
	for i, a := range l {
		list[i] = fn(a)
	}
	return starlark.NewList(list)
}

func fromList[T any](l *starlark.List, fn func(value starlark.Value) (T, error)) ([]T, error) {
	if l == nil {
		return nil, nil
	}

	list := make([]T, l.Len())
	var idx int
	var err error

	var merr error
	for a := range starlark.Elements(l) {
		list[idx], err = fn(a)
		if err != nil {
			merr = errors.Join(merr, fmt.Errorf("element %d: %w", idx, err))
		}
		idx++
	}

	return list, merr
}

func mapList[T, V any](l []T, fn func(T) V) []V {
	v := make([]V, len(l))
	for i, a := range l {
		v[i] = fn(a)
	}
	return v
}

func union[K comparable, V any](m1, m2 map[K]V) map[K]V {
	result := make(map[K]V)
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

func splitKwargs(kwargs []starlark.Tuple, relevant ...string) ([]starlark.Tuple, []starlark.Tuple) {
	keys := make(map[string]bool)
	for _, r := range relevant {
		keys[r] = true
	}

	var relevantKwargs []starlark.Tuple
	var otherKwargs []starlark.Tuple
	for _, kwarg := range kwargs {
		k, ok := kwarg[0].(starlark.String)
		if !ok {
			continue
		}

		if keys[string(k)] {
			relevantKwargs = append(relevantKwargs, kwarg)
		} else {
			otherKwargs = append(otherKwargs, kwarg)
		}
	}

	return relevantKwargs, otherKwargs
}
