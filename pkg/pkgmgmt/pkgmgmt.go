package pkgmgmt

import (
	"fmt"
	"os/exec"

	"get.porter.sh/porter/pkg/context"
)

// PackageManager handles searching, installing and communicating with packages.
type PackageManager interface {
	List() ([]string, error)
	GetPackageDir(name string) (string, error)
	GetMetadata(name string) (PackageMetadata, error)
	Install(InstallOptions) error
	Uninstall(UninstallOptions) error

	// Run a command against the installed package.
	Run(pkgContext *context.Context, name string, commandOpts CommandOptions) error
}

type PreRunHandler func(command string, cmd *exec.Cmd)

// CommandOptions is data necessary to execute a command against a package (mixin or plugin).
type CommandOptions struct {
	// Runtime specifies if the client or runtime executable should be targeted.
	Runtime bool

	// Command to pass to the package.
	Command string

	// Input to pipe to stdin.
	Input string

	// File argument to specify as --file.
	File string

	// PreRun allows the caller to tweak the command before it is executed.
	// This is only necessary if being called directly from a runner, if
	// using a PackageManager, this is set for you.
	PreRun PreRunHandler
}

// GetPackageListURL returns the URL for package listings of the provided type
func GetPackageListURL(pkgType string) string {
	return fmt.Sprintf("https://cdn.porter.sh/%ss/index.json", pkgType)
}
