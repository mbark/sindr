# sindr bash completion (uses urfave/cli native completion)
_sindr_completion() {
    local cur prev words cword
    _init_completion || return

    # Use urfave/cli's built-in completion
    local completions
    if [[ "$cur" == -* ]]; then
        # For flags, add a dummy '-' to trigger flag completion
        completions=$(sindr "${words[@]:1:$cword-1}" - --generate-shell-completion 2>/dev/null)
    else
        # For commands and subcommands
        completions=$(sindr "${words[@]:1:$cword-1}" --generate-shell-completion 2>/dev/null)
    fi
    
    COMPREPLY=($(compgen -W "$completions" -- "$cur"))
}

complete -F _sindr_completion sindr