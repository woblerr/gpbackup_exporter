package gpbckpexporter

import "testing"

func TestConvertBoolToFloat64(t *testing.T) {
	type args struct {
		value bool
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			"True",
			args{true},
			1,
		},
		{
			"False",
			args{false},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := convertBoolToFloat64(tt.args.value); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestConvertEmptyLabel(t *testing.T) {
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
			if got := convertEmptyLabel(tt.args.str); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}

func TestConvertStatusFloat64(t *testing.T) {
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
			if got := convertStatusFloat64(tt.args.valueStatus); got != tt.want {
				t.Errorf("\nVariables do not match:\n%v\nwant:\n%v", got, tt.want)
			}
		})
	}
}
