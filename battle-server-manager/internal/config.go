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

type Configuration struct {
	Region     string
	Name       string
	Host       string
	CertsPath  string
	Executable Executable
	Ports      Ports
}
