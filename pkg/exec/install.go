package exec

// InstallOptions represent the options for an exec mixin install action
type InstallOptions struct {
	File string
}

func (m *Mixin) Install(commandFile string) error {
	err := m.loadAction(commandFile)
	if err != nil {
		return err
	}
	return m.Execute()
}
