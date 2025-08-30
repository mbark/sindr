package internal

import (
	"bytes"
	"errors"
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

	evaled, err := evaluateTemplateString(string(tmplVal), thread, kwargs)
	return starlark.String(evaled), err
}

func evaluateTemplateString(
	tmpl string,
	thread *starlark.Thread,
	kwargs []starlark.Tuple,
) (string, error) {
	values := make(map[string]any)
	var merr error

	addKeyVal := func(key string, val starlark.Value) {
		goValue, err := toGoValue(val)
		if err != nil {
			merr = errors.Join(merr, fmt.Errorf("invalid value for key %s: %w", string(key), err))
			return
		}

		values[key] = goValue
	}

	c := thread.Local("ctx")
	ctx, ok := c.(*Context)
	if ok {
		for k, v := range ctx.Flags.data {
			addKeyVal(k, v)
		}
		for k, v := range ctx.Args.data {
			addKeyVal(k, v)
		}
	}

	for _, kv := range kwargs {
		key, ok := kv[0].(starlark.String)
		if !ok {
			continue
		}
		addKeyVal(string(key), kv[1])
	}
	if merr != nil {
		return "", merr
	}

	tmplString := strings.TrimSpace(tmpl)
	t := template.Must(template.New("").Parse(tmplString))
	var buf bytes.Buffer
	err := t.Execute(&buf, values)
	if err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
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
