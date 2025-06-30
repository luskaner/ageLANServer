package server

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/config/shared"
	"github.com/spf13/viper"
	"net"
	"net/netip"
)

type Announcement struct {
	Disabled       bool
	Port           uint   `validate:"min=1025,max=65535"`
	MulticastGroup net.IP `validate:"ip_v4,ip_multicast"`
}

type Listen struct {
	Hosts      []string           `validate:"unique,dive,ip4_addr|hostname|hostname_rfc1123|fqdn"`
	Interfaces []shared.Interface `validate:"dive"`
}

type Logging struct {
	Enabled bool
	Console bool
}

type Config struct {
	Listen                 Listen
	Announcement           Announcement
	Logging                Logging
	Id                     string             `validate:"uuid"`
	GameTitles             []common.GameTitle `validate:"unique,dive,supported_game_title"`
	GeneratePlatformUserId bool
}

func Validator() (error, *validator.Validate) {
	return shared.Validator()
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("Id", uuid.NewString())
	v.SetDefault("Listen.Hosts", []string{netip.IPv4Unspecified().String()})
	v.SetDefault("Announcement.Port", common.AnnouncePort)
	v.SetDefault("Announcement.MulticastGroup", net.ParseIP(common.AnnounceMulticastGroup))
	v.SetDefault("GameTitles", common.SupportedGameTitleSlice)
}

func Unmarshal(v *viper.Viper, c *Config) error {
	if err := shared.Unmarshal(v, c); err != nil {
		return err
	}
	return nil
}
