package gpbckpexporter

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/prometheus/common/promlog"
	"github.com/woblerr/gpbackup_exporter/gpbckpfunc"
	"github.com/woblerr/gpbackup_exporter/gpbckpstruct"
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

func TestGetEmptyLabel(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"Empty",
			args{""},
			"none",
		},
		{
			"NotEmpty",
			args{"text"},
			"text",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getEmptyLabel(tt.args.str); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestGetStatusFloat64(t *testing.T) {
	type args struct {
		valueStatus string
	}
	tests := []struct {
		name string
		args args
		want float64
	}{

		{
			"Failure",
			args{"Failure"},
			1,
		},
		{
			"NotFailure",
			args{"text"},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getStatusFloat64(tt.args.valueStatus); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
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

func TestGetExporterMetrics(t *testing.T) {
	type args struct {
		exporterVer         string
		testText            string
		setUpMetricValueFun setUpMetricValueFunType
	}
	templateMetrics := `# HELP gpbackup_exporter_info Information about gpbackup exporter.
# TYPE gpbackup_exporter_info gauge
gpbackup_exporter_info{version="unknown"} 1
`
	tests := []struct {
		name string
		args args
	}{
		{"GetExporterInfoGood",
			args{
				`unknown`,
				templateMetrics,
				setUpMetricValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getExporterMetrics(tt.args.exporterVer, tt.args.setUpMetricValueFun, getLogger())
			reg := prometheus.NewRegistry()
			reg.MustRegister(gpbckpExporterInfoMetric)
			metricFamily, err := reg.Gather()
			if err != nil {
				fmt.Println(err)
			}
			out := &bytes.Buffer{}
			for _, mf := range metricFamily {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					panic(err)
				}
			}
			if tt.args.testText != out.String() {
				t.Errorf("\nVariables do not match:\n%s\nwant:\n%s", tt.args.testText, out.String())
			}
		})
	}
}

func TestGetExporterInfoErrorsAndDebugs(t *testing.T) {
	type args struct {
		exporterVer         string
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetExporterInfoError",
			args{
				`unknown`,
				fakeSetUpMetricValue,
				1,
				1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			logger := log.NewLogfmtLogger(out)
			lc := log.With(logger, level.AllowInfo())
			getExporterMetrics(tt.args.exporterVer, tt.args.setUpMetricValueFun, lc)
			errorsOutputCount := strings.Count(out.String(), "level=error")
			debugsOutputCount := strings.Count(out.String(), "level=debug")
			if tt.args.errorsCount != errorsOutputCount || tt.args.debugsCount != debugsOutputCount {
				t.Errorf("\nVariables do not match:\nerrors=%d, debugs=%d\nwant:\nerrors=%d, debugs=%d",
					tt.args.errorsCount, tt.args.debugsCount,
					errorsOutputCount, debugsOutputCount)
			}
		})
	}
}

// All metrics exist and all labels are corrected.
// gpbackup version >= 1.23.0
func TestGetBackupMetrics(t *testing.T) {
	type args struct {
		backupData          gpbckpstruct.BackupConfig
		setUpMetricValueFun setUpMetricValueFunType
		testText            string
	}
	templateMetrics := `# HELP gpbackup_backup_deleted_status Backup deletion status.
# TYPE gpbackup_backup_deleted_status gauge
gpbackup_backup_deleted_status{backup_type="full",database_name="test",date_deleted="none",object_filtering="none",plugin="none",timestamp="20230118152654"} 0
# HELP gpbackup_backup_duration_seconds Backup duration.
# TYPE gpbackup_backup_duration_seconds gauge
gpbackup_backup_duration_seconds{backup_type="full",database_name="test",object_filtering="none",plugin="none",timestamp="20230118152654"} 2
# HELP gpbackup_backup_info Backup info.
# TYPE gpbackup_backup_info gauge
gpbackup_backup_info{backup_dir="/data/backups",backup_type="full",backup_ver="1.26.0",compression_type="gzip",database_name="test",database_ver="6.23.0",object_filtering="none",plugin="none",plugin_ver="none",timestamp="20230118152654",with_statistic="false"} 1
# HELP gpbackup_backup_status Backup status.
# TYPE gpbackup_backup_status gauge
gpbackup_backup_status{backup_type="full",database_name="test",object_filtering="none",plugin="none",timestamp="20230118152654"} 0
`
	tests := []struct {
		name string
		args args
	}{
		{"GetBackupMetricsGood",
			args{
				templateBackupConfig(),
				setUpMetricValue,
				templateMetrics,
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getBackupMetrics(tt.args.backupData, tt.args.setUpMetricValueFun, getLogger())
			reg := prometheus.NewRegistry()
			reg.MustRegister(
				gpbckpBackupStatusMetric,
				gpbckpBackupDataDeletedStatusMetric,
				gpbckpBackupInfoMetric,
				gpbckpBackupDurationMetric,
			)
			metricFamily, err := reg.Gather()
			if err != nil {
				fmt.Println(err)
			}
			out := &bytes.Buffer{}
			for _, mf := range metricFamily {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					panic(err)
				}
			}
			if tt.args.testText != out.String() {
				t.Errorf("\nVariables do not match:\n%s\nwant:\n%s", tt.args.testText, out.String())
			}
		})
	}
}

func TestGetBackupMetricsErrorsAndDebugs(t *testing.T) {
	type args struct {
		backupData          gpbckpstruct.BackupConfig
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetBackupMetricsErrorGetDurationGood",
			args{
				templateBackupConfig(),
				fakeSetUpMetricValue,
				4,
				4,
			},
		},
		{"GetBackupMetricsErrorGetDurationError",
			args{
				gpbckpstruct.BackupConfig{},
				fakeSetUpMetricValue,
				4,
				3,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			logger := log.NewLogfmtLogger(out)
			lc := log.With(logger, level.AllowInfo())
			getBackupMetrics(tt.args.backupData, tt.args.setUpMetricValueFun, lc)
			errorsOutputCount := strings.Count(out.String(), "level=error")
			debugsOutputCount := strings.Count(out.String(), "level=debug")
			if tt.args.errorsCount != errorsOutputCount || tt.args.debugsCount != debugsOutputCount {
				t.Errorf("\nVariables do not match:\nerrors=%d, debugs=%d\nwant:\nerrors=%d, debugs=%d",
					tt.args.errorsCount, tt.args.debugsCount,
					errorsOutputCount, debugsOutputCount)
			}
		})
	}
}

