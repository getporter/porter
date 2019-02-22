package exec

func (m *Mixin) Uninstall(commandFile string) error {
	// re-use Install's logic
	return m.Install(commandFile)
}
