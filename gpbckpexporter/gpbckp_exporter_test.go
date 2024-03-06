package gpbckpexporter

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/exporter-toolkit/web"
)

func TestSetPromPortAndPath(t *testing.T) {
	var (
		testFlagsConfig = web.FlagConfig{
			WebListenAddresses: &([]string{":9854"}),
			WebSystemdSocket:   func(i bool) *bool { return &i }(false),
			WebConfigFile:      func(i string) *string { return &i }(""),
		}
		testEndpoint = "/metrics"
	)
	SetPromPortAndPath(testFlagsConfig, testEndpoint)
	if testFlagsConfig.WebListenAddresses != webFlagsConfig.WebListenAddresses ||
		testFlagsConfig.WebSystemdSocket != webFlagsConfig.WebSystemdSocket ||
		testFlagsConfig.WebConfigFile != webFlagsConfig.WebConfigFile ||
		testEndpoint != webEndpoint {
		t.Errorf("\nVariables do not match,\nlistenAddresses: %v, want: %v;\n"+
			"systemSocket: %v, want: %v;\nwebConfig: %v, want: %v;\nendpoint: %s, want: %s",
			ptrToVal(testFlagsConfig.WebListenAddresses), ptrToVal(webFlagsConfig.WebListenAddresses),
			ptrToVal(testFlagsConfig.WebSystemdSocket), ptrToVal(webFlagsConfig.WebSystemdSocket),
			ptrToVal(testFlagsConfig.WebConfigFile), ptrToVal(webFlagsConfig.WebConfigFile),
			testEndpoint, webEndpoint,
		)
	}
}

func TestGetGPBackupInfo(t *testing.T) {
	type args struct {
		historyFile string
		bckpType    string
		bckpIncl    []string
		bckpExcl    []string
		cDepth      int
	}
	tests := []struct {
		name     string
		args     args
		testText string
	}{
		{
			"GoodDataReturn",
			args{`backupconfigs:
- backupdir: "/data/backups"
  backupversion: 1.26.0
  compressed: true
  compressiontype: gzip
  databasename: test
  databaseversion: 6.23.0
  dataonly: false
  datedeleted: ""
  excluderelations: []
  excludeschemafiltered: false
  excludeschemas: []
  excludetablefiltered: false
  includerelations: []
  includeschemafiltered: false
  includeschemas: []
  includetablefiltered: false
  incremental: false
  leafpartitiondata: false
  metadataonly: false
  plugin: ""
  pluginversion: ""
  restoreplan: []
  singledatafile: false
  timestamp: "20230118152654"
  endtime: "20230118152656"
  withoutglobals: false
  withstatistics: false
  status: Success
  backupconfigs:
- backupdir: "/data/backups"
  backupversion: 1.26.0
  compressed: true
  compressiontype: gzip
  databasename: test
  databaseversion: 6.23.0
  dataonly: false
  datedeleted: ""
  excluderelations: []
  excludeschemafiltered: false
  excludeschemas: []
  excludetablefiltered: false
  includerelations: []
  includeschemafiltered: false
  includeschemas: []
  includetablefiltered: false
  incremental: false
  leafpartitiondata: false
  metadataonly: true
  plugin: ""
  pluginversion: ""
  restoreplan: []
  singledatafile: false
  timestamp: "20230118162654"
  endtime: "20230118162656"
  withoutglobals: false
  withstatistics: false
  status: Success`,
				"",
				[]string{""},
				[]string{""},
				0,
			},
			"level=debug msg=\"Metric gpbackup_backup_status\" value=0 labels=full,test,none,none,20230118152654",
		},
		{
			"FailedDataReturn",
			args{"return error", "", []string{""}, []string{""}, 0},
			"level=error msg=\"Read gpbackup history file failed\" err=\"Error for testing\"",
		},
		{
			"InvalidDataReturn",
			args{"42", "", []string{""}, []string{""}, 0},
			"level=error msg=\"Parse YAML failed\" err=\"yaml: unmarshal errors:\\n  line 1: cannot unmarshal !!int `42` into gpbckpstruct.History\"",
		},
		{
			"NoDataReturn",
			args{"", "", []string{""}, []string{""}, 0},
			"level=warn msg=\"No backup data returned\"",
		},
		{
			"UseDepthAndOlderDepthInterval",
			args{`backupconfigs:
- backupdir: "/data/backups"
  backupversion: 1.26.0
  compressed: true
  compressiontype: gzip
  databasename: test
  databaseversion: 6.23.0
  dataonly: false
  datedeleted: ""
  excluderelations: []
  excludeschemafiltered: false
  excludeschemas: []
  excludetablefiltered: false
  includerelations: []
  includeschemafiltered: false
  includeschemas: []
  includetablefiltered: false
  incremental: false
  leafpartitiondata: false
  metadataonly: false
  plugin: ""
  pluginversion: ""
  restoreplan: []
  singledatafile: false
  timestamp: "20230118152654"
  endtime: "20230118152656"
  withoutglobals: false
  withstatistics: false
  status: Success`,
				"",
				[]string{""},
				[]string{""},
				14},
			"level=warn msg=\"No succeed backups\"",
		},
		{
			"DBinIncludeAndExclude",
			args{"", "", []string{"test"}, []string{"test"}, 0},
			"level=warn msg=\"DB is specified in include and exclude lists\" DB=test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetMetrics()
			execReadFile = fakeReadFile
			defer func() { execReadFile = os.ReadFile }()
			out := &bytes.Buffer{}
			lc := log.NewLogfmtLogger(out)
			GetGPBackupInfo(
				tt.args.historyFile,
				tt.args.bckpType,
				tt.args.bckpIncl,
				tt.args.bckpExcl,
				tt.args.cDepth,
				lc,
			)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func TestDBNotInExclude(t *testing.T) {
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
			false,
		},
		{
			"Exclude",
			args{"test", []string{"demo"}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := dbNotInExclude(tt.args.db, tt.args.listExclude); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func fakeReadFile(filename string) ([]byte, error) {
	if filename == "return error" {
		return []byte{}, errors.New("Error for testing")
	}
	buf := bytes.NewBufferString(filename)
	return io.ReadAll(buf)
}

// Helper for displaying web.FlagConfig values test messages.
func ptrToVal[T any](v *T) T {
	return *v
}
