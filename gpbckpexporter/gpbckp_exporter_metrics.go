package gpbckpexporter

import (
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gpbckpExporterInfoMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_exporter_info",
		Help: "Information about gpbackup exporter.",
	},
		[]string{"version"})
	gpbckpExporterStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_exporter_status",
		Help: "gpbackup exporter get data status.",
	},
		[]string{"database_name"})
)

// Set exporter info metrics:
//   - gpbackup_exporter_info
func getExporterMetrics(exporterVer string, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
	setUpMetric(
		gpbckpExporterInfoMetric,
		"gpbackup_exporter_info",
		1,
		setUpMetricValueFun,
		logger,
		exporterVer,
	)
}

// Set exporter metrics:
//   - gpbackup_exporter_status
func getExporterStatusMetrics(dbStatus dbStatusMap, setUpMetricValueFun setUpMetricValueFunType, logger log.Logger) {
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
