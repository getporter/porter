package mixin

func IsCoreMixinCommand(value string) bool {
	switch value {
	case "install", "upgrade", "uninstall", "build", "schema", "version":
		return true
	default:
		return false
	}
}
