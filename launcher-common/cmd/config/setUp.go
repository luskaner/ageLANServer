package config

import (
	"net"
	"runtime"

	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
)

type SetupBaseValues struct {
	MapIp            net.IP
	AddLocalCertData []byte
}

func (v *SetupBaseValues) MapIpRef() *net.IP {
	return &v.MapIp
}

func (v *SetupBaseValues) AddLocalCertDataRef() *[]byte {
	return &v.AddLocalCertData
}

type SetupValues struct {
	*CommonBaseValues
	*SetupBaseValues
	AddUserCertData []byte
	AddCACertData   []byte
	AgentStart      bool
	AgentEndOnError bool
}

func newSetupValues() SetupValues {
	return SetupValues{
		CommonBaseValues: NewCommonBaseValues(),
		SetupBaseValues:  &SetupBaseValues{},
	}
}

func (v *SetupValues) AgentStartRef() *bool {
	return &v.AgentStart
}

func (v *SetupValues) AgentEndOnErrorRef() *bool {
	return &v.AgentEndOnError
}

func (v *SetupValues) AddUserCertDataRef() *[]byte {
	return &v.AddUserCertData
}

func (v *SetupValues) AddCACertDataRef() *[]byte {
	return &v.AddCACertData
}

func initSetUp(flags *pflag.FlagSet) (values *SetupBaseValues) {
	values = &SetupBaseValues{}
	flags.IPVarP(
		values.MapIpRef(),
		"ip",
		"i",
		nil,
		"IP to resolve in local DNS server.",
	)
	flags.BytesBase64VarP(
		values.AddLocalCertDataRef(),
		"localCert",
		"l",
		nil,
		"Add the certificate to the local machine's trusted root store",
	)
	return values
}

func RegularSetUpFlagSet() (values *SetupValues, flags *pflag.FlagSet) {
	values = new(newSetupValues())
	flags = pflag.NewFlagSet("setup", pflag.ContinueOnError)
	values.SetupBaseValues = initSetUp(flags)
	values.CommonBaseValues = AddCommonFlags(
		flags,
		"Only relevant when using 'ip' option. If empty, it will use the system path",
		"It requires the 'localCert' option to be set. If non-empty the certificate will be saved only to the specified path.",
		"Backup metadata. Unnecessary for AoE:DE",
		"Backup profiles.",
	)
	if runtime.GOOS != "linux" {
		flags.BytesBase64VarP(values.AddUserCertDataRef(), "userCert", "u", nil, "Add the certificate to the user's trusted root store")
	}
	flags.BytesBase64VarP(values.AddCACertDataRef(), "caStoreCert", "s", nil, "Add the certificate to the game's trusted root store. For all except AoE I: DE and AoE IV: AE.")
	flags.BoolVarP(values.AgentStartRef(), "agentStart", "g", false, "Start the 'config-admin-agent' if it is not running, we are not admin and is needed for admin action.")
	flags.BoolVarP(values.AgentEndOnErrorRef(), "agentEndOnError", "r", false, "Stop the 'config-admin-agent' if it is running and any admin action failed.")
	_ = flags.MarkHidden("agentStart")
	_ = flags.MarkHidden("agentEndOnError")
	return
}

type AdminSetupValues struct {
	*SetupBaseValues
	*commonCmd.GameIdValues
	*commonCmd.LogRootValues
}

func newAdminSetupValues() AdminSetupValues {
	return AdminSetupValues{
		SetupBaseValues: &SetupBaseValues{},
		GameIdValues:    &commonCmd.GameIdValues{},
		LogRootValues:   &commonCmd.LogRootValues{},
	}
}

func AdminSetupFlagSet() (values *AdminSetupValues, flags *pflag.FlagSet) {
	values = new(newAdminSetupValues())
	flags = pflag.NewFlagSet("setup", pflag.ContinueOnError)
	values.SetupBaseValues = initSetUp(flags)
	commonCmd.LogRootCommand(flags, values.LogRootRef())
	commonCmd.GameVarCommand(flags, values.GameIdRef())
	return
}
