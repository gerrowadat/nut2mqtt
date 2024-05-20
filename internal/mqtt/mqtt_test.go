package mqtt

import (
	"testing"

	channels "github.com/gerrowadat/nut2mqtt/internal/channels"
)

func TestTopicFromUPSVariableUpdate(t *testing.T) {
	type args struct {
		up *channels.UPSVariableUpdate
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "SingleLayer",
			args: args{up: &channels.UPSVariableUpdate{Host: "host1", UpsName: "ups1", VarName: "charge", Content: "100"}},
			want: "hosts/host1/ups1/charge",
		},
		{
			name: "MultiLayer",
			args: args{up: &channels.UPSVariableUpdate{Host: "host1", UpsName: "ups1", VarName: "battery.charge", Content: "100"}},
			want: "hosts/host1/ups1/battery/charge",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TopicFromUPSVariableUpdate(tt.args.up); got != tt.want {
				t.Errorf("TopicFromUPSVariableUpdate() = %v, want %v", got, tt.want)
			}
		})
	}
}
