package manifest

import (
	"github.com/cnabio/cnab-go/claim"
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
	return []string{claim.ActionInstall, claim.ActionUpgrade, claim.ActionUninstall}
}
