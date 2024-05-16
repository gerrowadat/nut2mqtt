package upsc

import (
	"reflect"
	"testing"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
)

var GetKeyValueFromListLine = getKeyValueFromListLine

// Mock UPSDClient that implements the UPSDClientIf interface
// and just returns whatever raw output is passed to NewUPSDMockClient()
type UPSDMockClient struct {
	host string
	port int
	raw  string
}

func NewUPSDMockClient(host string, port int, raw string) *UPSDMockClient {
	return &UPSDMockClient{host: host, port: port, raw: raw}
}

func (upsd_c *UPSDMockClient) Host() string {
	return upsd_c.host
}

func (upsd_c *UPSDMockClient) Port() int {
	return upsd_c.port
}

func (upsd_c *UPSDMockClient) Request(cmd string) (string, error) {
	return upsd_c.raw, nil
}

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
			name:    "Blank",
			args:    args{line: ""},
			wantk:   "",
			wantv:   "",
			wantErr: false,
		},
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
		{
			name:    "ListNoLines",
			args:    args{response: "BEGIN LIST UPS\nEND LIST UPS\n", cmd: "LIST UPS"},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name:    "ListNoBegin",
			args:    args{response: "UPS myups \"description\"\nEND LIST UPS\n", cmd: "LIST UPS"},
			wantErr: true,
		},
		{
			name:    "ListNoEnd",
			args:    args{response: "BEGIN LIST UPS\nUPS myups \"description\"\n", cmd: "LIST UPS"},
			wantErr: true,
		},
		{
			name:    "ListOneElem",
			args:    args{response: "BEGIN LIST UPS\nUPS myups \"description\"\nEND LIST UPS\n", cmd: "LIST UPS"},
			wantErr: false,
			want:    map[string]string{"myups": "description"},
		},
		{
			name:    "ListSeveralElem",
			args:    args{response: "BEGIN LIST UPS\nUPS myups \"description\"\nUPS myotherups \"other one\"\nEND LIST UPS\n", cmd: "LIST UPS"},
			wantErr: false,
			want:    map[string]string{"myups": "description", "myotherups": "other one"},
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

func TestUpsdCommand(t *testing.T) {
	type args struct {
		upsd_c *UPSDClient
		cmd    string
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "NonListOrGet",
			args:    args{upsd_c: &UPSDClient{}, cmd: "FUNGE blarg"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := UpsdCommand(tt.args.upsd_c, tt.args.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpsdCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UpsdCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUPSes(t *testing.T) {
	type args struct {
		upsd_c UPSDClientIf
	}
	tests := []struct {
		name    string
		args    args
		want    []*channels.UPSInfo
		wantErr bool
	}{
		{
			name:    "NoUPSes",
			args:    args{upsd_c: NewUPSDMockClient("localhost", 3493, "BEGIN LIST UPS\nEND LIST UPS\n")},
			want:    []*channels.UPSInfo{},
			wantErr: false,
		},
		{
			name:    "OneResponse",
			args:    args{upsd_c: NewUPSDMockClient("localhost", 3493, "BEGIN LIST UPS\nUPS myups \"description\"\nEND LIST UPS\n")},
			want:    []*channels.UPSInfo{{Name: "myups", Description: "description", Vars: map[string]string{}}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUPSes(tt.args.upsd_c)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUPSes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUPSes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetUpdatedVars(t *testing.T) {
	type args struct {
		upsd_c UPSDClientIf
		u      *channels.UPSInfo
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]string
		wantErr bool
	}{
		{
			name: "NoChanges",
			args: args{upsd_c: NewUPSDMockClient("localhost", 3493, "BEGIN LIST VAR myups\nVAR myups stuff.things \"yokes\"\nEND LIST VAR myups"),
				u: &channels.UPSInfo{Name: "myups", Description: "description", Vars: map[string]string{"stuff.things": "yokes"}}},
			want:    map[string]string{},
			wantErr: false,
		},
		{
			name: "UpdateInPlace",
			args: args{upsd_c: NewUPSDMockClient("localhost", 3493, "BEGIN LIST VAR myups\nVAR myups stuff.things \"1\"\nEND LIST VAR myups"),
				u: &channels.UPSInfo{Name: "myups", Description: "description", Vars: map[string]string{"stuff.things": "2"}}},
			want:    map[string]string{"stuff.things": "1"},
			wantErr: false,
		},
		{
			name: "NewVarAppears",
			args: args{upsd_c: NewUPSDMockClient("localhost", 3493, "BEGIN LIST VAR myups\nVAR myups stuff.things \"1\"\nVAR myups yokes.etc \"2\"\nEND LIST VAR myups"),
				u: &channels.UPSInfo{Name: "myups", Description: "description", Vars: map[string]string{"stuff.things": "1"}}},
			want:    map[string]string{"yokes.etc": "2"},
			wantErr: false,
		},
		{
			name: "OldVarDisppears",
			args: args{upsd_c: NewUPSDMockClient("localhost", 3493, "BEGIN LIST VAR myups\nVAR myups stuff.things \"1\"\nEND LIST VAR myups"),
				u: &channels.UPSInfo{Name: "myups", Description: "description", Vars: map[string]string{"stuff.things": "1", "yokes.etc": "2"}}},
			want:    map[string]string{"yokes.etc": ""},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUpdatedVars(tt.args.upsd_c, tt.args.u)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUpdatedVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetUpdatedVars() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVarDiff(t *testing.T) {
	type args struct {
		old *channels.UPSInfo
		new *channels.UPSInfo
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "NoChanges",
			args: args{old: &channels.UPSInfo{Vars: map[string]string{"stuff.things": "yokes"}}, new: &channels.UPSInfo{Vars: map[string]string{"stuff.things": "yokes"}}},
			want: map[string]string{},
		},
		{
			name: "ChangeFromNone",
			args: args{old: &channels.UPSInfo{}, new: &channels.UPSInfo{Vars: map[string]string{"stuff.things": "yokes"}}},
			want: map[string]string{"stuff.things": "yokes"},
		},
		{
			name: "ChangeExisting",
			args: args{old: &channels.UPSInfo{Vars: map[string]string{"stuff.things": "yokes"}}, new: &channels.UPSInfo{Vars: map[string]string{"stuff.things": "notyokes"}}},
			want: map[string]string{"stuff.things": "notyokes"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetVarDiff(tt.args.old, tt.args.new); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetVarDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}
