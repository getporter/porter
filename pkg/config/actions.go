package config

import "fmt"

type Action string

const (
	ActionInstall   Action = "install"
	ActionUpgrade   Action = "upgrade"
	ActionUninstall Action = "uninstall"
)

func ParseAction(value string) (Action, error) {
	action := Action(value)
	switch action {
	case ActionInstall, ActionUpgrade, ActionUninstall:
		return action, nil
	default:
		return "", fmt.Errorf("invalid action %q", value)
	}
}
