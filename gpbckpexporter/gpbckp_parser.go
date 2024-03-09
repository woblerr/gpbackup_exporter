package gpbckpexporter

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/log"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/woblerr/gpbackman/gpbckpconfig"
)

const emptyLabel = "none"

type setUpMetricValueFunType func(metric *prometheus.GaugeVec, value float64, labels ...string) error

type backupMap map[string]time.Time
type lastBackupMap map[string]backupMap

func getExporterMetrics(exporterVer string, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
	level.Debug(logger).Log(
		"msg", "Metric gpbackup_exporter_info",
		"value", 1,
		"labels", exporterVer,
	)
	err := setUpMetricValueFun(
		gpbckpExporterInfoMetric,
		1,
		exporterVer,
	)
	if err != nil {
		level.Error(logger).Log(
			"msg", "Metric gpbackup_exporter_info set up failed",
			"err", err,
		)
	}
}

func setUpMetricValue(metric *prometheus.GaugeVec, value float64, labels ...string) error {
	metricVec, err := metric.GetMetricWithLabelValues(labels...)
	if err != nil {
		return err
	}
	// The situation should be handled by the prometheus libraries.
	// But, anything is possible.
	if metricVec == nil {
		err := errors.New("metric is nil")
		return err
	}
	metricVec.Set(value)
	return nil
}

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
	level.Debug(logger).Log(
		"msg", "Metric gpbackup_backup_status",
		"value", getStatusFloat64(backupData.Status),
		"labels",
		strings.Join(
			[]string{
				bckpType,
				backupData.DatabaseName,
				getEmptyLabel(backpObjectFiltering),
				getEmptyLabel(backupData.Plugin),
				backupData.Timestamp,
			}, ",",
		),
	)
	err = setUpMetricValueFun(
		gpbckpBackupStatusMetric,
		getStatusFloat64(backupData.Status),
		bckpType,
		backupData.DatabaseName,
		getEmptyLabel(backpObjectFiltering),
		getEmptyLabel(backupData.Plugin),
		backupData.Timestamp,
	)
	if err != nil {
		level.Error(logger).Log(
			"msg", "Metric gpbackup_backup_status set up failed",
			"err", err,
		)
	}
	bckpDateDeleted, bckpDeletedStatus := getDeletedStatusCode(backupData.DateDeleted)
	level.Debug(logger).Log(
		"msg", "Metric gpbackup_backup_deleted_status",
		"value", bckpDeletedStatus,
		"labels",
		strings.Join(
			[]string{
				bckpType,
				backupData.DatabaseName,
				bckpDateDeleted,
				getEmptyLabel(backpObjectFiltering),
				getEmptyLabel(backupData.Plugin),
				backupData.Timestamp,
			}, ",",
		),
	)
	err = setUpMetricValueFun(
		gpbckpBackupDataDeletedStatusMetric,
		bckpDeletedStatus,
		bckpType,
		backupData.DatabaseName,
		bckpDateDeleted,
		getEmptyLabel(backpObjectFiltering),
		getEmptyLabel(backupData.Plugin),
		backupData.Timestamp,
	)
	if err != nil {
		level.Error(logger).Log(
			"msg", "Metric gpbackup_backup_deleted_status set up failed",
			"err", err,
		)
	}
	level.Debug(logger).Log(
		"msg", "Metric gpbackup_backup_info",
		"value", 1,
		"labels",
		strings.Join(
			[]string{
				getEmptyLabel(backupData.BackupDir),
				backupData.BackupVersion,
				bckpType,
				getEmptyLabel(backupData.CompressionType),
				backupData.DatabaseName,
				backupData.DatabaseVersion,
				getEmptyLabel(backpObjectFiltering),
				getEmptyLabel(backupData.Plugin),
				getEmptyLabel(backupData.PluginVersion),
				backupData.Timestamp,
				strconv.FormatBool(backupData.WithStatistics),
			}, ",",
		),
	)
	err = setUpMetricValueFun(
		gpbckpBackupInfoMetric,
		1,
		getEmptyLabel(backupData.BackupDir),
		backupData.BackupVersion,
		bckpType,
		getEmptyLabel(backupData.CompressionType),
		backupData.DatabaseName,
		backupData.DatabaseVersion,
		getEmptyLabel(backpObjectFiltering),
		getEmptyLabel(backupData.Plugin),
		getEmptyLabel(backupData.PluginVersion),
		backupData.Timestamp,
		strconv.FormatBool(backupData.WithStatistics),
	)
	if err != nil {
		level.Error(logger).Log(
			"msg", "Metric gpbackup_backup_info set up failed",
			"err", err,
		)
	}

	bckpDuration, err = backupData.GetBackupDuration()
	if err != nil {
		level.Error(logger).Log(
			"msg", "Failed to parse dates to calculate duration",
			"err", err,
		)
	} else {
		level.Debug(logger).Log(
			"msg", "Metric gpbackup_backup_duration_seconds",
			"value", bckpDuration,
			"labels",
			strings.Join(
				[]string{
					bckpType,
					backupData.DatabaseName,
					bckpDateDeleted,
					getEmptyLabel(backpObjectFiltering),
					getEmptyLabel(backupData.Plugin),
					backupData.Timestamp,
				}, ",",
			),
		)
		err = setUpMetricValueFun(
			gpbckpBackupDurationMetric,
			bckpDuration,
			bckpType,
			backupData.DatabaseName,
			getEmptyLabel(backpObjectFiltering),
			getEmptyLabel(backupData.Plugin),
			backupData.Timestamp,
		)
		if err != nil {
			level.Error(logger).Log(
				"msg", "Metric gpbackup_backup_info set up failed",
				"err", err,
			)
		}
	}
}

