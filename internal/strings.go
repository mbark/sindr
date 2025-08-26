package internal

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"go.starlark.net/starlark"
)

func SindrString(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	if args.Len() < 1 {
		return nil, fmt.Errorf(
			"string() requires at least 1 positional argument (the template string)",
		)
	}

	tmplVal, ok := args.Index(0).(starlark.String)
	if !ok {
		return nil, fmt.Errorf("first argument must be a string")
	}

	values := make(map[string]any)
	for _, kv := range kwargs {
		key, ok := kv[0].(starlark.String)
		if !ok {
			continue
		}

		goValue, err := toGoValue(kv[1])
		if err != nil {
			return nil, fmt.Errorf("invalid value for key %s: %w", string(key), err)
		}

		values[string(key)] = goValue
	}

	tmplString := strings.TrimSpace(string(tmplVal))
	t := template.Must(template.New("").Parse(tmplString))
	var buf bytes.Buffer
	err := t.Execute(&buf, values)
	if err != nil {
		return nil, fmt.Errorf("execute template: %w", err)
	}

	return starlark.String(buf.String()), nil
}

func toGoValue(value starlark.Value) (any, error) {
	switch val := value.(type) {
	case starlark.String:
		return string(val), nil
	case starlark.Bool:
		return bool(val), nil
	case starlark.Int:
		i, ok := val.Int64()
		if ok {
			return i, nil
		}

		return nil, fmt.Errorf("invalid int value: %s", val.String())

	case *starlark.Dict:
		m := make(map[string]any)
		for k, v := range val.Entries() {
			s, err := cast[starlark.String](k)
			if err != nil {
				return nil, fmt.Errorf("invalid dict key: %w", err)
			}

			goValue, err := toGoValue(v)
			if err != nil {
				return nil, fmt.Errorf("invalid dict value: %w", err)
			}

			m[string(s)] = goValue
		}
		return m, nil

	case *starlark.List:
		var list []any
		for v := range val.Elements() {
			goValue, err := toGoValue(v)
			if err != nil {
				return nil, fmt.Errorf("invalid list value: %w", err)
			}
			list = append(list, goValue)
		}

		return list, nil

	default:
		return nil, fmt.Errorf("type %T is not supported", val)
	}
}
