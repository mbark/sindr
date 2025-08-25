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

		switch val := kv[1].(type) {
		case starlark.String:
			values[string(key)] = string(val)
		case starlark.Bool:
			values[string(key)] = bool(val)
		case starlark.Int:
			i, ok := val.Int64()
			if ok {
				values[string(key)] = i
			}
		}
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
