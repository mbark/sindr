package main

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

func MapTable(idx int, lv lua.LValue, i interface{}) error {
	tbl, ok := lv.(*lua.LTable)
	if !ok {
		return ErrBadType{Index: idx, Typ: lua.LTTable}
	}

	if err := gluamapper.Map(tbl, i); err != nil {
		return ErrBadArg{Index: idx, Message: fmt.Errorf("invalid config: %w", err).Error()}
	}

	return nil
}

func MapArray[T any](idx int, lv lua.LValue) ([]T, error) {
	_, ok := lv.(*lua.LTable)
	if !ok {
		return nil, ErrBadType{Index: idx, Typ: lua.LTTable}
	}

	val := gluamapper.ToGoValue(lv, gluamapper.Option{})
	v, ok := val.([]T)
	if !ok {
		return nil, ErrBadArg{Index: idx, Message: fmt.Errorf("invalid array, expected %T, got %T", v, val).Error()}
	}

	return v, nil
}

func MapString(idx int, lv lua.LValue) (string, error) {
	str, ok := lv.(lua.LString)
	if !ok {
		return "", ErrBadType{Index: idx, Typ: lua.LTString}
	}

	return string(str), nil
}

func MapFunction(idx int, lv lua.LValue) (*lua.LFunction, error) {
	f, ok := lv.(*lua.LFunction)
	if !ok {
		return nil, ErrBadType{Index: idx, Typ: lua.LTFunction}
	}

	return f, nil
}
