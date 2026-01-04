package goreleaser

import mapset "github.com/deckarep/golang-set/v2"

var (
	Arch386   Architecture = X8632{}
	ArchAmd64 Architecture = X8664{}
	ArchArm32 Architecture = Arm32{}
	ArchArm64 Architecture = Arm64{}
)

var (
	OSWindows OperatingSystem = Windows{}
	OSLinux   OperatingSystem = Linux{}
	OSMacOS   OperatingSystem = MacOS{}
)

type OperatingSystem interface {
	Name() string
	FriendlyName() string
	Archs() mapset.Set[Architecture]
}

type Architecture interface {
	Name() string
	InstructionSet() mapset.Set[string]
	FriendlyName() string
}

type X8632 struct{}

func (a X8632) Name() string {
	return "386"
}

func (a X8632) FriendlyName() string {
	return "x86-32"
}

func (a X8632) InstructionSet() mapset.Set[string] {
	return mapset.NewSet[string]("386", "softfloat")
}

type X8664 struct{}

func (a X8664) Name() string {
	return "amd64"
}

func (a X8664) FriendlyName() string {
	return "x86-64"
}

func (a X8664) InstructionSet() mapset.Set[string] {
	return mapset.NewSet[string]("v1", "v2", "v3", "v4")
}

type Arm32 struct{}

func (a Arm32) Name() string {
	return "arm"
}

func (a Arm32) FriendlyName() string {
	return a.Name()
}

func (a Arm32) InstructionSet() mapset.Set[string] {
	return mapset.NewSet[string]("5", "6", "7")
}

type Arm64 struct{}

func (a Arm64) Name() string {
	return "arm64"
}

func (a Arm64) FriendlyName() string {
	return a.Name()
}

func (a Arm64) InstructionSet() mapset.Set[string] {
	base := []string{
		"v8.0", "v8.1", "v8.2", "v8.3", "v8.4", "v8.5", "v8.6", "v8.7", "v8.8", "v8.9",
		"v9.0", "v9.1", "v9.2", "v9.3", "v9.4", "v9.5",
	}
	suffixes := []string{
		"",
		",lse",
		",crypto",
		",lse,crypto",
	}
	set := mapset.NewSet[string]()
	for _, v := range base {
		for _, s := range suffixes {
			set.Add(v + s)
		}
	}
	return set
}

type Windows struct{}

func (w Windows) Name() string {
	return "windows"
}

func (w Windows) FriendlyName() string {
	return "win10"
}

func (w Windows) Archs() mapset.Set[Architecture] {
	return mapset.NewSet[Architecture](Arch386, ArchAmd64, ArchArm64)
}

type Linux struct{}

func (l Linux) Name() string {
	return "linux"
}

func (l Linux) FriendlyName() string {
	return l.Name()
}

func (l Linux) Archs() mapset.Set[Architecture] {
	return mapset.NewSet[Architecture](Arch386, ArchAmd64, ArchArm32, ArchArm64)
}

type MacOS struct{}

func (m MacOS) Name() string {
	return "darwin"
}

func (m MacOS) FriendlyName() string {
	return "mac"
}

func (m MacOS) Archs() mapset.Set[Architecture] {
	return mapset.NewSet[Architecture](ArchAmd64, ArchArm64)
}
