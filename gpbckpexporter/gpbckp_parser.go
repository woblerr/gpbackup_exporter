package gpbckpexporter

import (
	"errors"
	"path/filepath"
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
type dbStatusMap map[string]bool

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

// Reset all metrics.
func resetMetrics() {
	resetBackupMetrics()
	resetLastBackupMetrics()
	resetExporterMetrics()
}

func setUpMetric(metric *prometheus.GaugeVec, metricName string, value float64, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger, labels ...string) {
	level.Debug(logger).Log(
		"msg", "Set up metric",
		"metric", metricName,
		"value", value,
		"labels", strings.Join(labels, ","),
	)
	err := setUpMetricValueFun(metric, value, labels...)
	if err != nil {
		level.Error(logger).Log(
			"msg", "Metric set up failed",
			"metric", metricName,
			"err", err,
		)
	}
}

func dbInList(db string, list []string) bool {
	if listEmpty(list) {
		return false
	}
	for _, val := range list {
		if val == db {
			return true
		}
	}
	return false
}

// Check list not empty.
func listEmpty(list []string) bool {
	return strings.Join(list, "") == ""
}

// Get and parse data from history database:
//   - file with extension .db (sqlite).
//
// Returns parsed data or error.
func parseBackupData(historyFile string, collectDeleted, collectFailed bool, logger log.Logger) (gpbckpconfig.History, error) {
	var parseHData gpbckpconfig.History
	if filepath.Ext(historyFile) != ".db" {
		return parseHData, errors.New("file has an extension other than db (sqlite)")
	}
	return getDataFromHistoryDB(historyFile, collectDeleted, collectFailed, logger)
}

func getDataFromHistoryDB(historyFile string, collectDeleted, collectFailed bool, logger log.Logger) (gpbckpconfig.History, error) {
	var hData gpbckpconfig.History
	hDB, err := gpbckpconfig.OpenHistoryDB(historyFile)
	if err != nil {
		level.Error(logger).Log("msg", "Open gpbackup history db failed", "err", err)
		return hData, err
	}
	defer func() {
		errClose := hDB.Close()
		if errClose != nil {
			level.Error(logger).Log("msg", "Close gpbackup history db failed", "err", errClose)
		}
	}()
	backupList, err := gpbckpconfig.GetBackupNamesDB(collectDeleted, collectFailed, hDB)
	if err != nil {
		level.Error(logger).Log("msg", "Get backups from history db failed", "err", err)
		return hData, err
	}
	// Get data for selected backups.
	for _, backupName := range backupList {
		backupData, err := gpbckpconfig.GetBackupDataDB(backupName, hDB)
		if err != nil {
			level.Error(logger).Log("msg", "Get backup data from history db failed", "err", err)
			return hData, err
		}
		hData.BackupConfigs = append(hData.BackupConfigs, backupData)
	}
	return hData, nil
}
