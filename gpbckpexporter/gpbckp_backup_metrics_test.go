package gpbckpexporter

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
	"github.com/woblerr/gpbackman/gpbckpconfig"
)

// All metrics exist and all labels are corrected.
// gpbackup version >= 1.23.0
func TestGetBackupMetrics(t *testing.T) {
	type args struct {
		backupData          gpbckpconfig.BackupConfig
		setUpMetricValueFun setUpMetricValueFunType
		testText            string
	}
	templateMetrics := `# HELP gpbackup_backup_deleted_status Backup deletion status.
# TYPE gpbackup_backup_deleted_status gauge
gpbackup_backup_deleted_status{backup_type="full",database_name="test",date_deleted="none",object_filtering="none",plugin="none",timestamp="20230118152654"} 0
# HELP gpbackup_backup_duration_seconds Backup duration.
# TYPE gpbackup_backup_duration_seconds gauge
gpbackup_backup_duration_seconds{backup_type="full",database_name="test",end_time="20230118152656",object_filtering="none",plugin="none",timestamp="20230118152654"} 2
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
			resetBackupMetrics()
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
		backupData          gpbckpconfig.BackupConfig
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
				gpbckpconfig.BackupConfig{},
				fakeSetUpMetricValue,
				5,
				4,
			},
		},
		{"GetBackupMetricsErrorGetBackupTypeAndObjectFilteringError",
			args{
				// Fake example for testing.
				gpbckpconfig.BackupConfig{
					DataOnly:              true,
					Incremental:           true,
					IncludeSchemaFiltered: true,
					ExcludeSchemaFiltered: true,
				},
				fakeSetUpMetricValue,
				7,
				4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetBackupMetrics()
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
