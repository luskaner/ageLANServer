//go:build !windows

package wine

import (
	commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"
	"github.com/luskaner/ageLANServer/common/game/wine"
)

type CustomExec interface {
	DoCustom(args []string, optionsFn func(options *commonExecutor.Options)) (result *commonExecutor.Result)
}

type Exec struct{}

func NewExec() (exec *Exec, ok bool) {
	if wine.Prefix() != "" {
		exec = &Exec{}
		ok = true
	}
	return
}

func (exec Exec) DoCustom(args []string, optionsFn func(options *commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{
		File:        "wine",
		SpecialFile: true,
		Args:        args,
		Pid:         true,
	}
	options.Args = append(options.Args, args...)
	if optionsFn != nil {
		optionsFn(&options)
	}
	result = options.Exec()
	return
}
