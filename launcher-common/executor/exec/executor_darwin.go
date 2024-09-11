package exec

func terminalArgs() []string {
	return []string{"open", "-a", "Terminal", "--args"}
}

func adminArgs() []string {
	return []string{"sudo"}
}
