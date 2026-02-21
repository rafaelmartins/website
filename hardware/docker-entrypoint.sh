#!/bin/bash -le

export PATH="/opt/ibom/InteractiveHtmlBom:${PATH}"

if [ x"${GITHUB_ACTIONS}" = xtrue ]; then
    if [ -n "${INPUT_CONFIGURATION}" ]; then
        INPUT_CONFIGURATION="$(realpath "${INPUT_CONFIGURATION}")"
    fi

    if [ -n "${INPUT_DESTINATION}" ]; then
        INPUT_DESTINATION="$(realpath -m "${INPUT_DESTINATION}")"
    fi

    if [ -n "${INPUT_SOURCE}" ]; then
        cd "${INPUT_SOURCE}"
    fi

    exec /bin/website \
        ${INPUT_CONFIGURATION+-c "${INPUT_CONFIGURATION}"} \
        ${INPUT_DESTINATION+-d "${INPUT_DESTINATION}"} \
        -w
fi

exec /bin/website -w "${@}"
