package internal

import (
	"errors"
	"fmt"

	"go.starlark.net/starlark"
)

func cast[T any](f any) (T, error) {
	v, ok := f.(T)
	if ok {
		return v, nil
	}

	var t T
	return v, fmt.Errorf("expected %T, got %T", t, f)
}

func castString[T ~string](f any) (string, error) {
	v, ok := f.(T)
	if ok {
		return string(v), nil
	}

	var t T
	return "", fmt.Errorf("expected %T, got %T", t, f)
}

func castInt(f any) (int, error) {
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

func mapList[T, V any](l []T, fn func(T) V) []V {
	v := make([]V, len(l))
	for i, a := range l {
		v[i] = fn(a)
	}
	return v
}
