package goreleaser

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/goreleaser/goreleaser/v2/pkg/config"
	"github.com/luskaner/ageLANServer/common"
)

func extChange(ext string) DestinationFn {
	return func(source string) Renders[FileData] {
		base := strings.TrimSuffix(source, filepath.Ext(source))
		if ext == "" {
			return LiteralString[FileData](base)
		}
		return LiteralString[FileData](base + "." + ext)
	}
}

type DestinationFn func(source string) Renders[FileData]
type DestinationsFnMap = map[OperatingSystem][]DestinationFn
type SourceIgnoreFn = map[OperatingSystem]func(path string) bool
type OperatingSystemsArchs = map[OperatingSystem]mapset.Set[Architecture]

type File struct {
	source      string
	destination string
	mode        os.FileMode
	os          *OperatingSystem
}

type FileData struct {
	BaseOS       string
	SrcScriptExt string
	DstScriptExt string
	DstDocExt    string
	Game         string
}

func NewFileData(os OperatingSystem) FileData {
	f := FileData{}
	switch os {
	case OSWindows:
		f.BaseOS = "windows"
		f.SrcScriptExt = "bat"
		f.DstScriptExt = f.SrcScriptExt
		f.DstDocExt = "txt"
	case OSLinux:
		f.BaseOS = "unix"
		f.SrcScriptExt = "sh"
		f.DstScriptExt = f.SrcScriptExt
	case OSMacOS:
		f.BaseOS = "unix"
		f.SrcScriptExt = "sh"
		f.DstScriptExt = "command"
	}
	return f
}

type Archive struct {
	name     string
	files    mapset.Set[File]
	targets  *BinaryTargets
	binaries map[string]*Binary
}

func NewArchive(name string, targets *BinaryTargets) *Archive {
	return &Archive{
		name:     name,
		files:    mapset.NewSet[File](),
		targets:  targets,
		binaries: make(map[string]*Binary),
	}
}

func NewMergedArchive(name string, archives ...*Archive) *Archive {
	mergedOsesArchs := archives[0].targets.Clone()
	for _, a := range archives[1:] {
		osesToDelete := make([]OperatingSystem, 0)
		for osKey, mergedArchs := range *mergedOsesArchs {
			archs, ok := (*a.targets)[osKey]
			if !ok {
				osesToDelete = append(osesToDelete, osKey)
				continue
			}
			archsToDelete := make([]Architecture, 0)
			for archKey, mergedInstSet := range mergedArchs {
				instSet, ok := archs[archKey]
				if !ok {
					archsToDelete = append(archsToDelete, archKey)
					continue
				}
				if mergedInstSet.IsEmpty() && instSet.IsEmpty() {
					continue
				} else if mergedInstSet.IsEmpty() {
					mergedArchs[archKey] = instSet.Clone()
				} else if instSet.IsEmpty() {
					continue
				} else {
					result := mergedInstSet.Intersect(instSet)
					if result.IsEmpty() {
						archsToDelete = append(archsToDelete, archKey)
					} else {
						mergedArchs[archKey] = result
					}
				}
			}
			for _, archKey := range archsToDelete {
				delete(mergedArchs, archKey)
			}
			if len(mergedArchs) == 0 {
				osesToDelete = append(osesToDelete, osKey)
			}
		}
		for _, osKey := range osesToDelete {
			delete(*mergedOsesArchs, osKey)
		}
	}
	oses := mapset.NewSetWithSize[OperatingSystem](len(*mergedOsesArchs))
	for osKey := range *mergedOsesArchs {
		oses.Add(osKey)
	}
	mergedArchive := NewArchive(name, mergedOsesArchs)
	for _, a := range archives {
		for file := range a.files.Iter() {
			if file.os == nil {
				mergedArchive.files.Add(file)
			} else if _, exists := (*mergedOsesArchs)[*file.os]; exists {
				mergedArchive.files.Add(file)
			}
		}
		for path, binary := range a.binaries {
			clonedBinary := binary.CloneForOperatingSystems(oses)
			if clonedBinary != nil {
				mergedArchive.binaries[path] = clonedBinary
			}
		}
	}
	return mergedArchive
}

func (a *Archive) Name() string {
	return a.name
}

func (a *Archive) AddSrcFile(source string) {
	a.files.Add(File{source: source})
}

func (a *Archive) AddSrcDstFile(source string, destination string) {
	a.files.Add(File{source: source, destination: destination})
}

func (a *Archive) AddSrcDstFileWithMode(source string, destination string, mode os.FileMode) {
	a.files.Add(File{source: source, destination: destination, mode: mode})
}

func (a *Archive) addFile(os OperatingSystem, fileMode os.FileMode, fileData FileData, source Renders[FileData], sourceIgnoreFn SourceIgnoreFn, destinationFn ...DestinationFn) {
	var sourceRendered string
	if sourceIgnoreFn != nil && sourceIgnoreFn[os] != nil {
		if sourceRendered = source.Render(fileData); sourceIgnoreFn[os](sourceRendered) {
			return
		}
	} else {
		sourceRendered = source.Render(fileData)
	}
	file := File{}
	file.source = filepath.ToSlash(sourceRendered)
	file.destination = sourceRendered
	for _, destFn := range destinationFn {
		file.destination = destFn(file.destination).Render(fileData)
	}
	file.destination = filepath.ToSlash(file.destination)
	file.os = &os
	if UnixBasedOperatingSystems.ContainsOne(os) {
		file.mode = fileMode
	}
	a.files.Add(file)
}

