package launcher

import (
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/go-playground/validator/v10"
	"github.com/luskaner/ageLANServer/common"
	"github.com/luskaner/ageLANServer/common/config/shared"
	"github.com/spf13/viper"
	"net/netip"
	"runtime"
	"slices"
)

type RebroadcastBattleServer struct {
	/*
		Disable indicates whether the BattleServer announcements will be not rebroadcasted.
		If both IPAddrs and Interfaces are empty, every IPAddr/interface will be selected.
	*/
	Disable bool
	// IPAddrs (v4) to rebroadcast the BattleServer announcements to.
	IPAddrs shared.MapsetWrapper[netip.Addr] `validate:"dive,ip_addr_v4"`
	// Network Interfaces to rebroadcast the BattleServer announcements to. Ignored if using IPAddrs.
	Interfaces shared.MapsetWrapper[shared.Interface]
}

type ServerQuery struct {
	// IPv4 configuration for server queries.
	IPv4 ServerQueryIPv4
	// IPv6 configuration for server queries.
	IPv6 ServerQueryIPv6
}

type ServerQueryIPv4 struct {
	// Ports to query for servers. If left empty, it defaults to [31978].
	Ports shared.MapsetWrapper[uint16]
	// MulticastGroups to query for servers. If left empty, it defaults to ["239.31.97.8"].
	MulticastGroups shared.MapsetWrapper[netip.Addr] `validate:"dive,ip_addr_v4,ip_multicast"`
	// Whether to avoid sending broadcast to discover the servers. Some servers might only work via it.
	DisableBroadcast bool
}

type ServerQueryIPv6 struct {
	// Ports to query for servers. If left empty, it defaults to [31979].
	Ports shared.MapsetWrapper[uint16]
	// MulticastGroups to query for servers. If left empty, it defaults to ["FF05::EF1F:6108"].
	MulticastGroups shared.MapsetWrapper[netip.Addr] `validate:"dive,ip_addr_v6,ip_multicast"`
	// Similar to broadcast in IPv4, If DisableLinkLocal is false, it will join the "FF02::1" group.
	DisableLinkLocal bool
}

type ServerConnect struct {
	// Host is the IPv4/IPv6 address or hostname of the server to connect to.
	Host string `validate:"host_omit|ip_addr|ip_addr_no_zone|hostname|hostname_rfc1123|fqdn"`
}

type ServerRun struct {
	// NoStop indicates if the server will not be stopped when the game exits (and shown in a window).
	NoStop bool
	/*
		Command to start the server.
		The --id and --gameTitles arguments will always be overriden even if passed as arguments.
		Run the server manually for full control.

		If left empty, or the first element is, it will try to find it in the following locations:
		1. "./server/"
		2. "../"
		3. "../server/"

		Note 1: Only servers v1.9.0 or higher are supported (recommended to use the same version as the launcher).
		Note 2: You may use environment variables.
	*/
	Command []string
}

type Server struct {
	IPProtocol common.IPProtocol
	// Mode indicates whether to connect, run or query the server.
	Mode common.LauncherServerMode
	// Connect configuration to the server. Applies when Mode is common.ModeConnect.
	Connect ServerConnect `validate:"required_if=Mode connect"`
	// Run configuration to for the server. Applies when Mode is common.ModeQueryOrRun or common.ModeRun.
	Run ServerRun `validate:"required_if=Mode run|required_if=Mode ''"`
	// Query configuration for servers. Applies when Mode is common.ModeQueryOrRun.
	Query ServerQuery `validate:"required_if=Mode ''"`
}

type Client struct {
	// GameTitle to launch.
	GameTitle common.GameTitle `validate:"required"`
	/*
		Type of Launcher.
		The possible values are:
		- "steam": The game will be launched using Steam.
		- "msstore": The game will be launched using the Microsoft Store (Windows only).
		- "path": The game will be launched using the PathCommand.

		If left empty (default), it will try to find the first installed launcher in the following order:
		1. Steam.
		2. Microsoft Store (Windows only).
	*/
	Launcher common.ClientLauncher
	/*
		PathCommand to use to start the game if Launcher is "path".
		The following variables are available (except first element) and then required to be used:
		- {HostFilePath} will be replaced by the host file path when Config.StoreToAddHost is common.StoreTmp.
		- {CertFilePath} will be replaced by the certificate file path when Config.StoreToAddCertificate is common.StoreTmp.
		Note: You may use environment variables.
	*/
	PathCommand             []string `validate:"required_if=Launcher path,required_vars"`
	Isolation               Isolation
	RebroadcastBattleServer RebroadcastBattleServer
}

type Isolation struct {
	/*
			WindowsUserProfilePath must point to the equivalent to %USERPROFILE% on Windows when running on Linux when
		    Config.Launcher is 'path', if that is 'steam' and this is empty then this resolves to;
			{ProtonPath}/steamapps/compatdata/{AppId}/pfx/drive_c/users/steamuser where:
			- {ProtonPath} is specified in 'dirs' as defined in /launcher-common/steam/steam_linux.go
			- {AppId} is the 'AppId' of the game as defined in /launcher-common/steam/steam.go
			Note: You may use environment variables.
	*/
	WindowsUserProfilePath string `validate:"required_on_linux_with_path_no_aoe1"`
	/*
		NoMetadata disable the backup/restore of the metadata folder.
		Requires WindowsUserProfilePath on Linux when Config.Launcher is 'path'.
	*/
	NoMetadata bool
}

