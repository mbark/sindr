# sindr PowerShell completion (uses urfave/cli native completion)
Register-ArgumentCompleter -Native -CommandName sindr -ScriptBlock {
    param($commandName, $wordToComplete, $cursorPosition)
    
    # Get the current command line up to the cursor
    $line = $args[2]
    $words = $line -split '\s+'
    
    # Remove 'sindr' from the beginning and get completions
    $completionWords = $words[1..($words.Length-1)]
    
    # Check if we're completing a flag (starts with -)
    if ($wordToComplete.StartsWith('-')) {
        # For flags, add a dummy '-' to trigger flag completion
        $completions = & sindr @completionWords - --generate-shell-completion 2>$null
    } else {
        # For commands and subcommands
        $completions = & sindr @completionWords --generate-shell-completion 2>$null
    }
    
    $completions | ForEach-Object {
        [System.Management.Automation.CompletionResult]::new($_, $_, 'ParameterValue', $_)
    }
}