func TestGetBackupLastMetrics(t *testing.T) {
	type args struct {
		lastBackups         lastBackupMap
		setUpMetricValueFun setUpMetricValueFunType
		testText            string
	}
	templateMetrics := `# HELP gpbackup_backup_since_last_completion_seconds Seconds since the last completed backup.
# TYPE gpbackup_backup_since_last_completion_seconds gauge
gpbackup_backup_since_last_completion_seconds{backup_type="data-only",database_name="test"} 7200
gpbackup_backup_since_last_completion_seconds{backup_type="full",database_name="test"} 18000
gpbackup_backup_since_last_completion_seconds{backup_type="incremental",database_name="test"} 14400
gpbackup_backup_since_last_completion_seconds{backup_type="metadata-only",database_name="test"} 10800
`
	tests := []struct {
		name string
		args args
	}{
		{"GetBackupMetricsGood",
			args{
				lastBackupMap{
					"test": backupMap{
						"full":          returnTimeTime("20230118150000"),
						"incremental":   returnTimeTime("20230118160000"),
						"metadata-only": returnTimeTime("20230118170000"),
						"data-only":     returnTimeTime("20230118180000"),
					}},
				setUpMetricValue,
				templateMetrics,
			},
		}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getBackupLastMetrics(tt.args.lastBackups, templateUnixTime(), tt.args.setUpMetricValueFun, getLogger())
			reg := prometheus.NewRegistry()
			reg.MustRegister(
				gpbckpBackupSinceLastCompletionSecondsMetric,
			)
			metricFamily, err := reg.Gather()
			if err != nil {
				fmt.Println(err)
			}
			out := &bytes.Buffer{}
			for _, mf := range metricFamily {
				if _, err := expfmt.MetricFamilyToText(out, mf); err != nil {
					panic(err)
				}
			}
			if tt.args.testText != out.String() {
				t.Errorf("\nVariables do not match:\n%s\nwant:\n%s", tt.args.testText, out.String())
			}
		})
	}
}

func TestGetBackupLastMetricsErrorsAndDebugs(t *testing.T) {
	type args struct {
		lastBackups         lastBackupMap
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetBackupMetricsError",
			args{
				lastBackupMap{
					"test": backupMap{
						"full": returnTimeTime("20230118150000"),
					}}, fakeSetUpMetricValue,
				1,
				1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			logger := log.NewLogfmtLogger(out)
			lc := log.With(logger, level.AllowInfo())
			getBackupLastMetrics(tt.args.lastBackups, templateUnixTime(), tt.args.setUpMetricValueFun, lc)
			errorsOutputCount := strings.Count(out.String(), "level=error")
			debugsOutputCount := strings.Count(out.String(), "level=debug")
			if tt.args.errorsCount != errorsOutputCount || tt.args.debugsCount != debugsOutputCount {
				t.Errorf("\nVariables do not match:\nerrors=%d, debugs=%d\nwant:\nerrors=%d, debugs=%d",
					tt.args.errorsCount, tt.args.debugsCount,
					errorsOutputCount, debugsOutputCount)
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
	return errors.New("сustorm error for test")
}

//nolint:unparam
func templateBackupConfig() gpbckpstruct.BackupConfig {
	return gpbckpstruct.BackupConfig{
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
		RestorePlan:           []gpbckpstruct.RestorePlanEntry{},
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
	rTime, err := time.Parse(gpbckpfunc.Layout, sTime)
	if err != nil {
		panic(err)
	}
	return rTime
}
