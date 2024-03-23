package gpbckpexporter

import (
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var gpbckpBackupSinceLastCompletionSecondsMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "gpbackup_backup_since_last_completion_seconds",
	Help: "Seconds since the last completed backup.",
},
	[]string{
		"backup_type",
		"database_name"})

// Set backup metrics:
//   - gpbackup_backup_since_last_completion_seconds
func getBackupLastMetrics(lastBackups lastBackupMap, currentUnixTime int64, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
	for db, bckps := range lastBackups {
		for bckpType, endTime := range bckps {
			// Seconds since the last completed backups.
			setUpMetric(
				gpbckpBackupSinceLastCompletionSecondsMetric,
				"gpbackup_backup_since_last_completion_seconds",
				time.Unix(currentUnixTime, 0).Sub(endTime).Seconds(),
				setUpMetricValueFun,
				logger,
				bckpType,
				db,
			)
		}
	}
}

func resetLastBackupMetrics() {
	gpbckpBackupSinceLastCompletionSecondsMetric.Reset()
}
