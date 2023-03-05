#!/usr/bin/env bash

set -e

# Basic command for execute gpbackup_exporter.
EXPORTER_COMMAND="/gpbackup_exporter \
--prom.endpoint=${EXPORTER_ENDPOINT} \
--prom.port=${EXPORTER_PORT} \
--prom.web-config=${EXPORTER_CONFIG} \
--collect.interval=${COLLECT_INTERVAL} \
--collect.depth=${COLLECT_DEPTH} \
--gpbackup.history-file=${HISTORY_FILE} \
--gpbackup.db-include=${DB_INCLUDE} \
--gpbackup.db-exclude=${DB_EXCLUDE} \
--gpbackup.backup-type=${BACKUP_TYPE}"

# Execute the final command.
exec ${EXPORTER_COMMAND}
