package internal

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"

	"github.com/urfave/cli/v3"
	"go.starlark.net/starlark"
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

	for name := range packageJson.Scripts {
		GlobalCLI.Command.Command.Commands = append(
			GlobalCLI.Command.Command.Commands,
			&cli.Command{
				Name:            name,
				SkipFlagParsing: true,
				Action: func(ctx context.Context, command *cli.Command) error {
					slog.With(slog.String("name", name)).Debug("running command")

					cmdArgs := []string{"run", name}
					if s := command.Args().Slice(); len(s) > 0 {
						cmdArgs = append(cmdArgs, "--")
						cmdArgs = append(cmdArgs, s...)
					}

					cmd := exec.CommandContext(ctx, bin, cmdArgs...)
					output, err := StartShellCmd(cmd, name, true)
					if err != nil {
						return err
					}
					slog.With(slog.Any("output", output), slog.Any("err", err)).
						Debug("command done")
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
