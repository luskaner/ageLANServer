package base

import commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"

func StartUri(uri string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: "open", Args: []string{uri}, SpecialFile: true, ExitCode: true, Wait: true}
	optionsFn(options)
	result = options.Exec()
	return
}
