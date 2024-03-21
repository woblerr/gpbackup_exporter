package gpbckpexporter

import (
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/woblerr/gpbackman/gpbckpconfig"
)

var (
	webFlagsConfig web.FlagConfig
	webEndpoint    string
)

// SetPromPortAndPath sets HTTP endpoint parameters
// from command line arguments:
// 'web.endpoint',
// 'web.listen-address',
// 'web.config.file',
// 'web.systemd-socket' (Linux only)
func SetPromPortAndPath(flagsConfig web.FlagConfig, endpoint string) {
	webFlagsConfig = flagsConfig
	webEndpoint = endpoint
}

// StartPromEndpoint run HTTP endpoint
func StartPromEndpoint(logger log.Logger) {
	go func(logger log.Logger) {
		http.Handle(webEndpoint, promhttp.Handler())
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`<html>
			<head><title>gpbackup exporter</title></head>
			<body>
			<h1>gpbackup exporter</h1>
			<p><a href='` + webEndpoint + `'>Metrics</a></p>
			</body>
			</html>`))
		})
		server := &http.Server{
			ReadHeaderTimeout: 5 * time.Second,
		}
		if err := web.ListenAndServe(server, &webFlagsConfig, logger); err != nil {
			level.Error(logger).Log("msg", "Run web endpoint failed", "err", err)
			os.Exit(1)
		}
	}(logger)
}

// GetGPBackupInfo get and parse gpbackup history file
func GetGPBackupInfo(historyFile, backupType string, dbInclude, dbExclude []string, collectDepth int, logger log.Logger) {
	var parseHData gpbckpconfig.History
	// The flag indicates whether it was possible to get data from the gpbackup history.
	// By default, it's set to true.
	getDataSuccessStatus := true
	// To calculate the time elapsed since the last completed backup for specific database.
	// For all databases values are calculated relative to one value.
	currentTime := time.Now()
	currentUnixTime := currentTime.Unix()
	// Calculate metrics collection depth.
	// For backups with timestamp older than this - metrics doesn't collect.
	collectDepthTime := currentTime.AddDate(0, 0, -collectDepth)
	parseHData, err := parseBackupData(historyFile, logger)
	if err != nil {
		level.Error(logger).Log("msg", "Get data failed", "err", err)
		getDataSuccessStatus = false
	}
	// Reset metrics.
	resetMetrics()
	// Loop over each database.
	// If database not set - perform a single loop step to get metrics for all databases.
	for _, db := range dbInclude {
		// Check that database from the include list is not in the exclude list.
		// If database not set - checking for entry into the exclude list will be performed later.
		if dbNotInExclude(db, dbExclude) {
			getExporterStatusMetrics(db, getDataSuccessStatus, setUpMetricValue, logger)
			if len(parseHData.BackupConfigs) != 0 {
				// Like lastbackups["testDB"]["full"] = time
				lastBackups := make(lastBackupMap)
				for i := 0; i < len(parseHData.BackupConfigs); i++ {
					// Check database name and compare with database include filter.
					if db == "" || db == parseHData.BackupConfigs[i].DatabaseName {
						bckpType, err := parseHData.BackupConfigs[i].GetBackupType()
						if err != nil {
							level.Error(logger).Log("msg", "Parse backup type value failed", "err", err)
						}
						// Check backup type and compare with backup type filter.
						if backupType == "" || backupType == bckpType {
							bckpStartTime, err := time.Parse(gpbckpconfig.Layout, parseHData.BackupConfigs[i].Timestamp)
							if err != nil {
								level.Error(logger).Log("msg", "Parse backup timestamp value failed", "err", err)
							}
							bckpStopTime, err := time.Parse(gpbckpconfig.Layout, parseHData.BackupConfigs[i].EndTime)
							if err != nil {
								level.Error(logger).Log("msg", "Parse backup end time value failed", "err", err)
							}
							// Only if set correct value for collectDepth.
							if collectDepth > 0 {
								// gpbackup_history.yml file is sorted by timestamp values.
								// The data of the most recent backup is always located at the beginning of the file.
								// When Unmarshal, we get a sorted list.
								// So as soon as we get the first value that is older than collectDepthTime,
								// the cycle can be braked.
								// If this behavior ever changes, then this code needs to be refactored.
								// See https://github.com/greenplum-db/gpbackup/blame/64c06479043d5a41ce4512ba0549483b71824c2a/history/history.go#L103
								// For results from gpbackup_history.db file a similar sort is performed to reuse this code.
								if collectDepthTime.Before(bckpStartTime) {
									getBackupMetrics(parseHData.BackupConfigs[i], setUpMetricValue, logger)
								} else {
									break
								}
							} else {
								getBackupMetrics(parseHData.BackupConfigs[i], setUpMetricValue, logger)
							}
							if parseHData.BackupConfigs[i].Status == "Success" {
								// Check specific database key already exist.
								if dbLastBackups, ok := lastBackups[parseHData.BackupConfigs[i].DatabaseName]; ok {
									// Check specific backup type key already exist.
									if _, ok := dbLastBackups[bckpType]; !ok {
										dbLastBackups[bckpType] = bckpStopTime
									}
									// A small note on the code above.
									// Since the history file is already sorted, the first occurrence will be the last backup.
									// However, if sorting is suddenly removed in the future, the code should be something like this:
									//	if curLastTime, ok := dbLastBackups[bckpType]; ok {
									//		if curLastTime.Before(bckpStopTime) {
									//			dbLastBackups[bckpType] = bckpStopTime
									//		}
									//	} else {
									//		dbLastBackups[bckpType] = bckpStopTime
									//	}
								} else {
									lastBackups[parseHData.BackupConfigs[i].DatabaseName] = backupMap{bckpType: bckpStopTime}
								}
							}
						}
					}
				}
				if len(lastBackups) != 0 {
					getBackupLastMetrics(lastBackups, currentUnixTime, setUpMetricValue, logger)
				} else {
					level.Warn(logger).Log("msg", "No succeed backups")
				}
			} else {
				level.Warn(logger).Log("msg", "No backup data returned")
			}
		} else {
			// When db is specified in both include and exclude lists, a warning is displayed in the log
			// and data for this db is not collected.
			// It is necessary to set zero metric value for this db.
			getDataSuccessStatus = false
			getExporterStatusMetrics(db, getDataSuccessStatus, setUpMetricValue, logger)
			level.Warn(logger).Log("msg", "DB is specified in include and exclude lists", "DB", db)
		}
	}
}

// GetExporterInfo set exporter info metric
func GetExporterInfo(exporterVersion string, logger log.Logger) {
	getExporterMetrics(exporterVersion, setUpMetricValue, logger)
}
