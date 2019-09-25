package config

type Action string

const (
	ActionInstall   Action = "install"
	ActionUpgrade   Action = "upgrade"
	ActionUninstall Action = "uninstall"
	ActionCustom    Action = "custom"
)

// IsCoreAction determines if the value is a core action from the CNAB spec.
func IsCoreAction(value Action) bool {
	for _, a := range GetCoreActions() {
		if value == a {
			return true
		}
	}
	return false
}

func GetCoreActions() []Action {
	return []Action{ActionInstall, ActionUpgrade, ActionUninstall}
}
