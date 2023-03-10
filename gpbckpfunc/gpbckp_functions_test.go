package gpbckpfunc

import (
	"testing"

	"github.com/woblerr/gpbackup_exporter/gpbckpstruct"
)

type testBackupType struct {
	incr         bool
	dataOnly     bool
	metaDataOnly bool
}

type testObjectFiltering struct {
	includeSchema bool
	includeTable  bool
	excludeSchema bool
	excludeTable  bool
}

func TestGetBackupType(t *testing.T) {
	type args struct {
		backupData gpbckpstruct.BackupConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Full",
			args{
				templateBackupConfig(
					testBackupType{false, false, false},
					testObjectFiltering{false, false, false, false},
				),
			},
			"full",
		},
		{
			"Incremental",
			args{
				templateBackupConfig(
					testBackupType{true, false, false},
					testObjectFiltering{false, false, false, false},
				),
			},
			"incremental",
		},
		{
			"Data-only",
			args{
				templateBackupConfig(
					testBackupType{false, true, false},
					testObjectFiltering{false, false, false, false},
				),
			},
			"data-only",
		},
		{
			"Metadata-only",
			args{
				templateBackupConfig(
					testBackupType{false, false, true},
					testObjectFiltering{false, false, false, false},
				),
			},
			"metadata-only",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBackupType(tt.args.backupData); got != tt.want {
				t.Errorf("\nVariables do not match:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestGetObjectFilteringInfo(t *testing.T) {
	type args struct {
		backupData gpbckpstruct.BackupConfig
	}
	tests := []struct {
		name string
		args args
		want string
	}{

		{
			"NoOptions",
			args{
				templateBackupConfig(
					testBackupType{false, false, false},
					testObjectFiltering{false, false, false, false},
				),
			},
			"",
		},
		{
			"Include-schema",
			args{
				templateBackupConfig(
					testBackupType{false, false, false},
					testObjectFiltering{true, false, false, false},
				),
			},
			"include-schema",
		},
		{
			"Include-table",
			args{
				templateBackupConfig(
					testBackupType{false, false, false},
					testObjectFiltering{false, true, false, false},
				),
			},
			"include-table",
		},
		{
			"Exclude-schema",
			args{
				templateBackupConfig(
					testBackupType{false, false, false},
					testObjectFiltering{false, false, true, false},
				),
			},
			"exclude-schema",
		},
		{
			"Exclude-table",
			args{
				templateBackupConfig(
					testBackupType{false, false, true},
					testObjectFiltering{false, false, false, true},
				),
			},
			"exclude-table",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetObjectFilteringInfo(tt.args.backupData); got != tt.want {
				t.Errorf("\nVariables do not match:\n%s\nwant:\n%s", got, tt.want)
			}
		})
	}
}

func TestGetBackupDuration(t *testing.T) {
	type args struct {
		start string
		end   string
	}
	tests := []struct {
		name    string
		args    args
		want    float64
		wantErr bool
	}{
		{
			"CorrectDuration",
			args{"20230118152654", "20230118152656"},
			2,
			false,
		},
		{
			"BadStart",
			args{"", "20230118152656"},
			0,
			true,
		},
		{
			"BadEnd",
			args{"20230118152654", ""},
			0,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBackupDuration(tt.args.start, tt.args.end)
			if (err != nil) != tt.wantErr {
				t.Errorf("\nVariables do not match:\nerror:\n%v\nwantErr:\n%v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

//nolint:unparam
func templateBackupConfig(bType testBackupType, oFiltering testObjectFiltering) gpbckpstruct.BackupConfig {
	return gpbckpstruct.BackupConfig{
		BackupDir:             "/data/backups",
		BackupVersion:         "1.26.0",
		Compressed:            true,
		CompressionType:       "gzip",
		DatabaseName:          "test",
		DatabaseVersion:       "6.23.0",
		DataOnly:              bType.dataOnly,
		DateDeleted:           "",
		ExcludeRelations:      []string{},
		ExcludeSchemaFiltered: oFiltering.excludeSchema,
		ExcludeSchemas:        []string{},
		ExcludeTableFiltered:  oFiltering.excludeTable,
		IncludeRelations:      []string{},
		IncludeSchemaFiltered: oFiltering.includeSchema,
		IncludeSchemas:        []string{},
		IncludeTableFiltered:  oFiltering.includeTable,
		Incremental:           bType.incr,
		LeafPartitionData:     false,
		MetadataOnly:          bType.metaDataOnly,
		Plugin:                "",
		PluginVersion:         "",
		RestorePlan:           []gpbckpstruct.RestorePlanEntry{},
		SingleDataFile:        false,
		Timestamp:             "20230118152654",
		EndTime:               "20230118152656",
		WithoutGlobals:        false,
		WithStatistics:        false,
		Status:                "Success",
	}
}