type Config struct {
	// Debug enables extra logging.
	Debug bool
	/*
		StoreToAddHost indicates the type of store to add the host mapping to if necessary:
		- common.StoreLocal for the system-wide store (needs admin).
		- common.StoreTmp for a temporary file. It follows the Windows format even on Linux.
		If common.StoreNone, it will not add the hosts.
	*/
	StoreToAddHost common.LauncherStore
	/*
		StoreToAddCertificate indicates the type of store to add the certificate to if necessary:
		- common.StoreUser for the current user store. Only available on Windows.
		- common.StoreLocal for the system-wide store (needs admin).
		- common.StoreTmp for a temporary file.
		If common.StoreNone, it will not add the certificate.
	*/
	StoreToAddCertificate common.LauncherUserStore
	/*
		SetupCommand is the executable to run (including arguments) before doing any configuration.
		The command must return a 0 exit code to continue. If you need to keep it running spawn a new detached process.
		If empty (default), nothing will be ran.
		Windows: Path names need to use double backslashes within single quotes or be within double quotes.
		Note: You may use environment variables.
	*/
	SetupCommand []string
	/*
		RevertCommand is the executable to run (including arguments) to run after:
		- SetupCommand has been run.
		- Game has exited and everything has been reverted. It may run before if there is an error.
		If empty (default), nothing will be ran.
		Windows: Path names need to use double backslashes within single quotes or be within double quotes.
		Note: You may use environment variables.
	*/
	RevertCommand []string
	// Isolation configuration to avoid issues when using the official launcher.
	Client Client
	Server Server
}

func requiredOnLinuxPathNoAoE1(fl validator.FieldLevel) bool {
	if runtime.GOOS != "linux" {
		return true
	}
	config := fl.Top().Interface().(Config)
	if config.Client.GameTitle == common.AoE1 {
		return true
	}
	if config.Client.Launcher != common.ClientLauncherPath {
		return true
	}
	if fl.Parent().Interface().(Isolation).NoMetadata {
		return true
	}
	return len(fl.Field().String()) > 0
}

func hostOmit(fl validator.FieldLevel) bool {
	if fl.Field().String() != "" {
		return false
	}
	config := fl.Top().Interface().(*Config)
	return config.Server.Mode != common.ModeConnect
}

func requiredVars(fl validator.FieldLevel) bool {
	config := fl.Top().Interface().(*Config)
	var foundHost bool
	var foundCert bool
	if !config.StoreToAddCertificate.IsTmp() {
		foundCert = true
	}
	if config.StoreToAddHost != common.StoreTmp {
		foundHost = true
	}
	if foundHost && foundCert {
		return true
	}
	command := fl.Field().Interface().([]string)
	if len(command) < 2 {
		return false
	}
	foundHost = foundHost || slices.Contains(command[1:], "{HostFilePath}")
	foundCert = foundCert || slices.Contains(command[1:], "{CertFilePath}")
	return foundHost && foundCert
}

func Validator() (error, *validator.Validate) {
	err, validate := shared.Validator()
	if err != nil {
		return err, nil
	}
	if err = validate.RegisterValidation("required_on_linux_with_path_no_aoe1", requiredOnLinuxPathNoAoE1); err != nil {
		return err, nil
	}
	if err = validate.RegisterValidation("host_omit", hostOmit); err != nil {
		return err, nil
	}
	if err = validate.RegisterValidation("required_vars", requiredVars); err != nil {
		return err, nil
	}
	return nil, validate
}

func SetDefaults(v *viper.Viper) {
	v.SetDefault("StoreToAddHost", "local")
	v.SetDefault("StoreToAddCertificate", "local")
	v.SetDefault("Server.IPProtocol", common.IPv4)
	v.SetDefault("RebroadcastBattleServer.Interfaces", shared.MapsetWrapper[shared.Interface]{
		Set: mapset.NewThreadUnsafeSet[shared.Interface](),
	})
	v.SetDefault("RebroadcastBattleServer.IPAddrs", shared.MapsetWrapper[netip.Addr]{
		Set: mapset.NewThreadUnsafeSet[netip.Addr](),
	})
	v.SetDefault("Server.Query.IPv4.Ports", shared.MapsetWrapper[uint16]{
		Set: mapset.NewThreadUnsafeSet[uint16](common.AnnouncePortIPv4),
	})
	ipV4MulticastGroup, _ := netip.ParseAddr(common.AnnounceMulticastGroupIPv4)
	v.SetDefault("Server.Query.IPv4.MulticastGroups", shared.MapsetWrapper[netip.Addr]{
		Set: mapset.NewThreadUnsafeSet(ipV4MulticastGroup),
	})
	v.SetDefault("Server.Query.IPv6.Ports", shared.MapsetWrapper[uint16]{
		Set: mapset.NewThreadUnsafeSet[uint16](common.AnnouncePortIPv6),
	})
	ipV6MulticastGroup, _ := netip.ParseAddr(common.AnnounceMulticastGroupIPv6)
	v.SetDefault("Server.Query.IPv6.MulticastGroups", shared.MapsetWrapper[netip.Addr]{
		Set: mapset.NewThreadUnsafeSet(ipV6MulticastGroup),
	})
}

func Unmarshal(v *viper.Viper, c *Config) error {
	if err := shared.Unmarshal(v, c); err != nil {
		return err
	}
	return nil
}
