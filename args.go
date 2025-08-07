package shmake

import (
	"fmt"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

var _ error = ErrBadType{}

type ErrBadType struct {
	Index int
	Typ   lua.LValueType
}

func (e ErrBadType) Error() string {
	return fmt.Sprintf("bad argument at index %d, expected: %v", e.Index, e.Typ)
}

type ErrBadArg struct {
	Index   int
	Message string
}

func (e ErrBadArg) Error() string {
	return fmt.Sprintf("bad argument at index %d: %s", e.Index, e.Message)
}

func MapTable[T any](l *lua.LState, idx int) (T, error) {
	var t T
	tbl, ok := l.Get(idx).(*lua.LTable)
	if !ok {
		return t, ErrBadType{Index: idx, Typ: lua.LTTable}
	}

	if err := gluamapper.Map(tbl, &t); err != nil {
		return t, ErrBadArg{Index: idx, Message: fmt.Errorf("invalid config: %w", err).Error()}
	}

	return t, nil
}

func MapOptionalTable[T any](l *lua.LState, idx int) (T, error) {
	var t T
	if l.GetTop() < idx {
		return t, nil
	}

	tbl, ok := l.Get(idx).(*lua.LTable)
	if !ok {
		return t, ErrBadType{Index: idx, Typ: lua.LTTable}
	}

	if err := gluamapper.Map(tbl, &t); err != nil {
		return t, ErrBadArg{Index: idx, Message: fmt.Errorf("invalid config: %w", err).Error()}
	}

	return t, nil
}

func MapArray[T any](l *lua.LState, idx int) ([]T, error) {
	lv := l.Get(idx)

	_, ok := lv.(*lua.LTable)
	if !ok {
		return nil, ErrBadType{Index: idx, Typ: lua.LTTable}
	}

	val := gluamapper.ToGoValue(lv, gluamapper.Option{})
	anyv, ok := val.([]any)
	if !ok {
		return nil, ErrBadArg{Index: idx, Message: fmt.Errorf("invalid array, expected array, got %T", val).Error()}
	}

	var arr []T
	for i, v := range anyv {
		if t, ok := v.(T); ok {
			arr = append(arr, t)
		} else {
			return nil, ErrBadArg{
				Index:   idx,
				Message: fmt.Errorf("invalid array, expected %T, got %T at index %d", arr, v, i).Error(),
			}
		}
	}

	return arr, nil
}

func MapString(l *lua.LState, idx int) (string, error) {
	str, ok := l.Get(idx).(lua.LString)
	if !ok {
		return "", ErrBadType{Index: idx, Typ: lua.LTString}
	}

	return string(str), nil
}

func MapFunction(l *lua.LState, idx int) (*lua.LFunction, error) {
	f, ok := l.Get(idx).(*lua.LFunction)
	if !ok {
		return nil, ErrBadType{Index: idx, Typ: lua.LTFunction}
	}

	return f, nil
}