func (a *Archive) AddSrcOsDstFile(source Renders[FileData], sourceIgnoreFn SourceIgnoreFn, destinationFn DestinationFn, destinationsFn DestinationsFnMap, fileMode os.FileMode, perGame bool) {
	if destinationFn == nil {
		destinationFn = func(source string) Renders[FileData] {
			return LiteralString[FileData](source)
		}
	}
	if destinationsFn == nil {
		destinationsFn = make(DestinationsFnMap)
	}
	destinationsFns := make(map[OperatingSystem][]DestinationFn)
	for oses := range *a.targets {
		destinationsFns[oses] = []DestinationFn{destinationFn}
		if value, exists := destinationsFn[oses]; exists {
			destinationsFns[oses] = append(destinationsFns[oses], value...)
		}
	}
	for operatingSystem, osDestinationFns := range destinationsFns {
		fileData := NewFileData(operatingSystem)
		if perGame {
			for game := range common.SupportedGames.Iter() {
				fileData.Game = game
				a.addFile(operatingSystem, fileMode, fileData, source, sourceIgnoreFn, osDestinationFns...)
			}
		} else {
			a.addFile(operatingSystem, fileMode, fileData, source, sourceIgnoreFn, osDestinationFns...)
		}
	}
}

func defaultDest(source string) string {
	src := filepath.ToSlash(filepath.Clean(source))
	parts := strings.SplitN(src, "/", 2)
	if len(parts) == 2 {
		return parts[1]
	}
	return src
}

func (a *Archive) AddScriptFiles(destDir string, source Renders[FileData], sourceIgnoreFn SourceIgnoreFn, destinationsFn DestinationsFnMap, perGame bool) {
	finalDestinationsFn := destinationsFn
	if finalDestinationsFn == nil {
		finalDestinationsFn = make(DestinationsFnMap)
	}
	if _, exists := finalDestinationsFn[OSMacOS]; !exists {
		finalDestinationsFn[OSMacOS] = []DestinationFn{}
	}
	finalDestinationsFn[OSMacOS] = append([]DestinationFn{extChange("command")}, finalDestinationsFn[OSMacOS]...)
	a.AddSrcOsDstFile(
		source,
		sourceIgnoreFn,
		func(source string) Renders[FileData] {
			return LiteralString[FileData](filepath.Join(destDir, filepath.Base(source)))
		},
		finalDestinationsFn,
		0744,
		perGame,
	)
}

func (a *Archive) AddConfigFiles(destDir string, source Renders[FileData], perGame bool) {
	a.AddSrcOsDstFile(
		source,
		nil,
		func(source string) Renders[FileData] {
			return NewTemplate[FileData](
				filepath.Join(
					destDir,
					strings.ReplaceAll(defaultDest(source), `game`, `{{.Game}}`),
				),
			)
		},
		nil,
		0,
		perGame,
	)
}

func (a *Archive) AddDocFiles(destDir string, destinationFn DestinationFn, destinationsFn DestinationsFnMap, sources ...string) {
	finalDestinationsFn := destinationsFn
	if finalDestinationsFn == nil {
		finalDestinationsFn = make(DestinationsFnMap)
	}
	if _, exists := finalDestinationsFn[OSWindows]; !exists {
		finalDestinationsFn[OSWindows] = []DestinationFn{}
	}
	if _, exists := finalDestinationsFn[OSMacOS]; !exists {
		finalDestinationsFn[OSMacOS] = []DestinationFn{}
	}
	if _, exists := finalDestinationsFn[OSLinux]; !exists {
		finalDestinationsFn[OSLinux] = []DestinationFn{}
	}
	if destinationFn == nil {
		destinationFn = func(source string) Renders[FileData] {
			return LiteralString[FileData](source)
		}
	}
	finalDestinationsFn[OSWindows] = append([]DestinationFn{extChange("txt")}, finalDestinationsFn[OSWindows]...)
	finalDestinationsFn[OSMacOS] = append([]DestinationFn{extChange("")}, finalDestinationsFn[OSMacOS]...)
	finalDestinationsFn[OSLinux] = append([]DestinationFn{extChange("")}, finalDestinationsFn[OSLinux]...)
	for _, source := range sources {
		a.AddSrcOsDstFile(
			LiteralString[FileData](source),
			nil,
			func(source string) Renders[FileData] {
				return destinationFn(filepath.Join(destDir, defaultDest(source)))
			},
			finalDestinationsFn,
			0,
			false,
		)
	}
}

func (a *Archive) AddMainBinary(binary *Binary) {
	a.binaries[filepath.Base(binary.main)] = binary
}

func (a *Archive) AddAuxiliarBinary(binary *Binary) {
	a.binaries[filepath.ToSlash(filepath.Join("bin", binary.main))] = binary
}

