package gpbckpexporter

import (
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/promlog"
	"github.com/woblerr/gpbackman/gpbckpconfig"
)

func TestGetDeletedStatusCode(t *testing.T) {
	type args struct {
		valueDateDeleted string
	}
	tests := []struct {
		name  string
		args  args
		want  string
		want1 float64
	}{
		{
			"Exist",
			args{""},
			"none",
			0,
		},
		{
			"InProgress",
			args{"In progress"},
			"none",
			2,
		},
		{
			"PluginDeleteFailed",
			args{"Plugin Backup Delete Failed"},
			"none",
			3,
		},
		{
			"LocalDeleteFailed",
			args{"Local Delete Failed"},
			"none",
			4,
		},
		{
			"ValidDate",
			args{"20230118150331"},
			"20230118150331",
			1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := getDeletedStatusCode(tt.args.valueDateDeleted)
			if got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("\nSecond variables do not match:\n%v\nwant:\n%v", got1, tt.want1)
			}
		})
	}
}

func TestSetUpMetricValue(t *testing.T) {
	type args struct {
		metric *prometheus.GaugeVec
		value  float64
		labels []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"setUpMetricValueError",
			args{gpbckpExporterInfoMetric, 0, []string{"demo", "bad"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setUpMetricValue(tt.args.metric, tt.args.value, tt.args.labels...); (err != nil) != tt.wantErr {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", err, tt.wantErr)
			}
		})
	}
}
func getLogger() log.Logger {
	var err error
	logLevel := &promlog.AllowedLevel{}
	err = logLevel.Set("info")
	if err != nil {
		panic(err)
	}
	promlogConfig := &promlog.Config{}
	promlogConfig.Level = logLevel
	return promlog.New(promlogConfig)
}

func fakeSetUpMetricValue(metric *prometheus.GaugeVec, value float64, labels ...string) error {
	return errors.New("custom error for test")
}

//nolint:unparam
func templateBackupConfig() gpbckpconfig.BackupConfig {
	return gpbckpconfig.BackupConfig{
		BackupDir:             "/data/backups",
		BackupVersion:         "1.26.0",
		Compressed:            true,
		CompressionType:       "gzip",
		DatabaseName:          "test",
		DatabaseVersion:       "6.23.0",
		DataOnly:              false,
		DateDeleted:           "",
		ExcludeRelations:      []string{},
		ExcludeSchemaFiltered: false,
		ExcludeSchemas:        []string{},
		ExcludeTableFiltered:  false,
		IncludeRelations:      []string{},
		IncludeSchemaFiltered: false,
		IncludeSchemas:        []string{},
		IncludeTableFiltered:  false,
		Incremental:           false,
		LeafPartitionData:     false,
		MetadataOnly:          false,
		Plugin:                "",
		PluginVersion:         "",
		RestorePlan:           []gpbckpconfig.RestorePlanEntry{},
		SingleDataFile:        false,
		Timestamp:             "20230118152654",
		EndTime:               "20230118152656",
		WithoutGlobals:        false,
		WithStatistics:        false,
		Status:                "Success",
	}
}

//nolint:unparam
func templateUnixTime() int64 {
	// Thu Jan 18 2023 20:00:00 UTC
	var curUnixTime int64 = 1674072000
	return curUnixTime
}

func returnTimeTime(sTime string) time.Time {
	var rTime time.Time
	rTime, err := time.Parse(gpbckpconfig.Layout, sTime)
	if err != nil {
		panic(err)
	}
	return rTime
}

func TestDbInList(t *testing.T) {
	type args struct {
		db          string
		listExclude []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"Include",
			args{"test", []string{"test"}},
			true,
		},
		{
			"Exclude",
			args{"test", []string{"demo"}},
			false,
		},
		{
			"Empty list",
			args{"test", []string{""}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dbInList(tt.args.db, tt.args.listExclude); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestListEmpty(t *testing.T) {
	tests := []struct {
		name string
		list []string
		want bool
	}{
		{
			name: "empty list",
			list: []string{},
			want: true,
		},
		{
			name: "non-empty list",
			list: []string{"a", "b", "c"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listEmpty(tt.list); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

// Test only function logic.
func TestParseBackupData(t *testing.T) {
	type args struct {
		historyFile string
		cDeleted    bool
		cFailed     bool
	}
	tests := []struct {
		name    string
		args    args
		want    gpbckpconfig.History
		wantErr bool
	}{
		{
			name: "Test yaml file",
			args: args{
				historyFile: "test*.yaml",
				cDeleted:    false,
				cFailed:     false,
			},
			want:    gpbckpconfig.History{},
			wantErr: false,
		},
		{
			name: "Test db file",
			args: args{
				historyFile: "test*.db",
				cDeleted:    false,
				cFailed:     false,
			},
			want:    gpbckpconfig.History{},
			wantErr: true,
		},
		{
			name: "test unknown file extension",
			args: args{
				historyFile: "test*.txt",
				cDeleted:    false,
				cFailed:     false,
			},
			want:    gpbckpconfig.History{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile, err := os.CreateTemp("", tt.args.historyFile)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())
			got, err := parseBackupData(tempFile.Name(), tt.args.cDeleted, tt.args.cFailed, getLogger())
			if (err != nil) != tt.wantErr {
				t.Errorf("\nVariables do not match:\n%v\nwantErrText:\n%v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}
