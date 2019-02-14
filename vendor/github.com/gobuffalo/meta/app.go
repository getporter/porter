package meta

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/flect/name"
	"github.com/pkg/errors"
	"github.com/rogpeppe/go-internal/modfile"
)

// PackageJSON stores package.json meta data used by Buffalo.
type PackageJSON struct {
	Scripts map[string]string `json:"scripts"`
}

// App represents meta data for a Buffalo application on disk.
type App struct {
	Pwd         string      `json:"pwd" toml:"-"`
	Root        string      `json:"root" toml:"-"`
	GoPath      string      `json:"go_path" toml:"-"`
	PackagePkg  string      `json:"package_path" toml:"-"`
	ActionsPkg  string      `json:"actions_path" toml:"-"`
	ModelsPkg   string      `json:"models_path" toml:"-"`
	GriftsPkg   string      `json:"grifts_path" toml:"-"`
	WithModules bool        `json:"with_modules" toml:"-"`
	Name        name.Ident  `json:"name" toml:"name"`
	Bin         string      `json:"bin" toml:"bin"`
	VCS         string      `json:"vcs" toml:"vcs"`
	WithPop     bool        `json:"with_pop" toml:"with_pop"`
	WithSQLite  bool        `json:"with_sqlite" toml:"with_sqlite"`
	WithDep     bool        `json:"with_dep" toml:"with_dep"`
	WithWebpack bool        `json:"with_webpack" toml:"with_webpack"`
	WithNodeJs  bool        `json:"with_nodejs" toml:"with_nodejs"`
	WithYarn    bool        `json:"with_yarn" toml:"with_yarn"`
	WithDocker  bool        `json:"with_docker" toml:"with_docker"`
	WithGrifts  bool        `json:"with_grifts" toml:"with_grifts"`
	AsWeb       bool        `json:"as_web" toml:"as_web"`
	AsAPI       bool        `json:"as_api" toml:"as_api"`
	PackageJSON PackageJSON `json:"-" toml:"-"`
}

// IsZero checks if the App struct has no set field.
func (a App) IsZero() bool {
	return a.String() == App{}.String()
}

func resolvePackageName(name string, pwd string) string {
	result, _ := envy.CurrentModule()

	if filepath.Base(result) != name {
		result = path.Join(result, name)
	}
	if envy.Mods() {
		modp := filepath.Join(pwd, name, "go.mod")
		if strings.HasSuffix(pwd, name) {
			modp = filepath.Join(pwd, "go.mod")
		}
		moddata, err := ioutil.ReadFile(modp)
		if err != nil {
			if envy.InGoPath() {
				p := envy.CurrentPackage()
				if !strings.HasSuffix(p, name) {
					return path.Join(p, name)
				}
				return p
			}
			return name
		}
		packagePath := modfile.ModulePath(moddata)
		if packagePath == "" {
			return name
		}
		return packagePath
	}

	return result
}

// ResolveSymlinks takes a path and gets the pointed path
// if the original one is a symlink.
func ResolveSymlinks(p string) string {
	cd, err := os.Lstat(p)
	if err != nil {
		return p
	}
	if cd.Mode()&os.ModeSymlink != 0 {
		// This is a symlink
		r, err := filepath.EvalSymlinks(p)
		if err != nil {
			return p
		}
		return r
	}
	return p
}

func (a App) String() string {
	b, _ := json.Marshal(a)
	return string(b)
}

// Encode the list of plugins, in TOML format, to the reader
func (a App) Encode(w io.Writer) error {
	if err := toml.NewEncoder(w).Encode(a); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// Decode the list of plugins, in TOML format, from the reader
func (a *App) Decode(r io.Reader) error {
	xa := New(".")
	if _, err := toml.DecodeReader(r, &xa); err != nil {
		return errors.WithStack(err)
	}
	(*a) = xa
	return nil
}

// PackageRoot sets the root package of the application and
// recalculates package related values
func (a *App) PackageRoot(pp string) {
	a.PackagePkg = pp
	a.ActionsPkg = pp + "/actions"
	a.ModelsPkg = pp + "/models"
	a.GriftsPkg = pp + "/grifts"
}

// NodeScript gets the "scripts" section from package.json and
// returns the matching script if it exists.
func (a App) NodeScript(name string) (string, error) {
	if !a.WithNodeJs {
		return "", errors.New("package.json not found")
	}

	s, ok := a.PackageJSON.Scripts[name]
	if ok {
		return s, nil
	}
	return "", fmt.Errorf("node script %s not found", name)
}
