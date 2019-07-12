package exec

// ExecOptions represent the options for any exec command
type ExecuteCommandOptions struct {
	File string
}

func (m *Mixin) ExecuteCommand(opts ExecuteCommandOptions) error {
	err := m.loadAction(opts.File)
	if err != nil {
		return err
	}
	return m.Execute()
}
