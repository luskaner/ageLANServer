package goreleaser

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"gopkg.in/yaml.v3"
)

func universalBinaries(binaries []config.Build) []config.UniversalBinary {
	var result []config.UniversalBinary
	for _, binary := range binaries {
		if slices.Contains(binary.Goos, OSMacOS.Goos()) {
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
	main := "./config-helper"
	for target, archs := range *Targets64Windows {
		for arch := range archs {
			binary := filepath.Join("bin", main)
			id := fmt.Sprintf("%s_%s_%s", binary, target.Goos(), arch.Goarch())
			project.Builds = append(
				project.Builds,
				config.Build{
					ID:     id,
					Goos:   []string{target.Goos()},
					Goarch: []string{arch.Goarch()},
					Main:   main,
					Binary: binary,
				},
			)
		}
	}
	for i, arch := range project.Archives {
		if !strings.Contains(arch.ID, "linux") {
			continue
		}
		amd64 := strings.Contains(arch.ID, "amd64")
		arm64 := strings.Contains(arch.ID, "arm64")
		var name string
		if amd64 {
			name = "dist/bin/config-helper_windows_amd64_windows_amd64_v1"
		} else if arm64 {
			name = "dist/bin/config-helper_windows_arm64_windows_arm64_v8.0"
		} else {
			continue
		}
		var targets []string
		if strings.Contains(arch.ID, "launcher") || strings.Contains(arch.ID, "battle-server-manager") {
			targets = []string{"."}
		} else if strings.Contains(arch.ID, "full") {
			targets = []string{"./launcher", "./battle-server-manager"}
		} else {
			continue
		}
		for _, target := range targets {
			project.Archives[i].Files = append(project.Archives[i].Files, config.File{
				Source:      name,
				Destination: target,
			})
		}
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
