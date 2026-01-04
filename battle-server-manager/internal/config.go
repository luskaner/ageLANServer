package internal

type Executable struct {
	Path      string
	ExtraArgs []string
}

type Ports struct {
	Bs        int
	WebSocket int
	OutOfBand int
}

type SSL struct {
	Auto     bool
	CertFile string
	KeyFile  string
}

type Configuration struct {
	Region     string
	Name       string
	Host       string
	Executable Executable
	Ports      Ports
	SSL        SSL
}
