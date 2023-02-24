#!/usr/bin/env bash

set -e

# Basic command for execute gpbackup_exporter.
EXPORTER_COMMAND="/gpbackup_exporter \
--prom.endpoint=${EXPORTER_ENDPOINT} \
--prom.port=${EXPORTER_PORT} \
--collect.interval=${COLLECT_INTERVAL} \
--gpbackup.history-file=${HISTORY_FILE}"

# Execute the final command.
exec ${EXPORTER_COMMAND}
