package porter

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"get.porter.sh/porter/pkg/mixin"
	"get.porter.sh/porter/pkg/pkgmgmt"
	"get.porter.sh/porter/pkg/pkgmgmt/feed"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/printer"
)

const (
	SkeletorRepo = "https://github.com/getporter/skeletor"
)

// PrintMixinsOptions represent options for the PrintMixins function
type PrintMixinsOptions struct {
	printer.PrintOptions
}

func (p *Porter) PrintMixins(ctx context.Context, opts PrintMixinsOptions) error {
	mixins, err := p.ListMixins(ctx)
	if err != nil {
		return err
	}

	switch opts.Format {
	case printer.FormatPlaintext:
		printMixinRow :=
			func(v interface{}) []string {
				m, ok := v.(mixin.Metadata)
				if !ok {
					return nil
				}
				return []string{m.Name, m.VersionInfo.Version, m.VersionInfo.Author}
			}
		return printer.PrintTable(p.Out, mixins, printMixinRow, "Name", "Version", "Author")
	case printer.FormatJson:
		return printer.PrintJson(p.Out, mixins)
	case printer.FormatYaml:
		return printer.PrintYaml(p.Out, mixins)
	default:
		return fmt.Errorf("invalid format: %s", opts.Format)
	}
}

func (p *Porter) ListMixins(ctx context.Context) ([]mixin.Metadata, error) {
	// List out what is installed on the file system
	names, err := p.Mixins.List()
	if err != nil {
		return nil, err
	}

	// Query each mixin and fill out their metadata
	mixins := make([]mixin.Metadata, len(names))
	for i, name := range names {
		m, err := p.Mixins.GetMetadata(ctx, name)
		if err != nil {
			fmt.Fprintf(p.Err, "could not get version from mixin %s: %s\n ", name, err.Error())
			continue
		}

		meta, _ := m.(*mixin.Metadata)
		mixins[i] = *meta
	}

	return mixins, nil
}

func (p *Porter) InstallMixin(ctx context.Context, opts mixin.InstallOptions) error {
	err := p.Mixins.Install(ctx, opts.InstallOptions)
	if err != nil {
		return err
	}

	mixin, err := p.Mixins.GetMetadata(ctx, opts.Name)
	if err != nil {
		return err
	}

	v := mixin.GetVersionInfo()
	fmt.Fprintf(p.Out, "installed %s mixin %s (%s)\n", opts.Name, v.Version, v.Commit)

	return nil
}

func (p *Porter) UninstallMixin(ctx context.Context, opts pkgmgmt.UninstallOptions) error {
	err := p.Mixins.Uninstall(ctx, opts)
	if err != nil {
		return err
	}

	fmt.Fprintf(p.Out, "Uninstalled %s mixin", opts.Name)

	return nil
}

func (p *Porter) GenerateMixinFeed(ctx context.Context, opts feed.GenerateOptions) error {
	f := feed.NewMixinFeed(p.Context)

	err := f.Generate(ctx, opts)
	if err != nil {
		return err
	}

	return f.Save(opts)
}

func (p *Porter) CreateMixinFeedTemplate() error {
	return feed.CreateTemplate(p.Context)
}

// MixinsCreateOptions represent options for Porter's mixin create command
type MixinsCreateOptions struct {
	MixinName      string
	AuthorName     string
	AuthorUsername string
	DirPath        string
}

func (o *MixinsCreateOptions) Validate(args []string, cxt *portercontext.Context) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("mixin name is required")
	}

	if len(args) > 1 {
		return fmt.Errorf("only one positional argument may be specified, the mixin name, but multiple were received: %s", args)
	}

	o.MixinName = args[0]

	if o.AuthorName == "" {
		return errors.New("must provide a value for flag --author")
	}

	if o.AuthorUsername == "" {
		return errors.New("must provide a value for flag --username")
	}

	if o.DirPath == "" {
		o.DirPath = cxt.Getwd()
	}

	if _, err := cxt.FileSystem.Stat(o.DirPath); err != nil {
		return fmt.Errorf("invalid --dir: %s: %w", o.DirPath, err)
	}

	return nil
}

func (p *Porter) CreateMixin(opts MixinsCreateOptions) error {
	skeletorDestPath := opts.DirPath + "/" + opts.MixinName

	if err := exec.Command("git", "clone", SkeletorRepo, skeletorDestPath).Run(); err != nil {
		return fmt.Errorf("failed cloning skeletor repo: %w", err)
	}

	err := os.Rename(skeletorDestPath+"/cmd/skeletor", skeletorDestPath+"/cmd/"+opts.MixinName)
	if err != nil {
		return err
	}

	err = os.Rename(skeletorDestPath+"/pkg/skeletor", skeletorDestPath+"/pkg/"+opts.MixinName)
	if err != nil {
		return err
	}

	replacementList := map[string]string{
		"get.porter.sh/mixin/skeletor":       fmt.Sprintf("github.com/%s/%s", opts.AuthorUsername, opts.MixinName),
		"PKG = get.porter.sh/mixin/$(MIXIN)": fmt.Sprintf("PKG = github.com/%s/%s", opts.AuthorUsername, opts.MixinName),
		"skeletor":                           opts.MixinName,
		"YOURNAME":                           opts.AuthorName,
	}

	for replaced, replacement := range replacementList {
		err := replaceStringInDir(skeletorDestPath, replaced, replacement)
		if err != nil {
			return err
		}
	}

	fmt.Fprintf(p.Out, "Created %s mixin\n", opts.MixinName)

	return nil
}

// replaceStringInDir walks through all the file in a designated directory and replace any occurence of a string with a particular replacement
// while skipping specifically directory .git and file README.md
func replaceStringInDir(dir, replaced, replacement string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}
		if !info.IsDir() && info.Name() != "README.md" {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			err = ioutil.WriteFile(path, bytes.Replace(content, []byte(replaced), []byte(replacement), -1), info.Mode().Perm())
			if err != nil {
				return err
			}
		}

		return nil
	})
}
