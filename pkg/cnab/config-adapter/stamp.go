package configadapter

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"get.porter.sh/porter/pkg"
	"get.porter.sh/porter/pkg/cnab"
	"get.porter.sh/porter/pkg/config"
	"get.porter.sh/porter/pkg/manifest"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/tracing"
	"github.com/Masterminds/semver/v3"
)

// Stamp contains Porter specific metadata about a bundle that we can place
// in the custom section of a bundle.json
type Stamp struct {
	// ManifestDigest takes into account all unique data that goes into a
	// porter build to help determine if the last build is stale.
	// * manifest
	// * mixins
	// * files in current directory (content and executable permissions)
	ManifestDigest string `json:"manifestDigest"`

	// Mixins used in the bundle.
	Mixins map[string]MixinRecord `json:"mixins"`

	// Manifest is the base64 encoded porter.yaml.
	EncodedManifest string `json:"manifest"`

	// Version and commit define the version of the Porter used when a bundle was built.
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	PreserveTags bool   `json:"preserveTags"`
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
	// Currently there can only be a single version of a mixin used in a bundle
	// I'm considering version as well for sorting in case that changes in the future once mixins are bundles
	// referenced by a bundle, and not embedded binaries
	iRecord := m[i]
	jRecord := m[j]
	if iRecord.Name == jRecord.Name {
		// Try to sort by the mixin's semantic version
		// If it doesn't parse, just fall through and sort as a string instead
		iVersion, iErr := semver.NewVersion(iRecord.Version)
		jVersion, jErr := semver.NewVersion(jRecord.Version)
		if iErr == nil && jErr == nil {
			return iVersion.LessThan(jVersion)
		} else {
			return iRecord.Version < jRecord.Version
		}
	}

	return iRecord.Name < jRecord.Name
}

func (m MixinRecords) Swap(i, j int) {
	tmp := m[i]
	m[i] = m[j]
	m[j] = tmp
}

func (c *ManifestConverter) GenerateStamp(ctx context.Context, preserveTags bool) (Stamp, error) {
	log := tracing.LoggerFromContext(ctx)

	stamp := Stamp{}

	// Remember the original porter.yaml, base64 encoded to avoid canonical json shenanigans
	rawManifest, err := manifest.ReadManifestData(c.config.Context, c.Manifest.ManifestPath)
	if err != nil {
		return Stamp{}, err
	}
	stamp.EncodedManifest = base64.StdEncoding.EncodeToString(rawManifest)
	stamp.PreserveTags = preserveTags

	stamp.Mixins = make(map[string]MixinRecord, len(c.Manifest.Mixins))
	usedMixins := c.getUsedMixinRecords()
	for _, record := range usedMixins {
		stamp.Mixins[record.Name] = record
	}

	digest, err := c.DigestManifest()
	if err != nil {
		// The digest is only used to decide if we need to rebuild, it is not an error condition to not
		// have a digest.
		log.Warn(fmt.Sprint("WARNING: Could not digest the porter manifest file: %w", err))
		stamp.ManifestDigest = "unknown"
	} else {
		stamp.ManifestDigest = digest
	}

	stamp.Version = pkg.Version
	stamp.Commit = pkg.Commit

	return stamp, nil
}

// hashBundleFiles walks the bundle directory and hashes all relevant files,
// including their content and executable permissions (cross-platform compatible).
// Returns a sorted map of relative file paths to their content hashes.
func (c *ManifestConverter) hashBundleFiles(bundleDir string) (map[string]string, error) {
	fileHashes := make(map[string]string)

	// Directories and files to skip
	skipDirs := map[string]bool{
		".cnab":        true,
		".git":         true,
		"node_modules": true,
		".porter":      true,
		"vendor":       true,
	}

	// Use afero Walk to work with virtual filesystem in tests
	err := c.config.FileSystem.Walk(bundleDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			// Skip files that don't exist or can't be accessed
			return nil
		}

		// Get relative path from bundle directory
		relPath, err := filepath.Rel(bundleDir, path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		// Skip hidden files and directories (cross-platform)
		if strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip specific directories
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}

		// Only process regular files
		if !info.Mode().IsRegular() {
			return nil
		}

		// Read file content
		content, err := c.config.FileSystem.ReadFile(path)
		if err != nil {
			// Skip files that can't be read
			return nil
		}

		// Hash the file content
		contentHash := sha256.Sum256(content)
		hashStr := hex.EncodeToString(contentHash[:])

		// Check if file is executable
		if isExecutable(info) {
			// Append executable marker to hash to ensure permission changes trigger rebuild
			hashStr += ":x"
		}

		// Use forward slashes for consistency across platforms
		normalizedPath := filepath.ToSlash(relPath)
		fileHashes[normalizedPath] = hashStr

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk bundle directory: %w", err)
	}

	return fileHashes, nil
}

// isExecutable checks if a file has the executable permission bit set.
// Bundles always run in a Linux distribution, so we only check Unix permissions on Linux.
// On other platforms, we skip the check to avoid false positives from permission emulation.
func isExecutable(info fs.FileInfo) bool {
	// Only check executable bit on Linux where bundles actually run
	if runtime.GOOS != "linux" {
		return false
	}
	mode := info.Mode()
	return mode&0111 != 0
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

	// Hash all files in the bundle directory to detect content and permission changes
	bundleDir := filepath.Dir(c.Manifest.ManifestPath)
	fileHashes, err := c.hashBundleFiles(bundleDir)
	if err != nil {
		// Log warning but continue - file hashing is an enhancement, not critical
		// This maintains backward compatibility if there are issues accessing files
		fmt.Fprintf(io.Discard, "WARNING: Could not hash bundle files: %v\n", err)
	} else {
		// Sort file paths for deterministic digest
		sortedPaths := make([]string, 0, len(fileHashes))
		for path := range fileHashes {
			sortedPaths = append(sortedPaths, path)
		}
		sort.Strings(sortedPaths)

		// Append file hashes to digest data
		for _, path := range sortedPaths {
			hash := fileHashes[path]
			data = append(data, []byte(path)...)
			data = append(data, []byte(hash)...)
		}
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

// getUsedMixinRecords returns a list of the mixins used by the bundle, including
// information about the installed mixin, such as its version.
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
