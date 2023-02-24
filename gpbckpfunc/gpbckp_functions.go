package gpbckpfunc

import (
	"time"

	"github.com/woblerr/gpbackup_exporter/gpbckpstruct"
)

const layout = "20060102150405"

// GetBackupType Get backup type.
// The "backup_type" label value is calculated, based on:
//   - full - contains user data, all global and local metadata for the database;
//   - incremental – contains user data, all global and local metadata changed since a previous full backup;
//   - metadata-only – contains only global and local metadata for the database;
//   - data-only – contains only user data from the database.
func GetBackupType(backupData gpbckpstruct.BackupConfig) string {
	var backupType string
	// For gpbackup you cannot combine --data-only or --metadata-only with --incremental (see docs).
	// So these flags cannot be set at the same time.
	switch {
	case backupData.Incremental:
		backupType = "incremental"
	case backupData.DataOnly:
		backupType = "data-only"
	case backupData.MetadataOnly:
		backupType = "metadata-only"
	default:
		backupType = "full"
	}
	return backupType
}

// GetObjectFilteringInfo Get object filtering information.
// The "object_filtering" label value is calculated, base on
// on whether at least one of the flags was specified:
//   - include-schema – at least one "--include-schema" option was specified;
//   - exclude-schema – at least one "--exclude-schema" option was specified;
//   - include-table – at least one "--include-table" option was specified;
//   - exclude-table – at least one "--exclude-table" option was specified;
//   - "" - no options was specified.
func GetObjectFilteringInfo(backupData gpbckpstruct.BackupConfig) string {
	var objectFiltering string
	switch {
	case backupData.IncludeSchemaFiltered:
		objectFiltering = "include-schema"
	case backupData.ExcludeSchemaFiltered:
		objectFiltering = "exclude-schema"
	case backupData.IncludeTableFiltered:
		objectFiltering = "include-table"
	case backupData.ExcludeTableFiltered:
		objectFiltering = "exclude-table"
	default:
		objectFiltering = ""
	}
	return objectFiltering
}

// GetBackupDuration Get backup duration.
func GetBackupDuration(start, end string) (float64, error) {
	var (
		zeroDuration       float64 = 0
		startTime, endTime time.Time
		err                error
	)
	startTime, err = time.Parse(layout, start)
	if err != nil {
		return zeroDuration, err
	}
	endTime, err = time.Parse(layout, end)
	if err != nil {
		return zeroDuration, err
	}
	return endTime.Sub(startTime).Seconds(), nil
}
