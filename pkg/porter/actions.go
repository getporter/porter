package porter

type Action string

const (
	ActionInstall   Action = "install"
	ActionUpgrade   Action = "upgrade"
	ActionUninstall Action = "uninstall"
)
