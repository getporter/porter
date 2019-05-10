package exec

// UpgradeOptions represent the options for an exec mixin upgrade action
type UpgradeOptions struct {
	File string
}

func (m *Mixin) Upgrade(commandFile string) error {
	// re-use Install's logic
	return m.Install(commandFile)
}
