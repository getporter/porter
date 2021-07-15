package manifest

import (
	"get.porter.sh/porter/pkg/cnab"
)

// IsCoreAction determines if the value is a core action from the CNAB spec.
func IsCoreAction(value string) bool {
	for _, a := range GetCoreActions() {
		if value == a {
			return true
		}
	}
	return false
}

func GetCoreActions() []string {
	return []string{cnab.ActionInstall, cnab.ActionUpgrade, cnab.ActionUninstall}
}
