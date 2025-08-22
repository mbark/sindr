package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"

	"github.com/mbark/shmake/internal/logger"
)

func ShmakeLoadPackageJson(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var file string
	var bin string
	err := starlark.UnpackArgs("load_package_json", args, kwargs,
		"file", &file,
		"bin?", &bin)
	if err != nil {
		return nil, err
	}

	if bin == "" {
		bin = "npm"
	}

	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var packageJson PackageJson
	err = json.Unmarshal(bs, &packageJson)
	if err != nil {
		return nil, err
	}

	logger := logger.WithStack(thread.CallStack())
	logger.LogVerbose(
		lipgloss.NewStyle().
			Faint(true).
			Bold(true).
			Render(fmt.Sprintf("Importing scripts from %s", file)),
	)
	for name := range packageJson.Scripts {
		logger.LogVerbose(
			lipgloss.NewStyle().Faint(true).Padding(0, 2).Render(fmt.Sprintf("Imported %s", name)),
		)
		GlobalCLI.Command.Command.Commands = append(
			GlobalCLI.Command.Command.Commands,
			&cli.Command{
				Name:            name,
				SkipFlagParsing: true,
				Action: func(ctx context.Context, command *cli.Command) error {
					cmdArgs := []string{"run", name}
					if s := command.Args().Slice(); len(s) > 0 {
						cmdArgs = append(cmdArgs, "--")
						cmdArgs = append(cmdArgs, s...)
					}

					cmd := exec.CommandContext(ctx, bin, cmdArgs...)
					logger.Log(commandStyle.Render(cmd.String()))
					_, err := StartShellCmd(logger, cmd, "", true)
					if err != nil {
						return err
					}
					return nil
				},
			},
		)
	}

	return starlark.None, nil
}

type PackageJson struct {
	Scripts map[string]string `json:"scripts"`
}
