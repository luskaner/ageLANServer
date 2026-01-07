package goreleaser

import (
	"os"
	"slices"

	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"gopkg.in/yaml.v3"
)

func universalBinaries(binaries []config.Build) []config.UniversalBinary {
	var result []config.UniversalBinary
	for _, binary := range binaries {
		if slices.Contains(binary.Goos, OSMacOS.Name()) {
			result = append(result, config.UniversalBinary{
				ID:           binary.ID,
				NameTemplate: binary.Binary,
				Replace:      true,
			})
		}
	}
	return result
}

func GenerateConfig(archives ...*Archive) error {
	project := config.Project{
		Version: 2,
	}
	for _, a := range archives {
		archiveBuilds := a.Builds(OSMacOS)
		project.Builds = append(project.Builds, archiveBuilds...)
		project.Archives = append(project.Archives, a.Archives(archiveBuilds)...)
	}
	project.UniversalBinaries = universalBinaries(project.Builds)

	project.Checksum = config.Checksum{
		NameTemplate: `{{ .ProjectName }}_{{ .RawVersion }}_checksums.txt`,
	}
	project.Signs = []config.Sign{
		{
			Artifacts: "checksum",
			Cmd:       "gpg2",
			Args:      []string{"--batch", "-u", "{{ .Env.GPG_FINGERPRINT }}", "--output", "${signature}", "--detach-sign", "${artifact}"},
		},
	}
	project.Release = config.Release{
		Draft:                    true,
		ReplaceExistingDraft:     false,
		UseExistingDraft:         true,
		ReplaceExistingArtifacts: true,
		Prerelease:               `auto`,
	}
	marshal, err := yaml.Marshal(&project)
	if err != nil {
		return err
	}
	return os.WriteFile(".goreleaser.yaml", marshal, 0o644)
}
