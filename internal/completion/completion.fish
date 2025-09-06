function __fish_sindr_complete
    set -l tokens (commandline -opc)
    set -l curtok (commandline -ct)
    sindr __complete -- $tokens $curtok
end

complete -c sindr -f -a '(__fish_sindr_complete)'
