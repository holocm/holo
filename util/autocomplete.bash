# bash completion for holo(8)

_holo() {
    local CURRENT_WORD
    CURRENT_WORD="${COMP_WORDS[COMP_CWORD]}"

    if [ "$COMP_CWORD" = 1 ]; then
        # autocomplete first argument (either a command verb or --help/--version)
        COMPREPLY=( $(compgen -W "--help --version apply diff scan selectors" -- "$CURRENT_WORD") )
        return 0
    elif [ "${COMP_WORDS[1]}" = "apply" ]; then
        # autocomplete for "holo apply" - argument is either an entity or -f/--force
        COMPREPLY=( $(compgen -W "$(holo selectors) -f --force" -- "$CURRENT_WORD") )
        return 0
    elif [ "${COMP_WORDS[1]}" = "diff" ]; then
        # autocomplete for "holo diff" - argument is an entity
        COMPREPLY=( $(compgen -W "$(holo selectors)" -- "$CURRENT_WORD") )
        return 0
    elif [ "${COMP_WORDS[1]}" = "scan" ]; then
        # autocomplete for "holo scan" - argument is either an entity or -p/--porcelain/-s/--short
        COMPREPLY=( $(compgen -W "$(holo selectors) -p --porcelain -s --short" -- "$CURRENT_WORD") )
        return 0
    fi
}
complete -F _holo holo
