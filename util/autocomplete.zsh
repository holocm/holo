#compdef holo
typeset -A opt_args

(( $+functions[_holo_command] )) || _holo_command()
{
    local -a _commands
    _commands=(
        'apply:Apply available configuration to some or all entities'
        'diff:Diff some or all entities against the last provisioned version'
        'scan:Scan for provisionable entities'
    )
    _describe -t commands 'holo command' _commands
    return 0
}

(( $+functions[_holo_selector] )) || _holo_selector()
{
    _alternative "selectors:Holo selectors:($(holo scan --porcelain | sed -n '/^ENTITY:\|^SOURCE:/ { s/^ENTITY: \|^SOURCE: //; p }' | sort -u))"
    return 0
}

(( $+functions[_holo_zsh_comp] )) || _holo_zsh_comp()
{
    if (( CURRENT == 2 )); then
        _arguments : \
            '--help[Print short usage information.]' \
            '--version[Print a short version string.]' \
            '1::holo command:_holo_command'
    else
        case "$words[2]" in
            apply)
                _arguments : \
                    {-f,--force}'[overwrite manual changes on entities]' \
                    '*:selector:_holo_selector'
                ;;
            diff)
                _holo_selector
                ;;
            scan)
                _arguments : \
                    '(-p --porcelain -s --short)'{-p,--porcelain}'[print raw scan reports]' \
                    '(-p --porcelain -s --short)'{-s,--short}'[print only entity names]' \
                    '*:selector:_holo_selector'
                ;;
        esac
    fi
    return 0
}

_holo_zsh_comp "$@"
