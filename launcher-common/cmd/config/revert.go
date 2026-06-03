package config

import (
	"runtime"

	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
)

type RevertBaseValues struct {
	UnmapIPs        bool
	RemoveLocalCert bool
	RemoveAll       bool
}

func (v *RevertBaseValues) UnmapIPsRef() *bool {
	return &v.UnmapIPs
}

func (v *RevertBaseValues) RemoveAllRef() *bool {
	return &v.RemoveAll
}

func (v *RevertBaseValues) RemoveLocalCertRef() *bool {
	return &v.RemoveLocalCert
}

func InitBaseRevert(flags *pflag.FlagSet) (values *RevertBaseValues) {
	values = &RevertBaseValues{}
	flags.BoolVarP(
		values.UnmapIPsRef(),
		"ip",
		"i",
		false,
		"Remove the IP mappings from the local DNS server",
	)
	flags.BoolVarP(
		values.RemoveLocalCertRef(),
		"localCert",
		"l",
		false,
		"Remove the certificate from the local machine's trusted root store",
	)
	flags.BoolVarP(
		values.RemoveAllRef(),
		"all",
		"a",
		false,
		"Removes all configuration. Equivalent to the rest of the flags being set without fail-fast.",
	)
	return values
}

type RevertValues struct {
	*CommonBaseValues
	*RevertBaseValues
	RemoveUserCert     bool
	RestoreCAStoreCert bool
	StopAgent          bool
}

func NewRevertValues() RevertValues {
	return RevertValues{
		CommonBaseValues: NewCommonBaseValues(),
		RevertBaseValues: &RevertBaseValues{},
	}
}

func (v *RevertValues) RemoveUserCertRef() *bool {
	return &v.RemoveUserCert
}

func (v *RevertValues) RestoreCAStoreCertRef() *bool {
	return &v.RestoreCAStoreCert
}

func (v *RevertValues) StopAgentRef() *bool {
	return &v.StopAgent
}

func RegularRevertFlagSet() (values *RevertValues, flags *pflag.FlagSet) {
	values = new(NewRevertValues())
	flags = pflag.NewFlagSet("revert", pflag.ContinueOnError)
	values.RevertBaseValues = InitBaseRevert(flags)
	values.CommonBaseValues = AddCommonFlags(
		flags,
		"",
		"",
		"Restore metadata. Unnecessary for AoE:DE",
		"Restore profiles.",
	)
	if runtime.GOOS != "linux" {
		flags.BoolVarP(values.RemoveUserCertRef(), "userCert", "u", false, "Remove the certificate from the user's trusted root store")
	}
	flags.BoolVarP(values.RestoreCAStoreCertRef(), "caStoreCert", "s", false, "Restore the game's trusted root store. For all except AoE I: DE and AoE IV: AE.")
	flags.BoolVarP(values.StopAgentRef(), "stopAgent", "g", false, "Stop the 'config-admin-agent' if it is running after all operations")
	_ = flags.MarkHidden("stopAgent")
	return
}

type AdminRevertValues struct {
	*RevertBaseValues
	*commonCmd.LogRootValues
}

func newAdminRevertValues() AdminRevertValues {
	return AdminRevertValues{
		RevertBaseValues: &RevertBaseValues{},
		LogRootValues:    &commonCmd.LogRootValues{},
	}
}

func AdminRevertFlagSet() (values *AdminRevertValues, flags *pflag.FlagSet) {
	values = new(newAdminRevertValues())
	flags = pflag.NewFlagSet("revert", pflag.ContinueOnError)
	values.RevertBaseValues = InitBaseRevert(flags)
	commonCmd.LogRootCommand(flags, values.LogRootRef())
	return
}
