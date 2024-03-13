package main

import (
	"errors"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	kingpin "github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/exporter-toolkit/web/kingpinflag"
	"github.com/woblerr/gpbackup_exporter/gpbckpexporter"
)

var version = "unknown"

func main() {
	var (
		webPath = kingpin.Flag(
			"web.endpoint",
			"Endpoint used for metrics.",
		).Default("/metrics").String()
		webAdditionalToolkitFlags = kingpinflag.AddFlags(kingpin.CommandLine, ":19854")
		collectionInterval        = kingpin.Flag(
			"collect.interval",
			"Collecting metrics interval in seconds.",
		).Default("600").Int()
		collectionDepth = kingpin.Flag(
			"collect.depth",
			"Metrics depth collection in days. Metrics for backup older than this interval will not be collected. 0 - disable.",
		).Default("0").Int()
		gpbckpHistoryFilePath = kingpin.Flag(
			"gpbackup.history-file",
			"Path to gpbackup_history.db or gpbackup_history.yaml.",
		).Default("").String()
		gpbckpIncludeDB = kingpin.Flag(
			"gpbackup.db-include",
			"Specific db for collecting metrics. Can be specified several times.",
		).Default("").PlaceHolder("\"\"").Strings()
		gpbckpExcludeDB = kingpin.Flag(
			"gpbackup.db-exclude",
			"Specific db to exclude from collecting metrics. Can be specified several times.",
		).Default("").PlaceHolder("\"\"").Strings()
		gpbckpBackupType = kingpin.Flag(
			"gpbackup.backup-type",
			"Specific backup type for collecting metrics. One of: [full, incremental, data-only, metadata-only].",
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
	level.Info(logger).Log(
		"mgs", "History database file path",
		"file", *gpbckpHistoryFilePath)
	if *collectionDepth > 0 {
		level.Info(logger).Log(
			"mgs", "Metrics depth collection in days",
			"depth", *collectionDepth)
	}
	if strings.Join(*gpbckpIncludeDB, "") != "" {
		for _, db := range *gpbckpIncludeDB {
			level.Info(logger).Log(
				"mgs", "Collecting metrics for specific DB",
				"DB", db)
		}
	}
	if strings.Join(*gpbckpExcludeDB, "") != "" {
		for _, db := range *gpbckpExcludeDB {
			level.Info(logger).Log(
				"mgs", "Exclude collecting metrics for specific DB",
				"DB", db)
		}
	}
	if *gpbckpBackupType != "" {
		level.Info(logger).Log(
			"mgs", "Collecting metrics for specific backup type",
			"type", *gpbckpBackupType)
	}
	// Setup parameters for exporter.
	gpbckpexporter.SetPromPortAndPath(*webAdditionalToolkitFlags, *webPath)
	level.Info(logger).Log(
		"mgs", "Use exporter parameters",
		"endpoint", *webPath,
		"config.file", *webAdditionalToolkitFlags.WebConfigFile,
	)
	// Check history file type.
	// This code will become not actual after full switch to sqlite history file format
	historyDB, err := checkHistoryFileType(*gpbckpHistoryFilePath)
	if err != nil {
		level.Error(logger).Log("msg", "Check gpbackup history file format failed", "err", err)
		// Immediately exit.
		os.Exit(1)
	}
	// Start exporter.
	gpbckpexporter.StartPromEndpoint(logger)
	// Set up exporter info metric.
	gpbckpexporter.GetExporterInfo(version, logger)
	for {
		// Get information form gpbackup_history.yaml.
		gpbckpexporter.GetGPBackupInfo(
			*gpbckpHistoryFilePath,
			*gpbckpBackupType,
			*gpbckpIncludeDB,
			*gpbckpExcludeDB,
			*collectionDepth,
			historyDB,
			logger,
		)
		// Sleep for 'collection.interval' seconds.
		time.Sleep(time.Duration(*collectionInterval) * time.Second)
	}
}

// Check history db file extension is:
//   - gpbackup_history.db (sqlite, after gpbackup 1.29.0);
//   - gpbackup_history.yaml (before gpbackup 1.29.0);
//
// Returns:
//   - true, if file extension is db;
//   - false, if file extension is yaml.
//   - error, in other cases.
//
// Check only file extension, not content or name.
func checkHistoryFileType(filePath string) (bool, error) {
	hFileExt := filepath.Ext(filePath)
	switch hFileExt {
	case ".db":
		return true, nil
	case ".yaml":
		return false, nil
	default:
		return false, errors.New("file extension is not db or yaml")
	}
}
