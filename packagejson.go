package shmake

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
)

func shmakeLoadPackageJson(
	thread *starlark.Thread,
	fn *starlark.Builtin,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
) (starlark.Value, error) {
	var file string
	err := starlark.UnpackArgs("load_package_json", args, kwargs,
		"file", &file)
	if err != nil {
		return nil, err
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

	for name := range packageJson.Scripts {
		gCLI.Command.Command.Commands = append(gCLI.Command.Command.Commands, &cli.Command{
			Name:            name,
			SkipFlagParsing: true,
			Action: func(ctx context.Context, command *cli.Command) error {
				slog.With(slog.String("name", name)).Debug("running command")

				args := []string{"run", name}
				args = append(args, command.Args().Slice()...)
				cmd := exec.CommandContext(ctx, "npm", args...)
				output, err := startShellCmd(cmd, name)
				if err != nil {
					return err
				}
				slog.With(slog.Any("output", output), slog.Any("err", err)).Debug("command done")
				return nil
			},
		})
	}

	return starlark.None, nil
}

type PackageJson struct {
	Scripts map[string]string `json:"scripts"`
}
