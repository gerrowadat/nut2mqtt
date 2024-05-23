// Program Control nonsense.

package control

import (
	"testing"
	"time"

	"github.com/gerrowadat/nut2mqtt/internal/channels"
)

func TestPruneUPSCache(t *testing.T) {
	type args struct {
		cache  map[string]*DecayingUPSCacheEntry
		expiry string
	}
	tests := []struct {
		name     string
		args     args
		want_len int
	}{
		{
			name: "NoCacheEntries",
			args: args{
				cache:  map[string]*DecayingUPSCacheEntry{},
				expiry: "30s",
			},
			want_len: 0,
		},
		{
			name: "OneUnexpiredCacheEntry",
			args: args{
				cache: map[string]*DecayingUPSCacheEntry{
					"test": {
						ups:       &channels.UPSInfo{},
						last_seen: time.Now(),
					},
				},
				expiry: "30s",
			},
			want_len: 1,
		},
		{
			name: "OneExpiredCacheEntry",
			args: args{
				cache: map[string]*DecayingUPSCacheEntry{
					"test": {
						ups:       &channels.UPSInfo{},
						last_seen: time.Now().Add(-time.Duration(31) * time.Second),
					},
				},
				expiry: "30s",
			},
			want_len: 0,
		},
		{
			name: "OneExpiredOneUnexpiredCacheEntry",
			args: args{
				cache: map[string]*DecayingUPSCacheEntry{
					"test": {
						ups:       &channels.UPSInfo{},
						last_seen: time.Now().Add(-time.Duration(31) * time.Second),
					},
					"test2": {
						ups:       &channels.UPSInfo{},
						last_seen: time.Now(),
					},
				},
				expiry: "30s",
			},
			want_len: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			duration, err := time.ParseDuration(tt.args.expiry)
			if err != nil {
				t.Errorf("PruneUPSCache() error = %v", err)
			}
			PruneUPSCache(tt.args.cache, duration)
			if len(tt.args.cache) != tt.want_len {
				t.Errorf("PruneUPSCache() = %v, want %v", len(tt.args.cache), tt.want_len)
			}
		})
	}
}
