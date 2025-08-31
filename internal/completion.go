package internal

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// ConfigureShellCompletion sets up custom completion command for sindr.
func ConfigureShellCompletion(sindrCLI *CLI) {
	// Create a hidden helper command for listing available commands
	helperCommand := &cli.Command{
		Name:   "__list-commands",
		Usage:  "Internal helper for completion (hidden)",
		Hidden: true,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			// List all available commands for the current config
			// Only list commands that are not built-in completion commands
			for _, command := range sindrCLI.Command.Command.Commands {
				isBuiltinCommand := command.Name == "__list-commands" ||
					command.Name == "completion" ||
					command.Name == "help"
				if !command.Hidden && !isBuiltinCommand {
					fmt.Println(command.Name)
				}
			}
			return nil
		},
	}

	// Create a custom completion command that generates dynamic completion scripts
	completionCommand := &cli.Command{
		Name:  "completion",
		Usage: "Output shell completion script for bash, zsh, fish, or Powershell",
		Description: `Output shell completion script for bash, zsh, fish, or Powershell.
Source the output to enable completion.

# .bashrc
source <(sindr completion bash)

# .zshrc
source <(sindr completion zsh)

# fish
sindr completion fish > ~/.config/fish/completions/sindr.fish

# Powershell
Output the script to path/to/autocomplete/sindr.ps1 an run it.
`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			args := cmd.Args().Slice()
			if len(args) == 0 {
				return fmt.Errorf("shell type is required (bash, zsh, fish, or powershell)")
			}

			shell := args[0]
			switch shell {
			case "fish":
				fmt.Print(generateFishCompletionScript())
			case "bash":
				fmt.Print(generateBashCompletionScript())
			case "zsh":
				fmt.Print(generateZshCompletionScript())
			case "powershell":
				fmt.Print(generatePowershellCompletionScript())
			default:
				return fmt.Errorf(
					"unsupported shell: %s (supported: bash, zsh, fish, powershell)",
					shell,
				)
			}
			return nil
		},
	}

	// Add both commands
	sindrCLI.Command.Command.Commands = append(
		sindrCLI.Command.Command.Commands,
		helperCommand,
		completionCommand,
	)
}

func generateFishCompletionScript() string {
	return `# sindr dynamic fish shell completion

# Helper function to get available commands dynamically
function __sindr_get_commands
    sindr __list-commands 2>/dev/null
end

# Helper function to check if we're in a subcommand context
function __fish_sindr_no_subcommand
    set -l cmd (commandline -opc)
    set -e cmd[1] # Remove 'sindr'
    
    if test (count $cmd) -eq 0
        return 0
    end
    
    # Get available commands and see if any match
    for command in (__sindr_get_commands)
        if contains -- $command $cmd
            return 1
        end
    end
    return 0
end

# Complete global flags (these are always available)
complete -c sindr -n '__fish_sindr_no_subcommand' -f -l cache-dir -r -d 'path to the Starlark config file'
complete -c sindr -n '__fish_sindr_no_subcommand' -f -l file-name -s f -r -d 'path to the Starlark config file'
complete -c sindr -n '__fish_sindr_no_subcommand' -f -l line-numbers -s l -d 'print logs with Starlark line numbers if possible'
complete -c sindr -n '__fish_sindr_no_subcommand' -f -l no-cache -s n -d 'ignore stored values in the cache'
complete -c sindr -n '__fish_sindr_no_subcommand' -f -l verbose -s v -d 'print logs to stdout'
complete -c sindr -n '__fish_sindr_no_subcommand' -f -l help -s h -d 'show help'

# Dynamic command completion
complete -x -c sindr -n '__fish_sindr_no_subcommand' -a '(__sindr_get_commands)'

# Help completion for all commands
complete -x -c sindr -n 'not __fish_sindr_no_subcommand' -a 'help' -d 'Shows a list of commands or help for one command'
complete -c sindr -f -l help -s h -d 'show help'
`
}

func generateBashCompletionScript() string {
	return `# sindr dynamic bash completion

_sindr_complete() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    
    # If we're completing the first argument (command name)
    if [[ ${#COMP_WORDS[@]} -eq 2 ]]; then
        local commands
        commands=$(sindr __list-commands 2>/dev/null)
        
        if [[ $? -eq 0 ]]; then
            COMPREPLY=($(compgen -W "${commands}" -- "${cur}"))
        else
            # Fallback to file completion if sindr fails
            COMPREPLY=($(compgen -f -- "${cur}"))
        fi
    else
        # For subcommands and flags, use basic completion
        COMPREPLY=($(compgen -W "help --help -h" -- "${cur}"))
    fi
    
    return 0
}

# Register the completion function
complete -F _sindr_complete sindr
`
}

func generateZshCompletionScript() string {
	return `# sindr dynamic zsh completion

_sindr() {
    local context state line
    typeset -A opt_args
    
    # If we're completing the first argument after sindr
    if [[ $CURRENT -eq 2 ]]; then
        local commands
        commands=(${(f)"$(sindr __list-commands 2>/dev/null)"})
        
        if [[ $? -eq 0 ]] && [[ ${#commands[@]} -gt 0 ]]; then
            _describe 'commands' commands
        else
            # Fallback to file completion if sindr fails
            _files
        fi
    else
        # For subcommands, provide basic completion
        local basic_completions
        basic_completions=('help:Shows help for commands')
        _describe 'subcommands' basic_completions
    fi
    
    return 0
}

# Register the completion function
compdef _sindr sindr
`
}

func generatePowershellCompletionScript() string {
	return `# sindr dynamic PowerShell completion

Register-ArgumentCompleter -Native -CommandName sindr -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    
    try {
        # Call sindr to get available commands
        $commands = & sindr __list-commands 2>$null
        
        if ($commands) {
            $commands | Where-Object { $_ -like "$wordToComplete*" } | ForEach-Object {
                [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
            }
        }
    }
    catch {
        # Fallback to empty completions if sindr fails
        @()
    }
}
`
}
