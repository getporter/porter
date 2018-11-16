package exec

func (m *Mixin) Uninstall(commandFile string) error {
	err := m.LoadInstruction(commandFile)
	if err != nil {
		return err
	}
	return m.Execute()
}
