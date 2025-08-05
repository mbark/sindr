package main

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/yuin/gluamapper"
	lua "github.com/yuin/gopher-lua"
)

func getStringModule() Module {
	return Module{
		exports: map[string]ModuleFunction{
			"template": templateString,
		},
	}
}

func templateString(_ *Runtime, L *lua.LState) ([]lua.LValue, error) {
	str, err := MapString(1, L.Get(1))
	if err != nil {
		return nil, err
	}

	values := make(map[string]any)
	if L.GetTop() > 1 {
		tbl, ok := L.Get(2).(*lua.LTable)
		if !ok {
			return nil, ErrBadType{Index: 2, Typ: lua.LTTable}
		}

		if err := gluamapper.NewMapper(gluamapper.Option{
			NameFunc: func(field string) string { return field },
		}).Map(tbl, &values); err != nil {
			return nil, ErrBadArg{Index: 2, Message: fmt.Errorf("invalid option: %w", err).Error()}
		}
	}
	fmt.Println("values", values)

	globals := L.Get(lua.GlobalsIndex)
	tbl, ok := globals.(*lua.LTable)
	if !ok {
		return nil, fmt.Errorf("globals is not a table")
	}

	tbl.ForEach(func(key lua.LValue, value lua.LValue) {
		skey, ok := key.(lua.LString)
		if !ok {
			return
		}
		var val any
		switch v := value.(type) {
		case lua.LString:
			val = string(v)
		case lua.LNumber:
			val = float64(v)
		case lua.LBool:
			val = bool(v)
		default:
			return
		}
		values[string(skey)] = val
	})

	t := template.Must(template.New("").Parse(str))
	var buf bytes.Buffer
	err = t.Execute(&buf, values)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return []lua.LValue{lua.LString(buf.String())}, nil
}
