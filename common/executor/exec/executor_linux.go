package exec

func visualAdminArgs() []arg {
	return []arg{newSafeArg("pkexec"), newSafeArg("--keep-cwd")}
}
