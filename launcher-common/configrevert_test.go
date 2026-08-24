package launcher_common

import (
	"testing"

	"github.com/luskaner/ageLANServer/launcher-common/cmd/config"
)

func TestNewConfigRevertFlagOptions(t *testing.T) {
	opts := NewConfigRevertFlagOptions()
	if opts == nil {
		t.Fatal("NewConfigRevertFlagOptions returned nil")
	}
	if opts.RevertValues == nil {
		t.Fatal("RevertValues should be initialized")
	}
}

func TestConfigRevertFlagOptions_FlagsDefaults(t *testing.T) {
	opts := NewConfigRevertFlagOptions()
	flags := opts.Flags()
	// Default: no flags set, should return empty or only logRoot/game
	for _, f := range flags {
		if f == "--logRoot" || f == "--game" {
			continue
		}
		t.Errorf("unexpected default flag: %q", f)
	}
}

func TestConfigRevertFlagOptions_FlagsRemoveAllClearsOthers(t *testing.T) {
	opts := NewConfigRevertFlagOptions()
	opts.RemoveAll = true
	opts.IPs = true
	opts.Certs = true
	opts.Metadata = true
	opts.Profiles = true
	opts.RemoveUserCert = true
	opts.RestoreCAStoreCert = true

	flags := opts.Flags()
	// When RemoveAll is true, individual flags should be cleared
	for _, f := range flags {
		if f == "--ip" || f == "--localCert" || f == "--metadata" || f == "--profiles" || f == "--userCert" || f == "--caStoreCert" {
			t.Errorf("RemoveAll should clear individual flag %q, but it was emitted", f)
		}
	}
}

func TestRevertRequiresAdminElevationValues(t *testing.T) {
	tests := []struct {
		name   string
		values config.RevertValues
		want   bool
	}{
		{"nothing set", config.RevertValues{
			RevertBaseValues: &config.RevertBaseValues{RevertMinimalValues: &config.RevertMinimalValues{}},
			CommonBaseValues: &config.CommonBaseValues{},
		}, false},
		{"certs without path", config.RevertValues{
			RevertBaseValues: &config.RevertBaseValues{RevertMinimalValues: &config.RevertMinimalValues{Certs: true}},
			CommonBaseValues: &config.CommonBaseValues{},
		}, true},
		{"certs with path", config.RevertValues{
			RevertBaseValues: &config.RevertBaseValues{RevertMinimalValues: &config.RevertMinimalValues{Certs: true}},
			CommonBaseValues: &config.CommonBaseValues{CertFilePath: "/cert"},
		}, false},
		{"ips without path", config.RevertValues{
			RevertBaseValues: &config.RevertBaseValues{RevertMinimalValues: &config.RevertMinimalValues{IPs: true}},
			CommonBaseValues: &config.CommonBaseValues{},
		}, true},
		{"ips with path", config.RevertValues{
			RevertBaseValues: &config.RevertBaseValues{RevertMinimalValues: &config.RevertMinimalValues{IPs: true}},
			CommonBaseValues: &config.CommonBaseValues{HostFilePath: "/hosts"},
		}, false},
		{"both without paths", config.RevertValues{
			RevertBaseValues: &config.RevertBaseValues{RevertMinimalValues: &config.RevertMinimalValues{Certs: true, IPs: true}},
			CommonBaseValues: &config.CommonBaseValues{},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RevertRequiresAdminElevationValues(&tt.values)
			if got != tt.want {
				t.Errorf("RevertRequiresAdminElevationValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
