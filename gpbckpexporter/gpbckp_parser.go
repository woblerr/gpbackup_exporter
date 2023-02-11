package gpbckpexporter

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/go-kit/log"
	"gopkg.in/yaml.v3"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/woblerr/gpbackup_exporter/gpbckpfunc"
	"github.com/woblerr/gpbackup_exporter/gpbckpstruct"
)

const emptyLabel = "none"

var hData gpbckpstruct.History

type setUpMetricValueFunType func(metric *prometheus.GaugeVec, value float64, labels ...string) error

func readHistoryFile(filename string) ([]byte, error) {
	data, err := ioutil.ReadFile(filename)
	return data, err
}

func parseResult(output []byte) error {
	err := yaml.Unmarshal(output, &hData)
	return err
}

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

func getBackupMetrics(backupData gpbckpstruct.BackupConfig, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
	var (
		bckpDuration float64
		err          error
	)
	level.Debug(logger).Log(
		"msg", "Metric gpbackup_backup_status",
		"value", getStatusFloat64(backupData.Status),
		"labels",
		strings.Join(
			[]string{
				gpbckpfunc.GetBackupType(backupData),
				backupData.DatabaseName,
				getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
				getEmptyLabel(backupData.Plugin),
				backupData.Timestamp,
			}, ",",
		),
	)
	err = setUpMetricValueFun(
		gpbckpBackupStatusMetric,
		getStatusFloat64(backupData.Status),
		gpbckpfunc.GetBackupType(backupData),
		backupData.DatabaseName,
		getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
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
				gpbckpfunc.GetBackupType(backupData),
				backupData.DatabaseName,
				bckpDateDeleted,
				getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
				getEmptyLabel(backupData.Plugin),
				backupData.Timestamp,
			}, ",",
		),
	)
	err = setUpMetricValueFun(
		gpbckpBackupDataDeletedStatusMetric,
		bckpDeletedStatus,
		gpbckpfunc.GetBackupType(backupData),
		backupData.DatabaseName,
		bckpDateDeleted,
		getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
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
				gpbckpfunc.GetBackupType(backupData),
				getEmptyLabel(backupData.CompressionType),
				backupData.DatabaseName,
				backupData.DatabaseVersion,
				getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
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
		gpbckpfunc.GetBackupType(backupData),
		getEmptyLabel(backupData.CompressionType),
		backupData.DatabaseName,
		backupData.DatabaseVersion,
		getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
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
	bckpDuration, err = gpbckpfunc.GetBackupDuration(backupData.Timestamp, backupData.EndTime)
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
					gpbckpfunc.GetBackupType(backupData),
					backupData.DatabaseName,
					bckpDateDeleted,
					getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
					getEmptyLabel(backupData.Plugin),
					backupData.Timestamp,
				}, ",",
			),
		)
		err = setUpMetricValueFun(
			gpbckpBackupDurationMetric,
			bckpDuration,
			gpbckpfunc.GetBackupType(backupData),
			backupData.DatabaseName,
			getEmptyLabel(gpbckpfunc.GetObjectFilteringInfo(backupData)),
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
// Based on available statuses from gpbackup_manager utility documentation
// (https://github.com/greenplum-db/gpdb/blob/98e79490a26d0d9db4c9239ee1c4b33d8af65ec0/gpdb-doc/dita/utility_guide/ref/gpbackup_manager.xml),
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
	case valueDateDeleted == "In progress":
		dateDeleted = emptyLabel
		deletedStatus = 2
	case valueDateDeleted == "Plugin Backup Delete Failed":
		dateDeleted = emptyLabel
		deletedStatus = 3
	case valueDateDeleted == "Local Delete Failed":
		dateDeleted = emptyLabel
		deletedStatus = 4
	default:
		dateDeleted = valueDateDeleted
		deletedStatus = 1
	}
	return dateDeleted, deletedStatus
}
