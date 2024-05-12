package upsc

import (
	"reflect"
	"testing"
)

var GetKeyValueFromListLine = getKeyValueFromListLine

func Test_getKeyValueFromListLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		wantk   string
		wantv   string
		wantErr bool
	}{
		{
			name:    "UknownNoun",
			args:    args{line: "TEAPOT name  capacity"},
			wantk:   "",
			wantv:   "",
			wantErr: true,
		},
		{
			name:    "NormalUPS",
			args:    args{line: "UPS myups \"makes a beeping sound\""},
			wantk:   "myups",
			wantv:   "makes a beeping sound",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k, v, err := GetKeyValueFromListLine(tt.args.line)
			if (err != nil) != tt.wantErr {
				t.Errorf("getKeyValueFromListLine() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if k != tt.wantk {
				t.Errorf("getKeyValueFromListLine() got = %v, want %v", k, tt.wantk)
			}
			if v != tt.wantv {
				t.Errorf("getKeyValueFromListLine() got1 = %v, want %v", k, tt.wantv)
			}
		})
	}
}

func Test_processUpsdResponse(t *testing.T) {
	type args struct {
		response string
		cmd      string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "GETNoLines",
			args:    args{response: "", cmd: "GET VAR myups stuff.things"},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "GETMultiLine",
			args:    args{response: "line\nother line", cmd: "GET VAR myups stuff.things"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "GETNormal",
			args:    args{response: "VAR myups stuff.things \"yokes\"", cmd: "GET VAR myups stuff.things"},
			want:    map[string]string{"stuff.things": "yokes"},
			wantErr: false,
		},
		{
			name:    "NotImplementedIndeed",
			args:    args{response: "UPS stuff \"things\"", cmd: "FUNGE blarg"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processUpsdResponse(tt.args.response, tt.args.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("processUpsdResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processUpsdResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
