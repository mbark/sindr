# sindr fish completion (uses urfave/cli native completion)

# Function to get completions from urfave/cli
function __sindr_complete
    set -l cmd (commandline -opc)[2..-1]  # Remove 'sindr' from command line
    sindr $cmd --generate-shell-completion 2>/dev/null
end

# Function to get flag completions when current token starts with -
function __sindr_complete_flags
    set -l cmd (commandline -opc)[2..-1]  # Remove 'sindr' from command line
    # Add a dummy '-' to trigger flag completion
    sindr $cmd - --generate-shell-completion 2>/dev/null
end

# Complete commands and subcommands (when not starting with -)
complete -c sindr -f -n "not string match -q -- '-*' (commandline -ct)" -a '(__sindr_complete)'

# Complete flags (when current token starts with -)
complete -c sindr -f -n "string match -q -- '-*' (commandline -ct)" -a '(__sindr_complete_flags)'