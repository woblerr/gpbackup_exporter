package gpbckpexporter

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gpbckpExporterStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_exporter_status",
		Help: "gpbackup exporter get data status.",
	},
		[]string{"database_name"})
)

// Set exporter metrics:
//   - gpbackup_exporter_status
func getExporterStatusMetrics(dbStatus dbStatusMap, setUpMetricValueFun setUpMetricValueFunType, logger *slog.Logger) {
	for dbName, status := range dbStatus {
		setUpMetric(
			gpbckpExporterStatusMetric,
			"gpbackup_exporter_status",
			convertBoolToFloat64(status),
			setUpMetricValueFun,
			logger,
			dbName,
		)
	}
}

func resetExporterMetrics() {
	gpbckpExporterStatusMetric.Reset()
}
