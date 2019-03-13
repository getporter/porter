package exec

func (m *Mixin) Install(commandFile string) error {
	err := m.loadAction(commandFile)
	if err != nil {
		return err
	}
	return m.Execute()
}
