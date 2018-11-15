package exec

func (m *Mixin) Install(commandFile string) error {
	err := m.LoadInstruction(commandFile)
	if err != nil {
		return err
	}
	return m.Execute()
}
