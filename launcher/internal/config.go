package internal

type executable struct {
	Executable     string
	ExecutableArgs []string
}

type Config struct {
	CanAddHost               bool
	CanTrustCertificate      string
	CanBroadcastBattleServer string
	Log                      bool
	IsolateMetadata          string
	IsolateProfiles          string
	SetupCommand             []string
	RevertCommand            []string
}

type BattleServerManager struct {
	executable `mapstructure:",squash"`
	Run        string
}

type Server struct {
	executable              `mapstructure:",squash"`
	Start                   string
	Host                    string
	Stop                    string
	SingleAutoSelect        bool
	AnnouncePorts           []int
	AnnounceMulticastGroups []string
	BattleServerManager     BattleServerManager
}

type Client struct {
	executable `mapstructure:",squash"`
	Path       string
}

type Configuration struct {
	Config Config
	Server Server
	Client Client
}
