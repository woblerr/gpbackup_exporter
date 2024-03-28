#!/usr/bin/env bash

PORT="${1:-19854}"
EXPORTER_TLS="${2:-false}"
EXPORTER_AUTH="${3:-false}"

# Users for test basic auth.
AUTH_USER="test"
AUTH_PASSWORD="test"

# Use http or https.
case ${EXPORTER_TLS} in
    "false")
        EXPORTER_URL="http://localhost:${PORT}/metrics"
        CURL_FLAGS=""
        ;;
    "true")
        EXPORTER_URL="https://localhost:${PORT}/metrics"
        CURL_FLAGS="-k"
        ;;
    *)
        echo "[ERROR] incorect value: get=${EXPORTER_TLS}, want=true or false"
        exit 1
        ;;
esac

# Use basic auth or not.
case ${EXPORTER_AUTH} in
    "false")
        ;;
    "true")
        CURL_FLAGS+=" -u ${AUTH_USER}:${AUTH_PASSWORD}"
        ;;
    *)
        echo "[ERROR] incorect value: get=${EXPORTER_AUTH}, want=true or false"
        exit 1
        ;;
esac

# A simple test to check the number of metrics.
# Format: regex for metric | repetitions.
declare -a REGEX_LIST=(
    '^gpbackup_backup_deleted_status{.*}|7'
    '^gpbackup_backup_duration_seconds{.*object_filtering="none",plugin="none".*}|5'
    '^gpbackup_backup_duration_seconds{.*object_filtering="include-table",plugin="gpbackup_s3_plugin".*}|2'
    '^gpbackup_backup_duration_seconds{.*}|7'
    '^gpbackup_backup_info{.*database_name="demo",.*,with_statistic="true".*}|1'
    '^gpbackup_backup_info{.*database_name="test",.*,with_statistic="true".*}|1'
    '^gpbackup_backup_info{.*backup_dir="/data/backups".*}|5'
    '^gpbackup_backup_info{.*backup_dir="none".*}|2'
    '^gpbackup_backup_since_last_completion_seconds{.*backup_type="data-only",database_name="test".*}|1'
    '^gpbackup_backup_since_last_completion_seconds{.*backup_type="full",database_name="test".*}|1'
    '^gpbackup_backup_since_last_completion_seconds{.*backup_type="incremental",database_name="test".*}|1'
    '^gpbackup_backup_since_last_completion_seconds{.*backup_type="metadata-only",database_name="test".*}|1'
    '^gpbackup_backup_since_last_completion_seconds{.*backup_type="metadata-only",database_name="demo".*}|1'
    '^gpbackup_backup_status{.*backup_type="data-only",database_name="demo".*} 1$|1'
    '^gpbackup_backup_status{.*} 0$|6'
    '^gpbackup_exporter_status{database_name="test"} 1$|1'
    '^gpbackup_exporter_status{database_name="demo"} 1$|1'
    '^gpbackup_exporter_info{.*}|1' 
)

# Check results.
for i in "${REGEX_LIST[@]}"
do
    regex=$(echo ${i} | cut -f1 -d'|')
    cnt=$(echo ${i} | cut -f2 -d'|')
    metric_cnt=$(curl -s ${CURL_FLAGS} ${EXPORTER_URL} | grep -E "${regex}" | wc -l | tr -d ' ')
    if [[ ${metric_cnt} != ${cnt} ]]; then
        echo "[ERROR] on regex '${regex}': get=${metric_cnt}, want=${cnt}"
        exit 1
    fi
done

echo "[INFO] all tests passed"
exit 0
