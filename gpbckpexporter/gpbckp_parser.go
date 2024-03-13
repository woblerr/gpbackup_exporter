package gpbckpexporter

import (
	"errors"
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

// Reset all metrics.
func resetMetrics() {
	resetBackupMetrics()
	resetLastBackupMetrics()
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
