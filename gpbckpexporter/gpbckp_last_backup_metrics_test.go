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
)

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
			resetLastBackupMetrics()
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
			resetLastBackupMetrics()
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
