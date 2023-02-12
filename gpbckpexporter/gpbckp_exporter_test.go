package gpbckpexporter

import (
	"bytes"
	"errors"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-kit/log"
)

func TestSetPromPortAndPath(t *testing.T) {
	var (
		testPort          = "19854"
		testEndpoint      = "/metrics"
		testTLSConfigPath = ""
	)
	SetPromPortAndPath(testPort, testEndpoint, testTLSConfigPath)
	if testPort != promPort || testEndpoint != promEndpoint || testTLSConfigPath != promTLSConfigPath {
		t.Errorf("\nVariables do not match,\nport: %s, want: %s;\nendpoint: %s, want: %s;\nconfig: %swant: %s",
			testPort, promPort,
			testEndpoint, promEndpoint,
			testTLSConfigPath, promTLSConfigPath,
		)
	}
}

func TestGetGPBackupInfo(t *testing.T) {
	type args struct {
		historyFile string
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
  databasename: core
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
  status: Success`},
			"level=debug msg=\"Metric gpbackup_backup_status\" value=0 labels=full,core,none,none,20230118152654",
		},
		{
			"FailedDataReturn",
			args{"return error"},
			"level=error msg=\"Read gpbackup history file failed\" err=\"Error for testing\"",
		},
		{
			"InvalidDataReturn",
			args{"42"},
			"level=error msg=\"Parse YAML failed\" err=\"yaml: unmarshal errors:\\n  line 1: cannot unmarshal !!int `42` into gpbckpstruct.History\"",
		},
		{
			"NoDataReturn",
			args{""},
			"level=warn msg=\"No backup data returned\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetMetrics()
			execReadFile = fakeReadFile
			defer func() { execReadFile = ioutil.ReadFile }()
			out := &bytes.Buffer{}
			lc := log.NewLogfmtLogger(out)
			GetGPBackupInfo(tt.args.historyFile, lc)
			if !strings.Contains(out.String(), tt.testText) {
				t.Errorf("\nVariable do not match:\n%s\nwant:\n%s", tt.testText, out.String())
			}
		})
	}
}

func fakeReadFile(filename string) ([]byte, error) {
	if filename == "return error" {
		return []byte{}, errors.New("Error for testing")
	}
	buf := bytes.NewBufferString(filename)
	return ioutil.ReadAll(buf)
}
