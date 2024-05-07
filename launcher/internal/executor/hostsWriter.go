package executor

import (
	"shared/executor"
)

const processName string = "launcher.hostsEditor.exe"
const ipFlag string = "-ip="
const addFlag string = "-add="

func run(admin bool, ip string, add bool) bool {
	var boolStr string
	if add {
		boolStr = "true"
	} else {
		boolStr = "false"
	}
	args := []string{ipFlag + ip, addFlag + boolStr}
	if admin {
		return executor.RunCustomExecutable("./"+processName, args...)
	}
	argsStr := ""
	for _, arg := range args {
		argsStr += ", '" + arg + "'"
	}
	return executor.RunCustomExecutable("powershell", "Start-Process", "admin.exe", "-WindowStyle Hidden", "-Wait", "-Verb", "runAs", "-ArgumentList", argsStr[2:])
}

func AddHost(admin bool, ip string) bool {
	return run(admin, ip, true)
}

func RemoveHost(admin bool) bool {
	return run(admin, "", false)
}
