package gpbckpexporter

import (
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var gpbckpExporterInfoMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "gpbackup_exporter_info",
	Help: "Information about gpbackup exporter.",
},
	[]string{"version"})

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
