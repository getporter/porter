package mixin

const (
	OutputsDir = "/cnab/app/porter/outputs"
)

func IsCoreMixinCommand(value string) bool {
	switch value {
	case "install", "upgrade", "uninstall", "build", "schema", "version":
		return true
	default:
		return false
	}
}
