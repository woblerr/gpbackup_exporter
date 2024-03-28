package gpbckpexporter

import (
	"bytes"
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
		historyData  string
		bckpType     string
		bckpCDeleted bool
		bckpCFailed  bool
		bckpIncl     []string
		bckpExcl     []string
		cDepth       int
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
				false,
				false,
				[]string{""},
				[]string{""},
				0,
			},
			`level=debug msg="Set up metric" metric=gpbackup_backup_status value=0 labels=full,test,none,none,20230118152654
level=debug msg="Set up metric" metric=gpbackup_backup_deleted_status value=0 labels=full,test,none,none,none,20230118152654
level=debug msg="Set up metric" metric=gpbackup_backup_info value=1 labels=/data/backups,1.26.0,full,gzip,test,6.23.0,none,none,none,20230118152654,false
level=debug msg="Set up metric" metric=gpbackup_backup_duration_seconds value=2 labels=full,test,20230118152656,none,none,20230118152654
level=debug msg="Set up metric" metric=gpbackup_backup_status value=0 labels=metadata-only,test,none,none,20230118162654
level=debug msg="Set up metric" metric=gpbackup_backup_deleted_status value=0 labels=metadata-only,test,none,none,none,20230118162654
level=debug msg="Set up metric" metric=gpbackup_backup_info value=1 labels=/data/backups,1.26.0,metadata-only,gzip,test,6.23.0,none,none,none,20230118162654,false
level=debug msg="Set up metric" metric=gpbackup_backup_duration_seconds value=2 labels=metadata-only,test,20230118162656,none,none,20230118162654
`,
		},
		{
			"FailedDataReturn",
			args{"return error", "", false, false, []string{""}, []string{""}, 0},
			"level=error msg=\"Parse YAML failed\" err=\"yaml: unmarshal errors:\\n  line 1: cannot unmarshal !!str `return ...` into gpbckpconfig.History\"",
		},
		{
			"NoDataReturn",
			args{"", "", false, false, []string{""}, []string{""}, 0},
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
				false,
				false,
				[]string{""},
				[]string{""},
				14,
			},
			"level=warn msg=\"No succeed backups\"",
		},
		{
			"DBinIncludeAndExclude",
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
				false,
				false,
				[]string{"test"},
				[]string{"test"},
				0,
			},
			"level=warn msg=\"DB is specified in include and exclude lists\" DB=test",
		},
		{
			"ErrorsInParseValues",
			// Set dataonly: true, incremental:true and metadataonly: true, that's invalid.
			args{`backupconfigs:
- backupdir: "/data/backups"
  backupversion: 1.26.0
  compressed: true
  compressiontype: gzip
  databasename: test
  databaseversion: 6.23.0
  dataonly: true
  datedeleted: ""
  excluderelations: []
  excludeschemafiltered: false
  excludeschemas: []
  excludetablefiltered: false
  includerelations: []
  includeschemafiltered: false
  includeschemas: []
  includetablefiltered: false
  incremental: true
  leafpartitiondata: true
  metadataonly: true
  plugin: ""
  pluginversion: ""
  restoreplan: []
  singledatafile: false
  timestamp: "test"
  endtime: "test"
  withoutglobals: false
  withstatistics: false
  status: Success`,
				"",
				false,
				false,
				[]string{""},
				[]string{""},
				0,
			},
			`level=error msg="Parse backup timestamp value failed" err="parsing time \"test\" as \"20060102150405\": cannot parse \"test\" as \"2006\""
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetMetrics()
			tempFile, err := fakeHistoryFileData(tt.args.historyData)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}
			defer os.Remove(tempFile.Name())
			out := &bytes.Buffer{}
			lc := log.NewLogfmtLogger(out)
			GetGPBackupInfo(
				tempFile.Name(),
				tt.args.bckpType,
				tt.args.bckpCDeleted,
				tt.args.bckpCFailed,
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

func fakeHistoryFileData(text string) (*os.File, error) {
	tempFile, err := os.CreateTemp("", "gpbackup_history*.yaml")
	if err != nil {
		return nil, err
	}
	if _, err := tempFile.Write([]byte(text)); err != nil {
		return nil, err
	}
	return tempFile, nil
}

// Helper for displaying web.FlagConfig values test messages.
func ptrToVal[T any](v *T) T {
	return *v
}
