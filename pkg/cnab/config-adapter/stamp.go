package configadapter

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
)

// Stamp contains Porter specific metadata about a bundle that we can place
// in the custom section of a bundle.json
type Stamp struct {
	// ManifestDigest takes into account all unique data that goes into a
	// porter build to help determine if the last build is stale.
	// * manifest
	// * mixins
	// * (TODO) files in current directory
	ManifestDigest string `json:"manifestDigest"`

	// Mixins used in the bundle.
	Mixins map[string]MixinRecord `json:"mixins"`

	// Manifest is the base64 encoded porter.yaml.
	EncodedManifest string `json:"manifest"`

	// Version and commit define the version of the Porter used when a bundle was built.
	Version string `json:"version"`
	Commit  string `json:"commit"`
}

// DecodeManifest base64 decodes the manifest stored in the stamp
func (s Stamp) DecodeManifest() ([]byte, error) {
	if s.EncodedManifest == "" {
		return nil, errors.New("no Porter manifest was embedded in the bundle")
	}

	resultB, err := base64.StdEncoding.DecodeString(s.EncodedManifest)
	if err != nil {
		return nil, fmt.Errorf("could not base64 decode the manifest in the stamp\n%s: %w", s.EncodedManifest, err)
	}

	return resultB, nil
}

func (s Stamp) WriteManifest(cxt *portercontext.Context, path string) error {
	manifestB, err := s.DecodeManifest()
	if err != nil {
		return err
	}

	err = cxt.FileSystem.WriteFile(path, manifestB, pkg.FileModeWritable)
	if err != nil {
		return fmt.Errorf("could not save decoded manifest to %s: %w", path, err)
	}

	return nil
}

// MixinRecord contains information about a mixin used in a bundle
// For now it is a placeholder for data that we would like to include in the future.
type MixinRecord struct {
	// Name of the mixin used in the bundle. This is used for sorting only, and
	// should not be written to the Porter's stamp in bundle.json because we are
	// storing these mixin records in a map, keyed by the mixin name.
	Name string `json:"-"`

	// Version of the mixin used in the bundle.
	Version string `json:"version"`
}

type MixinRecords []MixinRecord

func (m MixinRecords) Len() int {
	return len(m)
}

func (m MixinRecords) Less(i, j int) bool {
	return m[i].Name < m[j].Name
}

func (m MixinRecords) Swap(i, j int) {
	tmp := m[i]
	m[i] = m[j]
	m[j] = tmp
}

func (c *ManifestConverter) GenerateStamp(ctx context.Context) (Stamp, error) {
	log := tracing.LoggerFromContext(ctx)

	stamp := Stamp{}

	// Remember the original porter.yaml, base64 encoded to avoid canonical json shenanigans
	rawManifest, err := manifest.ReadManifestData(c.config.Context, c.Manifest.ManifestPath)
	if err != nil {
		return Stamp{}, err
	}
	stamp.EncodedManifest = base64.StdEncoding.EncodeToString(rawManifest)

	stamp.Mixins = make(map[string]MixinRecord, len(c.Manifest.Mixins))
	usedMixins := c.getUsedMixinRecords()
	for _, record := range usedMixins {
		stamp.Mixins[record.Name] = record
	}

	digest, err := c.DigestManifest()
	if err != nil {
		// The digest is only used to decide if we need to rebuild, it is not an error condition to not
		// have a digest.
		log.Error(fmt.Errorf("WARNING: Could not digest the porter manifest file: %w", err))
		stamp.ManifestDigest = "unknown"
	} else {
		stamp.ManifestDigest = digest
	}

	stamp.Version = pkg.Version
	stamp.Commit = pkg.Commit

	return stamp, nil
}

func (c *ManifestConverter) DigestManifest() (string, error) {
	if exists, _ := c.config.FileSystem.Exists(c.Manifest.ManifestPath); !exists {
		return "", fmt.Errorf("the specified porter configuration file %s does not exist", c.Manifest.ManifestPath)
	}

	data, err := c.config.FileSystem.ReadFile(c.Manifest.ManifestPath)
	if err != nil {
		return "", fmt.Errorf("could not read manifest at %q: %w", c.Manifest.ManifestPath, err)
	}

	v := pkg.Version
	data = append(data, v...)

	usedMixins := c.getUsedMixinRecords()
	sort.Sort(usedMixins) // Ensure that this is sorted so the digest is consistent
	for _, mixinRecord := range usedMixins {
		data = append(append(data, mixinRecord.Name...), mixinRecord.Version...)
	}

	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:]), nil
}

func LoadStamp(bun cnab.ExtendedBundle) (Stamp, error) {
	// TODO(carolynvs): can we simplify some of this by using the extended bundle?
	data, ok := bun.Custom[config.CustomPorterKey]
	if !ok {
		return Stamp{}, fmt.Errorf("porter stamp (custom.%s) was not present on the bundle", config.CustomPorterKey)
	}

	dataB, err := json.Marshal(data)
	if err != nil {
		return Stamp{}, fmt.Errorf("could not marshal the porter stamp %q: %w", string(dataB), err)
	}

	stamp := Stamp{}
	err = json.Unmarshal(dataB, &stamp)
	if err != nil {
		return Stamp{}, fmt.Errorf("could not unmarshal the porter stamp %q: %w", string(dataB), err)
	}

	return stamp, nil
}

// getUsedMixinRecords compare the mixins defined in the manifest and the ones installed and then retrieve the mixin's version info
func (c *ManifestConverter) getUsedMixinRecords() MixinRecords {
	usedMixins := make(MixinRecords, 0)

	for _, usedMixin := range c.Manifest.Mixins {
		for _, installedMixin := range c.InstalledMixins {
			if usedMixin.Name == installedMixin.Name {
				usedMixins = append(usedMixins, MixinRecord{
					Name:    installedMixin.Name,
					Version: installedMixin.GetVersionInfo().Version,
				})
			}
		}
	}

	sort.Sort(usedMixins)
	return usedMixins
}
