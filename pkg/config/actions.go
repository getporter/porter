package config

import (
	"fmt"
	"strings"
)

type Action string

const (
	ActionInstall    Action = "install"
	ActionUpgrade    Action = "upgrade"
	ActionUninstall  Action = "uninstall"
	ErrInvalidAction string = "invalid action"
)

// IsSupportedAction determines if the value is an action supported by Porter.
func IsSupportedAction(value string) bool {
	_, err := ParseAction(value)
	return err == nil
}

// ParseAction converts a string into an Action, or returns an error message.
func ParseAction(value string) (Action, error) {
	action := Action(value)
	switch action {
	case ActionInstall, ActionUpgrade, ActionUninstall:
		return action, nil
	default:
		return "", fmt.Errorf("%s %q", ErrInvalidAction, value)
	}
}

func GetSupportActions() []Action {
	return []Action{ActionInstall, ActionUpgrade, ActionUninstall}
}

// IsInvalidActionError determines if an error is the error returned by ParseAction when
// a value isn't a valid action.
func IsInvalidActionError(err error) bool {
	return strings.HasPrefix(err.Error(), ErrInvalidAction)
}
