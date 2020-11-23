#!/usr/bin/env bash

# Provide tab completion for mage
# Copyright https://github.com/yohanyflores
# https://github.com/magefile/mage/issues/113#issuecomment-437710124

_mage_completions()
{
    local cur prev opts

    # namespaces with colon
    _get_comp_words_by_ref -n : cur

    # prev word.
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    case "${prev}" in
        -compile)
            COMPREPLY+=( $(compgen -f -- "${cur}") )
            ;;
        -d)
            COMPREPLY+=( $(compgen -d -- "${cur}") )
            ;;
        -gocmd)
            COMPREPLY+=( $(compgen -f -- "${cur}") )
            ;;
        -t)
            opts="30s 1m 1m30s 2m 2m30s 3m 3m30s 4m 4m30s 5m 10m 20m 30m 1h"
            COMPREPLY+=( $(compgen -W "${opts}" -- ${cur}) )
            ;;
        *)
            if [[ ${cur} == -* ]]; then
                opts="$(mage -h | grep "^\\s*-" | awk '{print $1}')"
            else
                opts="$(mage -l | tail -n +2 | awk '{print tolower($1)}')"
            fi
            COMPREPLY+=( $(compgen -W "${opts}" -- "${cur}"))
            ;;
    esac

    __ltrim_colon_completions "$cur"
}

complete -F _mage_completions mage
