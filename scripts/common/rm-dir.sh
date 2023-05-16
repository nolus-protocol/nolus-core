#!/bin/bash

# Protects us from accidental directory deletes.

rm_dir() {
    local dir_for_deletion="$1"
    local dir_final=$(basename ${dir_for_deletion})
    local dir_main=$(dirname ${dir_for_deletion})
    local dir_we_delete=$(echo ${dir_final} | sed 's/\///g')

    cd ${dir_main}

    if [[ -d ${dir_we_delete} ]]; then
        rm -rf ${dir_we_delete}
    fi
}
