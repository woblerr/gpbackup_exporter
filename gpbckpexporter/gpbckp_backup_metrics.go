package gpbckpexporter

import (
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/woblerr/gpbackman/gpbckpconfig"
)

var (
	gpbckpBackupStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_status",
		Help: "Backup status.",
	},
		[]string{
			"backup_type",
			"database_name",
			"object_filtering",
			"plugin",
			"timestamp"})
	gpbckpBackupDataDeletedStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_deletion_status",
		Help: "Backup deletion status.",
	},
		[]string{
			"backup_type",
			"database_name",
			"date_deleted",
			"object_filtering",
			"plugin",
			"timestamp"})
	gpbckpBackupInfoMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_info",
		Help: "Backup info.",
	},
		[]string{
			"backup_dir",
			"backup_ver",
			"backup_type",
			"compression_type",
			"database_name",
			"database_ver",
			"object_filtering",
			"plugin",
			"plugin_ver",
			"timestamp",
			"with_statistic"})
	gpbckpBackupDurationMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_duration_seconds",
		Help: "Backup duration.",
	},
		[]string{
			"backup_type",
			"database_name",
			"end_time",
			"object_filtering",
			"plugin",
			"timestamp"})
)

// Set backup metrics:
//   - gpbackup_backup_status
//   - gpbackup_backup_deletion_status
//   - gpbackup_backup_info
//   - gpbackup_backup_duration_seconds
func getBackupMetrics(backupData gpbckpconfig.BackupConfig, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
	var (
		bckpDuration float64
		err          error
	)
	bckpType, err := backupData.GetBackupType()
	if err != nil {
		level.Error(logger).Log("msg", "Parse backup type value failed", "err", err)
	}
	backpObjectFiltering, err := backupData.GetObjectFilteringInfo()
	if err != nil {
		level.Error(logger).Log("msg", "Parse object filtering value failed", "err", err)
	}
	bckpDuration, err = backupData.GetBackupDuration()
	if err != nil {
		level.Error(logger).Log(
			"msg", "Failed to parse dates to calculate duration",
			"err", err,
		)
	}
	bckpDateDeleted, bckpDeletedStatus := getDeletedStatusCode(backupData.DateDeleted)
	// Backup status.
	setUpMetric(
		gpbckpBackupStatusMetric,
		"gpbackup_backup_status",
		convertStatusFloat64(backupData.Status),
		setUpMetricValueFun,
		logger,
		bckpType,
		backupData.DatabaseName,
		convertEmptyLabel(backpObjectFiltering),
		convertEmptyLabel(backupData.Plugin),
		backupData.Timestamp,
	)
	// Backup deletion status.
	setUpMetric(
		gpbckpBackupDataDeletedStatusMetric,
		"gpbackup_backup_deletion_status",
		bckpDeletedStatus,
		setUpMetricValueFun,
		logger,
		bckpType,
		backupData.DatabaseName,
		bckpDateDeleted,
		convertEmptyLabel(backpObjectFiltering),
		convertEmptyLabel(backupData.Plugin),
		backupData.Timestamp,
	)
	// Backup info.
	setUpMetric(
		gpbckpBackupInfoMetric,
		"gpbackup_backup_info",
		1,
		setUpMetricValueFun,
		logger,
		convertEmptyLabel(backupData.BackupDir),
		backupData.BackupVersion,
		bckpType,
		convertEmptyLabel(backupData.CompressionType),
		backupData.DatabaseName,
		backupData.DatabaseVersion,
		convertEmptyLabel(backpObjectFiltering),
		convertEmptyLabel(backupData.Plugin),
		convertEmptyLabel(backupData.PluginVersion),
		backupData.Timestamp,
		strconv.FormatBool(backupData.WithStatistics),
	)
	// Backup duration.
	setUpMetric(
		gpbckpBackupDurationMetric,
		"gpbackup_backup_duration_seconds",
		bckpDuration,
		setUpMetricValueFun,
		logger,
		bckpType,
		backupData.DatabaseName,
		// End time may be not set, if backup in progress.
		convertEmptyLabel(backupData.EndTime),
		convertEmptyLabel(backpObjectFiltering),
		convertEmptyLabel(backupData.Plugin),
		backupData.Timestamp,
	)
}

func resetBackupMetrics() {
	gpbckpBackupStatusMetric.Reset()
	gpbckpBackupDataDeletedStatusMetric.Reset()
	gpbckpBackupInfoMetric.Reset()
	gpbckpBackupDurationMetric.Reset()
}
