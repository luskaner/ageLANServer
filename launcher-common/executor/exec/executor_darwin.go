package exec

func terminalArgs() []string {
	return []string{"open", "-a", "Terminal", "--args"}
}

func adminArgs(_ bool) []string {
	return []string{"sudo"}
}
