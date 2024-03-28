#!/usr/bin/env bash

set -e

# Basic command for execute gpbackup_exporter.
EXPORTER_COMMAND="/gpbackup_exporter \
--web.endpoint=${EXPORTER_ENDPOINT} \
--web.listen-address=:${EXPORTER_PORT} \
--web.config.file=${EXPORTER_CONFIG} \
--collect.interval=${COLLECT_INTERVAL} \
--collect.depth=${COLLECT_DEPTH} \
--gpbackup.history-file=${HISTORY_FILE} \
--gpbackup.db-include=${DB_INCLUDE} \
--gpbackup.db-exclude=${DB_EXCLUDE} \
--gpbackup.backup-type=${BACKUP_TYPE}"

# Check variable for enabling collecting metrics for deleted backups.
[ "${COLLECT_DELETED}" == "true" ] &&  EXPORTER_COMMAND="${EXPORTER_COMMAND} --gpbackup.collect-deleted"

# Check variable for enabling collecting metrics for failed backups.
[ "${COLLECT_FAILED}" == "true" ] && EXPORTER_COMMAND="${EXPORTER_COMMAND} --gpbackup.collect-failed"

# Execute the final command.
exec ${EXPORTER_COMMAND}
