package pkgmgmt

import (
	"context"
	"net/url"
	"os/exec"
	"path"

	"get.porter.sh/porter/pkg/portercontext"
)

// PackageManager handles searching, installing and communicating with packages.
type PackageManager interface {
	List() ([]string, error)
	GetPackageDir(name string) (string, error)
	GetMetadata(ctx context.Context, name string) (PackageMetadata, error)
	Install(ctx context.Context, opts InstallOptions) error
	Uninstall(ctx context.Context, opts UninstallOptions) error

	// Run a command against the installed package.
	Run(ctx context.Context, pkgContext *portercontext.Context, name string, commandOpts CommandOptions) error
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

// GetPackageListURL returns the URL for package listings of the provided type.
func GetPackageListURL(mirror url.URL, pkgType string) string {
	mirror.Path = path.Join(mirror.Path, pkgType+"s", "index.json")
	return mirror.String()
}
