package gpbckpexporter

import (
	"errors"
	"path/filepath"
	"sort"
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

func dbNotInExclude(db string, listExclude []string) bool {
	// Check that exclude list is empty.
	// If so, no excluding databases are set during startup.
	if strings.Join(listExclude, "") != "" {
		for _, val := range listExclude {
			if val == db {
				return false
			}
		}
	}
	return true
}

// Get and parse data from history database:
//   - file with extension .db (sqlite, after gpbackup 1.29.0);
//   - file with extension .yaml (before gpbackup 1.29.0);
//
// Returns parsed data or error.
func parseBackupData(historyFile string, logger log.Logger) (gpbckpconfig.History, error) {
	var parseHData gpbckpconfig.History
	hFileExt := filepath.Ext(historyFile)
	switch hFileExt {
	case ".yaml":
		return getDataFromHistoryFile(historyFile, logger)
	case ".db":
		return getDataFromHistoryDB(historyFile, logger)
	default:
		return parseHData, errors.New("file has an extension other than yaml or db (sqlite)")
	}
}

func getDataFromHistoryFile(historyFile string, logger log.Logger) (gpbckpconfig.History, error) {
	var hData gpbckpconfig.History
	historyData, err := gpbckpconfig.ReadHistoryFile(historyFile)
	if err != nil {
		level.Error(logger).Log("msg", "Read gpbackup history file failed", "err", err)
		return hData, err
	}
	hData, err = gpbckpconfig.ParseResult(historyData)
	if err != nil {
		level.Error(logger).Log("msg", "Parse YAML failed", "err", err)
		return hData, err
	}
	return hData, nil
}

func getDataFromHistoryDB(historyFile string, logger log.Logger) (gpbckpconfig.History, error) {
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
	// Get all backups: active, deleted and failed.
	backupList, err := gpbckpconfig.GetBackupNamesDB(true, true, hDB)
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
	// Sort backups.
	// Since both database formats (yaml and sqlite) are supported simultaneously,
	// it is necessary to sort the result by field Timestamp.
	// Similar to how it is done for yaml format.
	// When switching to sqlite format only, this code will become irrelevant.
	sort.Slice(hData.BackupConfigs, func(i, j int) bool {
		return hData.BackupConfigs[i].Timestamp > hData.BackupConfigs[j].Timestamp
	})
	return hData, nil
}
