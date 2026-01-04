package goreleaser

import (
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
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

// TODO: Change if other universal binaries are added
func mergeArchivesForUniversalBinaries(archives *[]config.Archive, universalBinaries []config.UniversalBinary) {
	var archivesToMerge []config.Archive
	binariesToMerge := mapset.NewSet[string]()
	for _, ub := range universalBinaries {
		if strings.HasSuffix(ub.ID, "_amd64") {
			binariesToMerge.Add(ub.ID)
		}
	}
	for i := len(*archives) - 1; i >= 0; i-- {
		a := (*archives)[i]
		if mapset.NewSet[string](a.IDs...).IsSubset(binariesToMerge) {
			archivesToMerge = append(archivesToMerge, a)
			*archives = append((*archives)[:i], (*archives)[i+1:]...)
		}
	}
	mergedArchive := config.Archive{
		ID:           "server_darwin",
		IDs:          binariesToMerge.ToSlice(),
		NameTemplate: `{{ .ProjectName }}_server_{{ .RawVersion }}_mac`,
		Formats:      archivesToMerge[0].Formats,
		Files:        archivesToMerge[0].Files,
	}
	*archives = append(*archives, mergedArchive)
}

// TODO: Change if other universal binaries are added
func mergeBuildsForUniversalBinaries(builds *[]config.Build) {
	var buildsToMerge []config.Build
	for i := len(*builds) - 1; i >= 0; i-- {
		b := (*builds)[i]
		if slices.Contains(b.Goos, OSMacOS.Name()) {
			buildsToMerge = append(buildsToMerge, b)
		}

	}
}

func mergeBuildsPerOS(builds *[]config.Build) {
	buildsPerOsMainBinary := make(map[string]map[string]map[string][]config.Build)
	for _, b := range *builds {
		for _, os := range b.Goos {
			if _, ok := buildsPerOsMainBinary[os]; !ok {
				buildsPerOsMainBinary[os] = make(map[string]map[string][]config.Build)
			}
			if _, ok := buildsPerOsMainBinary[os][b.Main]; !ok {
				buildsPerOsMainBinary[os][b.Main] = make(map[string][]config.Build)
			}
			buildsPerOsMainBinary[os][b.Main][b.Binary] = append(buildsPerOsMainBinary[os][b.Main][b.Binary], b)
		}
	}
}

func GenerateConfig(archives ...*Archive) {
	project := config.Project{
		Version: 2,
	}
	for _, a := range archives {
		archiveBuilds := a.Builds()
		project.Builds = append(project.Builds, archiveBuilds...)
		project.Archives = append(project.Archives, a.Archives(archiveBuilds)...)
	}
	//mergeBuildsForUniversalBinaries(&project.Builds)
	project.UniversalBinaries = universalBinaries(project.Builds)
	mergeArchivesForUniversalBinaries(&project.Archives, project.UniversalBinaries)

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
		return
	}
	println(string(marshal))
}
