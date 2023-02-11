package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	gpbckp "github.com/woblerr/gpbackup_exporter/exporter"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var version = "unknown"

func main() {
	var (
		promPort = kingpin.Flag(
			"prom.port",
			"Port for prometheus metrics to listen on.",
		).Default("19854").String()
		promPath = kingpin.Flag(
			"prom.endpoint",
			"Endpoint used for metrics.",
		).Default("/metrics").String()
		promTLSConfigFile = kingpin.Flag(
			"prom.web-config",
			"[EXPERIMENTAL] Path to config yaml file that can enable TLS or authentication.",
		).Default("").String()
		collectionInterval = kingpin.Flag(
			"collect.interval",
			"Collecting metrics interval in seconds.",
		).Default("600").Int()
		gpbckpHistoryFilePath = kingpin.Flag(
			"gpbackup.history-file",
			"Path to gpbackup_history.yaml",
		).Default("").String()
	)
	// Set logger config.
	promlogConfig := &promlog.Config{}
	// Add flags log.level and log.format from promlog package.
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	// Add short help flag.
	kingpin.HelpFlag.Short('h')
	// Load command line arguments.
	kingpin.Parse()
	// Setup signal catching.
	sigs := make(chan os.Signal, 1)
	// Catch  listed signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	// Set logger.
	logger := promlog.New(promlogConfig)
	// Method invoked upon seeing signal.
	go func(logger log.Logger) {
		s := <-sigs
		level.Warn(logger).Log(
			"msg", "Stopping exporter",
			"name", filepath.Base(os.Args[0]),
			"signal", s)
		os.Exit(1)
	}(logger)
	level.Info(logger).Log(
		"msg", "Starting exporter",
		"name", filepath.Base(os.Args[0]),
		"version", version)
	// Setup parameters for exporter.
	gpbckp.SetPromPortAndPath(*promPort, *promPath, *promTLSConfigFile)
	level.Info(logger).Log(
		"mgs", "Use port and HTTP endpoint",
		"port", *promPort,
		"endpoint", *promPath,
		"web-config", *promTLSConfigFile,
	)
	// Start exporter.
	gpbckp.StartPromEndpoint(logger)
	// Set up exporter info metric.
	gpbckp.GetExporterInfo(version, logger)
	for {
		// Reset metrics.
		gpbckp.ResetMetrics()
		// Get information form gpbackup_history.yaml.
		gpbckp.GetGPBackupInfo(*gpbckpHistoryFilePath, logger)
		// Sleep for 'collection.interval' seconds.
		time.Sleep(time.Duration(*collectionInterval) * time.Second)
	}
}
