# sindr zsh completion (uses urfave/cli native completion)
#compdef sindr

_sindr() {
    local line state

    # Get completions from urfave/cli
    local completions
    if [[ "${words[CURRENT]}" == -* ]]; then
        # For flags, add a dummy '-' to trigger flag completion
        completions=(${(f)"$(sindr ${words[2,CURRENT-1]} - --generate-shell-completion 2>/dev/null)"})
    else
        # For commands and subcommands
        completions=(${(f)"$(sindr ${words[2,CURRENT-1]} --generate-shell-completion 2>/dev/null)"})
    fi
    
    _describe 'commands' completions
}

compdef _sindr sindr