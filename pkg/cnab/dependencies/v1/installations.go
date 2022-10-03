package v1

import (
	"fmt"
)

// BuildPrerequisiteInstallationName generates the name of a prerequisite dependency installation.
func BuildPrerequisiteInstallationName(installation string, dependency string) string {
	return fmt.Sprintf("%s-%s", installation, dependency)
}
