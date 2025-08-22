package internal

import (
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/joho/godotenv"
	"go.starlark.net/starlark"

	"github.com/mbark/shmake/internal/logger"
)

func ShmakeDotenv(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var list *starlark.List
	var overload bool
	err := starlark.UnpackArgs("dotenv", args, kwargs, "files?", &list, "overload?", &overload)
	if err != nil {
		return nil, err
	}

	files, err := fromList[string](list, func(value starlark.Value) (string, error) {
		s, err := cast[starlark.String](value)
		return string(s), err
	})
	if err != nil {
		return nil, err
	}

	if len(files) == 0 {
		files = []string{".env"}
	}

	logger.Log(
		lipgloss.NewStyle().Bold(true).Render("loading " + strings.Join(files, ", ")),
	)
	envMap, err := godotenv.Read(files...)
	if err != nil {
		return nil, err
	}

	res := loadEnvMap(envMap, overload)
	if len(res.exported) > 0 {
		logger.LogVerbose(
			lipgloss.NewStyle().Bold(true).Faint(true).Render("export"),
			lipgloss.NewStyle().Faint(true).Render(strings.Join(res.exported, " ")),
		)
	}
	if len(res.overloaded) > 0 {
		logger.LogVerbose(
			lipgloss.NewStyle().Bold(true).Faint(true).Render("overload"),
			lipgloss.NewStyle().Faint(true).Render(strings.Join(res.overloaded, " ")),
		)
	}
	if len(res.skipped) > 0 {
		logger.LogVerbose(
			lipgloss.NewStyle().Bold(true).Faint(true).Render("skip"),
			lipgloss.NewStyle().Faint(true).Render(strings.Join(res.skipped, " ")),
		)
	}
	if len(res.exported) == 0 && len(res.overloaded) == 0 && len(res.skipped) == 0 {
		logger.Log(
			lipgloss.NewStyle().Bold(true).Faint(true).Render("no environment exported"),
		)
	}

	return starlark.None, err
}

type loadResult struct {
	overloaded []string
	exported   []string
	skipped    []string
}

// loadEnvMap is based on loadFile from https://github.com/joho/godotenv
func loadEnvMap(envMap map[string]string, overload bool) loadResult {
	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	var res loadResult
	for key, value := range envMap {
		switch {
		case currentEnv[key] && overload:
			res.overloaded = append(res.overloaded, key)
		case !currentEnv[key]:
			res.exported = append(res.exported, key)
		default:
			res.skipped = append(res.skipped, key)
			continue
		}

		_ = os.Setenv(key, value)
	}

	return res
}
