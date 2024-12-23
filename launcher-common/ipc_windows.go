package launcher_common

func ConfigAdminIpcPath() string {
	return `\\.\pipe\` + configAdminIpcName
}
