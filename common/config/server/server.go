package server

import (
	"net/netip"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/config/shared"
	"github.com/spf13/viper"
)

type Network struct {
	IPProtocol common.IPProtocol
	// HTTPS server Listen configuration.
	Listen Listen
	// Announcement configuration for server queries.
	Announcement Announcement
}

type AnnouncementIPv4 struct {
	// The Port to listen for server queries in IPv4.
	Port uint16
	// The MulticastGroup to listen for server queries in IPv4.
	MulticastGroup netip.Addr `validate:"ip_addr_v4,ip_multicast"`
}

type AnnouncementIPv6 struct {
	// The Port to listen for server queries in IPv6.
	Port uint16
	// The MulticastGroups to listen for server queries in IPv6.
	MulticastGroup netip.Addr `validate:"ip_addr_v6,ip_multicast"`
	// Similar to broadcast in IPv4, If DisableLinkLocal is false, the server will subscribe to "FF02::1" group.
	DisableLinkLocal bool
}

type Announcement struct {
	// Disabled indicates that will not listen for server queries.
	Disabled bool
	/*
		IPv4 indicates whether the server will listen for server queries in IPv4.
		The hosts being the ones in Listen.Hosts or Listen.Interfaces that are IPv4.
	*/
	IPv4 AnnouncementIPv4
	/*
		IPv6 indicates whether the server will listen for server queries in IPv6.
		The hosts being the ones in Listen.Hosts or Listen.Interfaces that are IPv6.
	*/
	IPv6 AnnouncementIPv6
}

type Hosts struct {
	// UseOnlyFirstResolvedIP indicates whether to only use the first resolved IPAddr address of each host and IPv4/IPv6.
	UseOnlyFirstResolvedIP bool
	// Values to listen to (IPAddr addresses, hostnames and domains). Mutually exclusive with Interfaces.
	Values shared.MapsetWrapper[string] `validate:"dive,ip_addr|ip_addr_no_zone|hostname|hostname_rfc1123|fqdn"`
}

type Listen struct {
	// DisableHttps indicates whether the server will only listen to HTTP instead.
	DisableHttps bool
	// Port is the port number to listen to.
	Port uint16
	// Hosts to listen to.
	Hosts Hosts
	// Network Interfaces to listen to. Mutually exclusive with Hosts.
	Interfaces shared.MapsetWrapper[shared.Interface] `validate:"dive"`
}

type Logging struct {
	// Enabled indicates whether logging for access requests is enabled.
	Enabled bool
	// Console indicates whether logs should be printed to the console instead of a file.
	Console bool
}

type Config struct {
	// Id to identify the server instance.
	Id uuid.UUID
	// GameTitles to support in the server.
	GameTitles shared.MapsetWrapper[common.GameTitle] `validate:"dive,supported_game_title"`
	// Network configuration for the server to listen to HTTPs requests and respond to discovery queries.
	Network Network `validate:"unique_ports"`
	// Logging configuration for access requests.
	Logging Logging
	// GeneratePlatformUserId indicates whether the server should generate a platform user ID for each player.
	GeneratePlatformUserId bool
}

func uniquePorts(fl validator.FieldLevel) bool {
	config := fl.Field().Interface().(Network)
	announcement := config.Announcement
	return config.Listen.Port != announcement.IPv4.Port && announcement.IPv4.Port != announcement.IPv6.Port
}

func Validator() (error, *validator.Validate) {
	err, validate := shared.Validator()
	if err != nil {
		return err, nil
	}
	if err = validate.RegisterValidation("unique_ports", uniquePorts); err != nil {
		return err, nil
	}
	return nil, validate
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("Id", uuid.NewString())
	v.SetDefault("Network.Listen.Hosts.Values", shared.MapsetWrapper[string]{
		Set: mapset.NewThreadUnsafeSet[string](),
	})
	v.SetDefault("Network.Listen.Hosts.Values", shared.MapsetWrapper[shared.Interface]{
		Set: mapset.NewThreadUnsafeSet[shared.Interface](),
	})
	v.SetDefault("Network.IPProtocol", common.IPv4)
	v.SetDefault("Network.Announcement.IPv4.Port", uint16(common.AnnouncePortIPv4))
	ipV4MulticastGroup, _ := netip.ParseAddr(common.AnnounceMulticastGroupIPv4)
	v.SetDefault("Network.Announcement.IPv4.MulticastGroup", ipV4MulticastGroup)
	v.SetDefault("Network.Announcement.IPv6.Port", uint16(common.AnnouncePortIPv6))
	ipV6MulticastGroup, _ := netip.ParseAddr(common.AnnounceMulticastGroupIPv6)
	v.SetDefault("Network.Announcement.IPv6.MulticastGroup", ipV6MulticastGroup)
	v.SetDefault("GameTitles", common.SupportedGameTitles)
}

func Unmarshal(v *viper.Viper, c *Config) error {
	if err := shared.Unmarshal(v, c); err != nil {
		return err
	}
	return nil
}
