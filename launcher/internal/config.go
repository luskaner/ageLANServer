package internal

type Executable struct {
	Path string   `mapstructure:"Executable"`
	Args []string `mapstructure:"ExecutableArgs"`
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
	Executable `mapstructure:",squash"`
	Run        string
}

type Server struct {
	Executable              `mapstructure:",squash"`
	Start                   string
	Host                    string
	Stop                    string
	SingleAutoSelect        bool
	AnnouncePorts           []int
	AnnounceMulticastGroups []string
	BattleServerManager     BattleServerManager
}

type Client struct {
	Executable `mapstructure:",squash"`
	Path       string
}

type Configuration struct {
	Config Config
	Server Server
	Client Client
}
