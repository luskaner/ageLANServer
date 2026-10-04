// Package cmd is the console launcher's front end.
//
// The session itself is not here. It is in launcher-common/launcher/session, so
// that anything that keeps running, such as a window, can perform one. What is
// left is the part only a console has: the flag set it parses, where the
// narration goes, and which backend answers the questions when the run asks.
package cmd

import (
	"context"

	"github.com/luskaner/ageLANServer/common"
	commonCmd "github.com/luskaner/ageLANServer/common/cmd"
	"github.com/luskaner/ageLANServer/launcher-common/launcher"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/ops/logger"
	"github.com/luskaner/ageLANServer/launcher-common/launcher/session"
	"github.com/luskaner/ageLANServer/launcher-common/ui"
	"github.com/luskaner/ageLANServer/launcher/internal/dialog"
	"github.com/spf13/pflag"
	"os"
)

// Version is set from main and shown in the session summary.
var Version string

// Execute parses the command line and performs the session.
func Execute() (err error, exitCode int) {
	singleFs := commonCmd.NewSingleFlagSet(runRoot, Version)
	// Registered by the shared launcher, not here, so that a frontend which
	// registers the same options on a flag set of its own gets the same names,
	// the same defaults and the same help out of the same files.
	if err := session.BindFlags(singleFs.Fs()); err != nil {
		return err, common.ErrSyntax
	}
	return singleFs.Execute()
}

// runRoot exists to be the function a flag set can call, so it takes no context
// and gets one.
func runRoot(fs *pflag.FlagSet) (err error, exitCode int) {
	return runSession(context.Background(), fs)
}

// runSession is what a console run performs. A frontend that keeps running
// calls session.Run with a context it can cancel instead.
func runSession(ctx context.Context, fs *pflag.FlagSet) (err error, exitCode int) {
	// Installed per run rather than at init: the reporter writes to a log file
	// that a test replaces, and one captured at init would keep writing to the
	// log of whichever run happened to be first.
	configure()
	return session.Run(ctx, fs)
}

// configure hands the session everything that is specific to a console.
//
// It is a function rather than an argument because the answer to "which backend
// asks the questions" is not known until a run has read its configuration, and
// it is the same answer for every run of this program.
func configure() {
	session.Configure(session.Setup{
		Version: Version,
		Report:  loggerReporter{},
		NewDialog: func(mode string) launcher.Resolution {
			resolution := dialog.New(mode)
			// The console backend prints its questions through the same sinks as
			// everything else, so they reach the log file too.
			dialog.SetOutput(dialog.Output{Println: loggerReporter{}.Println, Printf: loggerReporter{}.Printf})
			return resolution
		},
		Stdin:     os.Stdin,
		Presenter: ui.Presenter{},
	})
}

// loggerReporter is the console's Reporter: the same lines, the same decoration,
// in the same places as before the session moved out of this module.
type loggerReporter struct{}

func (loggerReporter) Ok(format string, a ...any)     { logger.Ok(format, a...) }
func (loggerReporter) Fail(format string, a ...any)   { logger.Fail(format, a...) }
func (loggerReporter) Warn(format string, a ...any)   { logger.Warn(format, a...) }
func (loggerReporter) Info(format string, a ...any)   { logger.Info(format, a...) }
func (loggerReporter) Step(format string, a ...any)   { logger.Step(format, a...) }
func (loggerReporter) Detail(format string, a ...any) { logger.Detail(format, a...) }
func (loggerReporter) Fault(format string, a ...any)  { logger.Fault(format, a...) }
func (loggerReporter) Println(a ...any)               { logger.Println(a...) }
func (loggerReporter) Printf(format string, a ...any) { logger.Printf(format, a...) }
