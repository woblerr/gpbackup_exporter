package gpbckpexporter

import (
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
)

var (
	promPort          string
	promEndpoint      string
	promTLSConfigPath string
)

// SetPromPortAndPath sets HTTP endpoint parameters
// from command line arguments 'prom.port', 'prom.endpoint' and 'prom.web-config'
func SetPromPortAndPath(port, endpoint, tlsConfigPath string) {
	promPort = port
	promEndpoint = endpoint
	promTLSConfigPath = tlsConfigPath
}

// StartPromEndpoint run HTTP endpoint
func StartPromEndpoint(logger log.Logger) {
	go func(logger log.Logger) {
		http.Handle(promEndpoint, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			<head><title>gpbackup exporter</title></head>
			<body>
			<h1>gpbackup exporter</h1>
			<p><a href='` + promEndpoint + `'>Metrics</a></p>
			</body>
			</html>`))
		})
		server := &http.Server{
			Addr:              ":" + promPort,
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := web.ListenAndServe(server, promTLSConfigPath, logger); err != nil {
			level.Error(logger).Log("msg", "Run web endpoint failed", "err", err)
			os.Exit(1)
		}
	}(logger)
}

// GetGPBackupInfo et and parse gpbackup history file
func GetGPBackupInfo(historyFile string, logger log.Logger) {
	historyData, err := readHistoryFile(historyFile)
	if err != nil {
		level.Error(logger).Log("msg", "Read gpbackup history file failed", "err", err)
	}
	parseHData, err := parseResult(historyData)
	if err != nil {
		level.Error(logger).Log("msg", "Parse YAML failed", "err", err)
	}
	if len(parseHData.BackupConfigs) != 0 {
		for i := 0; i < len(parseHData.BackupConfigs); i++ {
			getBackupMetrics(parseHData.BackupConfigs[i], setUpMetricValue, logger)
		}
	} else {
		level.Warn(logger).Log("msg", "No backup data returned")
	}
}

// GetExporterInfo set exporter info metric
func GetExporterInfo(exporterVersion string, logger log.Logger) {
	getExporterMetrics(exporterVersion, setUpMetricValue, logger)
}

// ResetMetrics reset metrics
func ResetMetrics() {
	gpbckpBackupStatusMetric.Reset()
	gpbckpBackupDataDeletedStatusMetric.Reset()
	gpbckpBackupInfoMetric.Reset()
	gpbckpBackupDurationMetric.Reset()
}
