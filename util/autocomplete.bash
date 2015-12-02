#!/bin/bash
_holo() {
    local CURRENT_WORD
    CURRENT_WORD="${COMP_WORDS[COMP_CWORD]}"

    if [ "$COMP_CWORD" = 1 ]; then
        # autocomplete first argument (either a command verb or --help/--version)
        COMPREPLY=( $(compgen -W "--help --version apply diff scan" -- "$CURRENT_WORD") )
        return 0
    elif [ "${COMP_WORDS[1]}" = "apply" ]; then
        # autocomplete for "holo apply" - argument is either an entity or -f/--force
        COMPREPLY=( $(compgen -W "$(holo scan --short) -f --force" -- "$CURRENT_WORD") )
        return 0
    elif [ "${COMP_WORDS[1]}" = "diff" ]; then
        # autocomplete for "holo diff" - argument is an entity
        COMPREPLY=( $(compgen -W "$(holo scan --short)" -- "$CURRENT_WORD") )
        return 0
    elif [ "${COMP_WORDS[1]}" = "scan" ]; then
        # autocomplete for "holo scan" - argument is either an entity or -s/--short
        COMPREPLY=( $(compgen -W "$(holo scan --short) -s --short" -- "$CURRENT_WORD") )
        return 0
    fi
}
complete -F _holo holo
