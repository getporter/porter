package exec

func (m *Mixin) Upgrade(commandFile string) error {
	// re-use Install's logic
	return m.Install(commandFile)
}
