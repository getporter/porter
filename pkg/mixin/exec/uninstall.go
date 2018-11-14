package exec

func (e *Exec) Uninstall(commandFile string) error {
	err := e.LoadInstruction(commandFile)
	if err != nil {
		return err
	}
	return e.Execute()
}