func getBackupLastMetrics(lastBackups lastBackupMap, currentUnixTime int64, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
	for db, bckps := range lastBackups {
		for bckpType, endTime := range bckps {
			level.Debug(logger).Log(
				"msg", "Metric gpbackup_backup_since_last_completion_seconds",
				"value", time.Unix(currentUnixTime, 0).Sub(endTime).Seconds(),
				"labels",
				strings.Join(
					[]string{
						bckpType,
						db,
					}, ",",
				),
			)
			err := setUpMetricValueFun(
				gpbckpBackupSinceLastCompletionSecondsMetric,
				time.Unix(currentUnixTime, 0).Sub(endTime).Seconds(),
				bckpType,
				db,
			)
			if err != nil {
				level.Error(logger).Log(
					"msg", "Metric gpbackup_backup_since_last_completion_seconds set up failed",
					"err", err,
				)
			}
		}
	}
}

// Convert backup status to float64.
func getStatusFloat64(valueStatus string) float64 {
	if valueStatus == "Failure" {
		return 1
	}
	return 0
}

func getEmptyLabel(str string) string {
	if str == "" {
		return emptyLabel
	}
	return str
}

// Get status code about backup deletion status.
// Based on available statuses from gpbackman utility documentation,
// but not limited to that.
//   - 0 - backup still exists;
//   - 1 - backup was successfully deleted;
//   - 2 - the deletion is in progress;
//   - 3 - last delete attempt failed to delete backup from plugin storage;
//   - 4 - last delete attempt failed to delete backup from local storage;
func getDeletedStatusCode(valueDateDeleted string) (string, float64) {
	var (
		dateDeleted   string
		deletedStatus float64
	)
	switch {
	case valueDateDeleted == "":
		dateDeleted = emptyLabel
		deletedStatus = 0
	case valueDateDeleted == gpbckpconfig.DateDeletedInProgress:
		dateDeleted = emptyLabel
		deletedStatus = 2
	case valueDateDeleted == gpbckpconfig.DateDeletedPluginFailed:
		dateDeleted = emptyLabel
		deletedStatus = 3
	case valueDateDeleted == gpbckpconfig.DateDeletedLocalFailed:
		dateDeleted = emptyLabel
		deletedStatus = 4
	default:
		dateDeleted = valueDateDeleted
		deletedStatus = 1
	}
	return dateDeleted, deletedStatus
}
