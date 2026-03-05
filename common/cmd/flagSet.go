package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"sort"

	commonLogger "github.com/luskaner/ageLANServer/common/logger"
	"github.com/spf13/pflag"
)

type commandHandler func(args []string) error
type RootFlagSet struct {
	fs       *pflag.FlagSet
	commands map[string]commandHandler
}

func NewRootFlagSet() *RootFlagSet {
	return &RootFlagSet{
		fs:       pflag.NewFlagSet("root", pflag.ContinueOnError),
		commands: make(map[string]commandHandler),
	}
}

func (r *RootFlagSet) RegisterCommand(name string, h commandHandler) {
	r.commands[name] = h
}

func (r *RootFlagSet) Execute(version string) error {
	var showHelp, showVersion bool
	addDefaultFlags(r.fs, &showHelp, &showVersion)
	if err := r.fs.Parse(os.Args[1:]); err != nil {
		return err
	}

	if !checkVersion(showVersion, version) {
		return nil
	}

	remaining := r.fs.Args()

	if showHelp || len(remaining) < 1 {
		commonLogger.Println("Available commands:")
		// deterministic order
		keys := make([]string, 0, len(r.commands))
		for k := range r.commands {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			commonLogger.Println("  ", k)
		}
		return nil
	}

	cmdName := remaining[0]
	if h, ok := r.commands[cmdName]; ok {
		return h(remaining[1:])
	}
	return errors.New("unknown command: " + cmdName + ", see --help")
}

func checkVersion(showVersion bool, version string) bool {
	if showVersion {
		if version == "" {
			commonLogger.Println("version: unknown")
		} else {
			commonLogger.Println("version:", version)
		}
		return false
	}
	return true
}

func addDefaultFlags(fs *pflag.FlagSet, showHelp *bool, showVersion *bool) {
	fs.SetInterspersed(false)
	fs.BoolVarP(showHelp, "help", "h", false, "Show help")
	fs.BoolVarP(showVersion, "version", "v", false, "Show version")
}

type singleCommandHandler func(fs *pflag.FlagSet) error

type SingleFlagSet struct {
	command     singleCommandHandler
	fs          *pflag.FlagSet
	version     string
	showHelp    bool
	showVersion bool
}

func NewSingleFlagSet(command singleCommandHandler, version string) *SingleFlagSet {
	s := &SingleFlagSet{
		command: command,
		fs:      pflag.NewFlagSet(filepath.Base(os.Args[0]), pflag.ContinueOnError),
		version: version,
	}
	addDefaultFlags(s.fs, &s.showHelp, &s.showVersion)
	return s
}

func (s *SingleFlagSet) Fs() *pflag.FlagSet {
	return s.fs
}

func (s *SingleFlagSet) Execute() error {
	if err := s.fs.Parse(os.Args[1:]); err != nil {
		return err
	}
	if !checkVersion(s.showVersion, s.version) {
		return nil
	}
	if s.showHelp {
		s.fs.PrintDefaults()
		return nil
	}
	return s.command(s.fs)
}
