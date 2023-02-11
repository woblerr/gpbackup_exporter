package exporter

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gpbckpBackupStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_status",
		Help: "Backup status.",
	},
		[]string{
			"backup_type",
			"database_name",
			"object_filtering",
			"plugin",
			"timestamp"})
	gpbckpkBackupDataDeletedStatusMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_deleted_status",
		Help: "Backup deleted status.",
	},
		[]string{
			"backup_type",
			"database_name",
			"date_deleted",
			"object_filtering",
			"plugin",
			"timestamp"})
	gpbckpBackupInfoMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_info",
		Help: "Backup info.",
	},
		[]string{
			"backup_dir",
			"backup_ver",
			"backup_type",
			"compression_type",
			"database_name",
			"database_ver",
			"object_filtering",
			"plugin",
			"plugin_ver",
			"timestamp",
			"with_statistic"})
	gpbckpBackupDurationMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_backup_duration_seconds",
		Help: "Backup duration.",
	},
		[]string{
			"backup_type",
			"database_name",
			"object_filtering",
			"plugin",
			"timestamp"})
	gpbckpExporterInfoMetric = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "gpbackup_exporter_info",
		Help: "Information about gpbackup exporter.",
	},
		[]string{"version"})
)
