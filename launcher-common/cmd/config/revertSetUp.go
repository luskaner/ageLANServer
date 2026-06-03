package config

import (
	"github.com/luskaner/ageLANServer/common/cmd"
	"github.com/spf13/pflag"
)

type CommonBaseValues struct {
	*cmd.GameIdValues
	*cmd.LogRootValues
	GamePath     string
	DataPath     string
	HostFilePath string
	CertFilePath string
	Metadata     bool
	Profiles     bool
}

func NewCommonBaseValues() *CommonBaseValues {
	return &CommonBaseValues{
		GameIdValues:  &cmd.GameIdValues{},
		LogRootValues: &cmd.LogRootValues{},
	}
}

func (v *CommonBaseValues) GamePathRef() *string {
	return &v.GamePath
}

func (v *CommonBaseValues) DataPathRef() *string {
	return &v.DataPath
}

func (v *CommonBaseValues) HostFilePathRef() *string {
	return &v.HostFilePath
}

func (v *CommonBaseValues) CertFilePathRef() *string {
	return &v.CertFilePath
}

func (v *CommonBaseValues) MetadataRef() *bool {
	return &v.Metadata
}

func (v *CommonBaseValues) ProfilesRef() *bool {
	return &v.Profiles
}

func AddCommonFlags(flags *pflag.FlagSet, hostFilePathDescSuffix string, certFilePathDescSuffix string, metadataDesc string, profilesDesc string) (values *CommonBaseValues) {
	values = NewCommonBaseValues()
	flags.StringVar(
		values.GamePathRef(),
		"gamePath",
		"",
		"Path to the game folder. Required when using 'caStoreCert' for all except AoE: DE and AoE IV: AE.",
	)
	flags.StringVar(
		values.DataPathRef(),
		"dataPath",
		"",
		"Path to the game user data. Required when using isolation.",
	)
	flags.StringVarP(values.HostFilePathRef(), "hostFilePath", "o", "", "Path to the host file. "+hostFilePathDescSuffix)
	flags.StringVarP(values.CertFilePathRef(), "certFilePath", "t", "", "Path to the certificate file. "+certFilePathDescSuffix)
	flags.BoolVarP(values.MetadataRef(), "metadata", "m", false, metadataDesc)
	flags.BoolVarP(values.ProfilesRef(), "profiles", "p", false, profilesDesc)
	cmd.LogRootCommand(flags, values.LogRootRef())
	cmd.GameVarCommand(flags, values.GameIdRef())
	return values
}
