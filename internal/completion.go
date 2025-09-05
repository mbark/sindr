package internal

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/urfave/cli/v3"

	"github.com/mbark/sindr/internal/logger"
)

var (
	//go:embed completion/bash.sh
	bashCompletionScript string
	//go:embed completion/zsh.sh
	zshCompletionScript string
	//go:embed completion/fish.sh
	fishCompletionScript string
	//go:embed completion/powershell.ps1
	powershellCompletionScript string
)

// ConfigureShellCompletionCommand creates a completion command for generating shell scripts.
func ConfigureShellCompletionCommand(command *cli.Command) {
	command.Action = func(ctx context.Context, cmd *cli.Command) error {
		args := cmd.Args().Slice()
		if len(args) == 0 {
			return fmt.Errorf("shell type is required (bash, zsh, fish, or powershell)")
		}

		shell := args[0]
		switch shell {
		case "bash":
			logger.Print(bashCompletionScript)
		case "zsh":
			logger.Print(zshCompletionScript)
		case "fish":
			logger.Print(fishCompletionScript)
		case "powershell":
			logger.Print(powershellCompletionScript)
		default:
			return fmt.Errorf(
				"unrecognized shell %s, supported shells are bash, zsh, fish, or powershell",
				shell,
			)
		}
		return nil
	}
}