func (a *Archive) CloneWithFilesPrefix(prefix string) *Archive {
	newArchive := NewArchive(a.name, a.targets)
	for file := range a.files.Iter() {
		newFile := file
		newFile.destination = filepath.ToSlash(filepath.Join(prefix, file.destination))
		newArchive.files.Add(newFile)
	}
	for path, binary := range a.binaries {
		newArchive.binaries[filepath.ToSlash(filepath.Join(prefix, path))] = binary
	}
	return newArchive
}

func archToValues(architecture Architecture, b config.Build) mapset.Set[string] {
	res := mapset.NewSet[string]()
	switch architecture {
	case Arch386:
		res.Append(b.Go386...)
	case ArchAmd64:
		res.Append(b.Goamd64...)
	case ArchArm32:
		res.Append(b.Goarm...)
	case ArchArm64:
		res.Append(b.Goarm64...)
	}
	return res
}

func build(path, main string, operatingSystem OperatingSystem, architecture Architecture, instructionSet string) config.Build {
	b := config.Build{
		Main:   main,
		Binary: path,
		Goos:   []string{operatingSystem.Name()},
		Goarch: []string{architecture.Name()},
	}
	id := fmt.Sprintf("%s_%s_%s", path, operatingSystem.Name(), architecture.Name())
	if instructionSet != "" {
		id = fmt.Sprintf("%s_%s", id, instructionSet)
		switch architecture {
		case Arch386:
			b.Go386 = []string{instructionSet}
		case ArchAmd64:
			b.Goamd64 = []string{instructionSet}
		case ArchArm32:
			b.Goarm = []string{instructionSet}
		case ArchArm64:
			b.Goarm64 = []string{instructionSet}
		}
	}
	b.ID = id
	return b
}

func (a *Archive) Builds() []config.Build {
	var builds []config.Build
	for path, binary := range a.binaries {
		for operatingSystem, architectures := range *binary.targets {
			for architecture, instructionSets := range architectures {
				if instructionSets.Cardinality() > 0 {
					for intructionSet := range instructionSets.Iter() {
						builds = append(builds, build(path, binary.main, operatingSystem, architecture, intructionSet))
					}
				} else {
					builds = append(builds, build(path, binary.main, operatingSystem, architecture, ""))
				}
			}
		}
	}
	return builds
}

func archive(name string, operatingSystem OperatingSystem, architecture Architecture, instructionSet string, allBuilds []config.Build, files mapset.Set[File]) *config.Archive {
	matchingBuildIds := mapset.NewSet[string]()
	for _, b := range allBuilds {
		if slices.Contains(b.Goos, operatingSystem.Name()) && slices.Contains(b.Goarch, architecture.Name()) {
			if instructionSet == "" {
				matchingBuildIds.Add(b.ID)
			} else if archValues := archToValues(architecture, b); archValues.Contains(instructionSet) {
				matchingBuildIds.Add(b.ID)
			}
		}
	}
	if matchingBuildIds.IsEmpty() {
		return nil
	}
	id := fmt.Sprintf("%s_%s_%s", name, operatingSystem.Name(), architecture.Name())
	nameTemplate := fmt.Sprintf(`{{ .ProjectName }}_%s_{{ .RawVersion }}_%s_%s`, name, operatingSystem.FriendlyName(), architecture.FriendlyName())
	if instructionSet != "" {
		id = fmt.Sprintf("%s-%s", id, instructionSet)
		nameTemplate = fmt.Sprintf(`%s-%s`, nameTemplate, instructionSet)
	}
	formats := mapset.NewSet[string]()
	if operatingSystem == OSWindows {
		formats.Add("zip")
	} else {
		formats.Add("tar.gz")
	}
	a := &config.Archive{
		IDs:          matchingBuildIds.ToSlice(),
		ID:           id,
		NameTemplate: nameTemplate,
		Formats:      formats.ToSlice(),
	}
	for file := range files.Iter() {
		if file.os == nil || *file.os == operatingSystem {
			f := config.File{
				Source:      file.source,
				Destination: file.destination,
			}
			if file.mode != 0 {
				f.Info.Mode = file.mode
			}
			a.Files = append(a.Files, f)
		}
	}
	return a
}

func (a *Archive) Archives(builds []config.Build) []config.Archive {
	var archives []config.Archive
	for operatingSystem, architectures := range *a.targets {
		for architecture, instructionSets := range architectures {
			if instructionSets.Cardinality() == 0 {
				if currentArchive := archive(a.name, operatingSystem, architecture, "", builds, a.files); currentArchive != nil {
					archives = append(archives, *currentArchive)
				}
			} else {
				for intructionSet := range instructionSets.Iter() {
					if currentArchive := archive(a.name, operatingSystem, architecture, intructionSet, builds, a.files); currentArchive != nil {
						archives = append(archives, *currentArchive)
					}
				}
			}
		}
	}
	return archives
}

func (a *Archive) RemoveFiles(sources ...string) {
	for _, file := range a.files.ToSlice() {
		if slices.Contains(sources, file.source) {
			a.files.Remove(file)
		}
	}
}
