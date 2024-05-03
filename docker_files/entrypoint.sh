#!/usr/bin/env bash

uid=$(id -u)

if [ "${uid}" = "0" ]; then
    # Custom time zone.
    if [ "${TZ}" != "Etc/UTC" ]; then
        cp /usr/share/zoneinfo/${TZ} /etc/localtime
        echo "${TZ}" > /etc/timezone
    fi
fi

if [ "${uid}" = "0" ]; then
    # Use nobody user.
    exec su-exec nobody "$@"
else
    exec "$@"
fi
