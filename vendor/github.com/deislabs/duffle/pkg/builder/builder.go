package builder

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"

	"github.com/deislabs/duffle/pkg/bundle"
	"github.com/deislabs/duffle/pkg/duffle/manifest"
	"github.com/deislabs/duffle/pkg/imagebuilder"
)

// Builder defines how to interact with a bundle builder
type Builder struct {
	ID      string
	LogsDir string
	// If this is true, versions will contain build metadata
	// Example:
	//   0.1.2+2c3c59e8a5adad62d2245cbb7b2a8685b1a9a717
	VersionWithBuildMetadata bool
	ImageBuilders            []imagebuilder.ImageBuilder
}

// New returns a new Builder
func New() *Builder {
	return &Builder{
		ID: getulid(),
	}
}

// Logs returns the path to the build logs.
//
// Set after Up is called (otherwise "").
func (b *Builder) Logs(appName string) string {
	return filepath.Join(b.LogsDir, appName, b.ID)
}

// Context contains information about the application
type Context struct {
	Manifest *manifest.Manifest
	AppDir   string
}

// AppContext contains state information carried across various duffle stage boundaries
type AppContext struct {
	Bldr *Builder
	Ctx  *Context
	Log  io.WriteCloser
	ID   string
}

// PrepareBuild prepares a build
func (b *Builder) PrepareBuild(bldr *Builder, mfst *manifest.Manifest, appDir string, imageBuilders []imagebuilder.ImageBuilder) (*AppContext, *bundle.Bundle, error) {
	b.ImageBuilders = imageBuilders

	ctx := &Context{
		AppDir:   appDir,
		Manifest: mfst,
	}

	bf := &bundle.Bundle{
		Name:        ctx.Manifest.Name,
		Description: ctx.Manifest.Description,
		Images:      ctx.Manifest.Images,
		Keywords:    ctx.Manifest.Keywords,
		Maintainers: ctx.Manifest.Maintainers,
		Actions:     ctx.Manifest.Actions,
		Parameters:  ctx.Manifest.Parameters,
		Credentials: ctx.Manifest.Credentials,
	}

	for _, imb := range imageBuilders {
		invImage := ctx.Manifest.InvocationImages[imb.Name()]
		if invImage == nil {
			return nil, nil, errors.New(fmt.Sprintf("could not find an invocation image for %s", imb.Name()))
		}
		registry := invImage.Configuration["registry"]
		if err := imb.PrepareBuild(ctx.AppDir, registry, ctx.Manifest.Name); err != nil {
			return nil, nil, err
		}

		ii := bundle.InvocationImage{}
		ii.Image = imb.URI()
		ii.ImageType = imb.Type()
		bf.InvocationImages = append(bf.InvocationImages, ii)

		baseVersion := mfst.Version
		if baseVersion == "" {
			baseVersion = "0.1.0"
		}
		newver, err := b.version(baseVersion, strings.Split(imb.URI(), ":")[1])
		if err != nil {
			return nil, nil, err
		}
		bf.Version = newver
	}

	app := &AppContext{
		ID:   bldr.ID,
		Bldr: bldr,
		Ctx:  ctx,
		Log:  os.Stdout,
	}

	return app, bf, nil
}

func (b *Builder) version(baseVersion, sha string) (string, error) {
	sv, err := semver.NewVersion(baseVersion)
	if err != nil {
		return baseVersion, err
	}

	if b.VersionWithBuildMetadata {
		newsv, err := sv.SetMetadata(sha)
		if err != nil {
			return baseVersion, err
		}
		return newsv.String(), nil
	}

	return sv.String(), nil
}

// Build passes the context of each component to its respective builder
func (b *Builder) Build(ctx context.Context, app *AppContext) error {
	if err := buildInvocationImages(ctx, b.ImageBuilders, app); err != nil {
		return fmt.Errorf("error building image: %v", err)
	}
	return nil
}

func buildInvocationImages(ctx context.Context, imageBuilders []imagebuilder.ImageBuilder, app *AppContext) (err error) {
	errc := make(chan error)

	go func() {
		defer close(errc)
		var wg sync.WaitGroup
		wg.Add(len(imageBuilders))

		for _, c := range imageBuilders {
			go func(c imagebuilder.ImageBuilder) {
				defer wg.Done()
				err = c.Build(ctx, app.Log)
				if err != nil {
					errc <- fmt.Errorf("error building image %v: %v", c.Name(), err)
				}
			}(c)
		}

		wg.Wait()
	}()

	for errc != nil {
		select {
		case err, ok := <-errc:
			if !ok {
				errc = nil
				continue
			}
			return err
		default:
			time.Sleep(time.Second)
		}
	}
	return nil
}
