package gpbckpexporter

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/expfmt"
)

func TestGetExporterStatusMetrics(t *testing.T) {
	type args struct {
		dbStatus            dbStatusMap
		testText            string
		setUpMetricValueFun setUpMetricValueFunType
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetExporterStatusGood",
			args{
				dbStatusMap{"test": true},
				`# HELP gpbackup_exporter_status gpbackup exporter get data status.
# TYPE gpbackup_exporter_status gauge
gpbackup_exporter_status{database_name="test"} 1
`,
				setUpMetricValue,
			},
		},
		{"GetExporterStatusBad",
			args{
				dbStatusMap{"test": false},
				`# HELP gpbackup_exporter_status gpbackup exporter get data status.
# TYPE gpbackup_exporter_status gauge
gpbackup_exporter_status{database_name="test"} 0
`,
				setUpMetricValue,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetExporterMetrics()
			getExporterStatusMetrics(tt.args.dbStatus, tt.args.setUpMetricValueFun, getLogger())
			reg := prometheus.NewRegistry()
			reg.MustRegister(gpbckpExporterStatusMetric)
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

func TestGetExporterStatusErrorsAndDebugs(t *testing.T) {
	type args struct {
		dbStatus            dbStatusMap
		setUpMetricValueFun setUpMetricValueFunType
		errorsCount         int
		debugsCount         int
	}
	tests := []struct {
		name string
		args args
	}{
		{"GetExporterInfoLogError",
			args{
				dbStatusMap{"test": true},
				fakeSetUpMetricValue,
				1,
				1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			lc := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelDebug}))
			getExporterStatusMetrics(tt.args.dbStatus, tt.args.setUpMetricValueFun, lc)
			errorsOutputCount := strings.Count(out.String(), "level=ERROR")
			debugsOutputCount := strings.Count(out.String(), "level=DEBUG")
			if tt.args.errorsCount != errorsOutputCount || tt.args.debugsCount != debugsOutputCount {
				t.Errorf("\nVariables do not match:\nerrors=%d, debugs=%d\nwant:\nerrors=%d, debugs=%d",
					tt.args.errorsCount, tt.args.debugsCount,
					errorsOutputCount, debugsOutputCount)
			}
		})
	}
}
