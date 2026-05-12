package custom

import commonExecutor "github.com/luskaner/ageLANServer/common/executor/exec"

type Exec struct {
	Executable string
}

func (exec Exec) execute(args []string, admin bool, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	options := commonExecutor.Options{File: exec.Executable, Args: args}
	if admin {
		options.AsAdmin = true
	}
	options.ShowWindow = true
	options.GUI = true
	optionsFn(options)
	result = options.Exec()
	return
}

func (exec Exec) Do(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	result = exec.execute(args, false, optionsFn)
	return
}

func (exec Exec) DoElevated(args []string, optionsFn func(options commonExecutor.Options)) (result *commonExecutor.Result) {
	result = exec.execute(args, true, optionsFn)
	return
}

func (exec Exec) String() string {
	return "Custom Path"
}
