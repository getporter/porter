package exec

// UninstallOptions represent the options for an exec mixin uninstall action
type UninstallOptions struct {
	File string
}

func (m *Mixin) Uninstall(commandFile string) error {
	// re-use Install's logic
	return m.Install(commandFile)
}
