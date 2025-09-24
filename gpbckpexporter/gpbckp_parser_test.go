package gpbckpexporter

import (
	"bytes"
	"errors"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
			args{gpbckpExporterStatusMetric, 0, []string{"demo", "bad"}},
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

func getLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func fakeSetUpMetricValue(metric *prometheus.GaugeVec, value float64, labels ...string) error {
	return errors.New("custom error for test")
}

// Create a SQLite database file with missing tables.
func createCorruptedDBFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "test_corrupted_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()
	return tempFile.Name()
}

// Create a database with invalid backup name.
func createDBWithInvalidBackupName(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "test_invalid_backup_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer tempFile.Close()
	// Create a database with valid schema but insert an invalid backup name.
	db, err := gpbckpconfig.OpenHistoryDB(tempFile.Name())
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS backups (
		timestamp TEXT PRIMARY KEY,
		date_deleted TEXT,
		database_name TEXT,
		status TEXT
	)`)
	if err != nil {
		t.Fatalf("Failed to create backups table: %v", err)
	}
	// Insert a backup with an invalid timestamp.
	_, err = db.Exec(`INSERT INTO backups (timestamp, date_deleted, database_name, status) VALUES ('invalid_backup_name', '', 'testdb', 'Success')`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	return tempFile.Name()
}

func templateBackupConfig() gpbckpconfig.BackupConfig {
	return gpbckpconfig.BackupConfig{
		BackupDir:             "/data/backups",
		BackupVersion:         "1.30.5",
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
			wantErr: true,
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

func TestGetDataFromHistoryDB(t *testing.T) {
	type args struct {
		historyFile    string
		collectDeleted bool
		collectFailed  bool
		cleanUpTestDB  bool
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		errText string
	}{
		{
			"InvalidDBFile",
			args{
				historyFile:    "/nonexistent/path/to/db.db",
				collectDeleted: false,
				collectFailed:  false,
				cleanUpTestDB:  false,
			},
			true,
			"level=ERROR msg=\"Get backups from history db failed\"",
		},
		{
			"CorruptedDBWithInvalidBackupData",
			args{
				historyFile:    createCorruptedDBFile(t),
				collectDeleted: false,
				collectFailed:  false,
				cleanUpTestDB:  true,
			},
			true,
			"level=ERROR msg=\"Get backups from history db failed\"",
		},
		{
			"DBWithInvalidBackupName",
			args{
				historyFile:    createDBWithInvalidBackupName(t),
				collectDeleted: false,
				collectFailed:  false,
				cleanUpTestDB:  true,
			},
			true,
			"level=ERROR msg=\"Get backup data from history db failed\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.cleanUpTestDB {
				defer os.Remove(tt.args.historyFile)
			}
			out := &bytes.Buffer{}
			logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelError}))
			_, err := getDataFromHistoryDB(tt.args.historyFile, tt.args.collectDeleted, tt.args.collectFailed, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("getDataFromHistoryDB() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && tt.errText != "" {
				logOutput := out.String()
				if !strings.Contains(logOutput, tt.errText) {
					t.Errorf("\nVariables do not match:\n%v\nwantErrText:\n%v", logOutput, tt.errText)
				}
			}
		})
	}
}
