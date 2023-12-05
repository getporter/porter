package cnab

import (
	"encoding/json"
	"fmt"
	"sort"

	depsv1ext "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v1"
	v2 "get.porter.sh/porter/pkg/cnab/extensions/dependencies/v2"
	"get.porter.sh/porter/pkg/portercontext"
	"get.porter.sh/porter/pkg/schema"
	"github.com/Masterminds/semver/v3"
	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/bundle/definition"
	"github.com/cnabio/cnab-go/claim"
	"github.com/google/go-containerregistry/pkg/crane"
)

const SupportedVersion = "1.0.0 || 1.1.0 || 1.2.0"

var DefaultSchemaVersion = semver.MustParse(string(BundleSchemaVersion()))

// ExtendedBundle is a bundle that has typed access to extensions declared in the bundle,
// allowing quick type-safe access to custom extensions from the CNAB spec.
type ExtendedBundle struct {
	bundle.Bundle
}

type DependencyLock struct {
	Alias        string
	Reference    string
	SharingMode  bool
	SharingGroup string
}

// NewBundle creates an ExtendedBundle from a given bundle.
func NewBundle(bundle bundle.Bundle) ExtendedBundle {
	return ExtendedBundle{bundle}
}

// LoadBundle from the specified filepath.
func LoadBundle(c *portercontext.Context, bundleFile string) (ExtendedBundle, error) {
	bunD, err := c.FileSystem.ReadFile(bundleFile)
	if err != nil {
		return ExtendedBundle{}, fmt.Errorf("cannot read bundle at %s: %w", bundleFile, err)
	}

	bun, err := bundle.Unmarshal(bunD)
	if err != nil {
		return ExtendedBundle{}, fmt.Errorf("cannot load bundle from\n%s at %s: %w", string(bunD), bundleFile, err)
	}

	return NewBundle(*bun), nil
}

func (b ExtendedBundle) Validate(cxt *portercontext.Context, strategy schema.CheckStrategy) error {
	err := b.Bundle.Validate()
	if err != nil {
		return fmt.Errorf("invalid bundle: %w", err)
	}

	supported, err := semver.NewConstraint(SupportedVersion)
	if err != nil {
		return fmt.Errorf("invalid supported version %s: %w", SupportedVersion, err)
	}
	isWarn, err := schema.ValidateSchemaVersion(strategy, supported, string(b.SchemaVersion), DefaultSchemaVersion)
	if err != nil && !isWarn {
		return err
	}

	if isWarn {
		fmt.Fprintln(cxt.Err, err)
	}

	return nil
}

// IsPorterBundle determines if the bundle was created by Porter.
func (b ExtendedBundle) IsPorterBundle() bool {
	_, madeByPorter := b.Custom[PorterExtension]
	return madeByPorter
}

// IsInternalParameter determines if the provided parameter is internal
// to Porter after analyzing the provided bundle.
func (b ExtendedBundle) IsInternalParameter(name string) bool {
	if param, exists := b.Parameters[name]; exists {
		if def, exists := b.Definitions[param.Definition]; exists {
			return def.Comment == PorterInternal
		}
	}
	return false
}

// IsInternalOutput determines if the provided output is internal
// to Porter after analyzing the provided bundle.
func (b ExtendedBundle) IsInternalOutput(name string) bool {
	if output, exists := b.Outputs[name]; exists {
		if def, exists := b.Definitions[output.Definition]; exists {
			return def.Comment == PorterInternal
		}
	}
	return false
}

// IsSensitiveParameter determines if the parameter contains a sensitive value.
func (b ExtendedBundle) IsSensitiveParameter(param string) bool {
	if param, exists := b.Parameters[param]; exists {
		if def, exists := b.Definitions[param.Definition]; exists {
			return def.WriteOnly != nil && *def.WriteOnly
		}
	}
	return false
}

// GetParameterType determines the type of parameter accounting for
// Porter-specific parameter types like file.
func (b ExtendedBundle) GetParameterType(def *definition.Schema) string {
	if b.IsFileType(def) {
		return "file"
	}

	if def.ID == claim.OutputInvocationImageLogs {
		return "string"
	}

	return fmt.Sprintf("%v", def.Type)
}

// IsFileType determines if the parameter/credential is of type "file".
func (b ExtendedBundle) IsFileType(def *definition.Schema) bool {
	return b.SupportsFileParameters() &&
		def.Type == "string" && def.ContentEncoding == "base64"
}

// ConvertParameterValue converts a parameter's value from an unknown type,
// it could be a string from stdin or another Go type, into the type of the
// parameter as defined in the bundle.
func (b ExtendedBundle) ConvertParameterValue(key string, value interface{}) (interface{}, error) {
	param, ok := b.Parameters[key]
	if !ok {
		return nil, fmt.Errorf("unable to convert the parameters' value to the destination parameter type because parameter %s not defined in bundle", key)
	}

	def, ok := b.Definitions[param.Definition]
	if !ok {
		return nil, fmt.Errorf("unable to convert the parameters' value to the destination parameter type because parameter %s has no definition", key)
	}

	if def.Type != nil {
		switch t := value.(type) {
		case string:
			typedValue, err := def.ConvertValue(t)
			if err != nil {
				return nil, fmt.Errorf("unable to convert parameter's %s value %s to the destination parameter type %s: %w", key, value, def.Type, err)
			}
			return typedValue, nil
		case json.Number:
			switch def.Type {
			case "integer":
				return t.Int64()
			case "number":
				return t.Float64()
			default:
				return t.String(), nil
			}
		default:
			return t, nil
		}
	} else {
		return value, nil
	}
}

func (b ExtendedBundle) WriteParameterToString(paramName string, value interface{}) (string, error) {
	return WriteParameterToString(paramName, value)
}

// WriteParameterToString changes a parameter's value from its type as
// defined by the bundle to its runtime string representation.
// The value should have already been converted to its bundle representation
// by calling ConvertParameterValue.
func WriteParameterToString(paramName string, value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	if stringVal, ok := value.(string); ok {
		return stringVal, nil
	}

	contents, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("could not marshal the value for parameter %s to a json string %#v: %w", paramName, value, err)
	}

	return string(contents), nil
}

// GetReferencedRegistries identifies all OCI registries used by the bundle
// from both the invocation image and the referenced images.
func (b ExtendedBundle) GetReferencedRegistries() ([]string, error) {
	regMap := make(map[string]struct{})
	for _, ii := range b.InvocationImages {
		imgRef, err := ParseOCIReference(ii.Image)
		if err != nil {
			return nil, fmt.Errorf("could not parse the bundle image %s as an OCI image reference: %w", ii.Image, err)
		}

		regMap[imgRef.Registry()] = struct{}{}
	}

	for key, img := range b.Images {
		imgRef, err := ParseOCIReference(img.Image)
		if err != nil {
			return nil, fmt.Errorf("could not parse the referenced image %s (%s) as an OCI image reference: %w", img.Image, key, err)
		}
		regMap[imgRef.Registry()] = struct{}{}
	}

	regs := make([]string, 0, len(regMap))
	for reg := range regMap {
		regs = append(regs, reg)
	}
	sort.Strings(regs)
	return regs, nil
}

func (b *ExtendedBundle) ResolveDependencies(bun ExtendedBundle) ([]DependencyLock, error) {
	if bun.HasDependenciesV2() {
		return b.ResolveSharedDeps(bun)
	}

	if !bun.HasDependenciesV1() {
		return nil, nil
	}
	rawDeps, err := bun.ReadDependenciesV1()
	// We need make sure the DependenciesV1 are ordered by the desired sequence
	orderedDeps := rawDeps.ListBySequence()

	if err != nil {
		return nil, fmt.Errorf("error executing dependencies for %s: %w", bun.Name, err)
	}

	q := make([]DependencyLock, 0, len(orderedDeps))
	for _, dep := range orderedDeps {
		ref, err := b.ResolveVersion(dep.Name, dep)
		if err != nil {
			return nil, err
		}

		lock := DependencyLock{
			Alias:       dep.Name,
			Reference:   ref.String(),
			SharingMode: false,
		}
		q = append(q, lock)
	}

	return q, nil
}

// ResolveSharedDeps only works with depsv2
func (b *ExtendedBundle) ResolveSharedDeps(bun ExtendedBundle) ([]DependencyLock, error) {
	v2, err := bun.ReadDependenciesV2()
	if err != nil {
		return nil, fmt.Errorf("error reading dependencies v2 for %s", bun.Name)
	}

	q := make([]DependencyLock, 0, len(v2.Requires))
	for name, d := range v2.Requires {
		d.Name = name

		if d.Sharing.Mode && d.Sharing.Group.Name == "" {
			return nil, fmt.Errorf("empty sharing group, sharing group name needs to be specified to be active")
		}
		if !d.Sharing.Mode && d.Sharing.Group.Name != "" {
			return nil, fmt.Errorf("empty sharing mode, sharing mode boolean set to `true` to be active")
		}

		ref, err := b.ResolveVersionv2(d.Name, d)
		if err != nil {
			return nil, err
		}

		lock := DependencyLock{
			Alias:        d.Name,
			Reference:    ref.String(),
			SharingMode:  d.Sharing.Mode,
			SharingGroup: d.Sharing.Group.Name,
		}
		q = append(q, lock)
	}
	return q, nil
}

// ResolveVersion returns the bundle name, its version and any error.
func (b *ExtendedBundle) ResolveVersion(name string, dep depsv1ext.Dependency) (OCIReference, error) {
	ref, err := ParseOCIReference(dep.Bundle)
	if err != nil {
		return OCIReference{}, fmt.Errorf("error parsing dependency (%s) bundle %q as OCI reference: %w", name, dep.Bundle, err)
	}

	// Here is where we could split out this logic into multiple strategy funcs / structs if necessary
	if dep.Version == nil || len(dep.Version.Ranges) == 0 {
		// Check if they specified an explicit tag in referenced bundle already
		if ref.HasTag() {
			return ref, nil
		}

		tag, err := b.determineDefaultTag(dep)
		if err != nil {
			return OCIReference{}, err
		}

		return ref.WithTag(tag)
	}

	return OCIReference{}, fmt.Errorf("not implemented: dependency version range specified for %s: %w", name, err)
}

func (b *ExtendedBundle) determineDefaultTag(dep depsv1ext.Dependency) (string, error) {
	tags, err := crane.ListTags(dep.Bundle)
	if err != nil {
		return "", fmt.Errorf("error listing tags for %s: %w", dep.Bundle, err)
	}

	allowPrereleases := false
	if dep.Version != nil && dep.Version.AllowPrereleases {
		allowPrereleases = true
	}

	var hasLatest bool
	versions := make(semver.Collection, 0, len(tags))
	for _, tag := range tags {
		if tag == "latest" {
			hasLatest = true
			continue
		}

		version, err := semver.NewVersion(tag)
		if err == nil {
			if !allowPrereleases && version.Prerelease() != "" {
				continue
			}
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		if hasLatest {
			return "latest", nil
		} else {
			return "", fmt.Errorf("no tag was specified for %s and none of the tags defined in the registry meet the criteria: semver formatted or 'latest'", dep.Bundle)
		}
	}

	sort.Sort(sort.Reverse(versions))

	return versions[0].Original(), nil
}

// BuildPrerequisiteInstallationName generates the name of a prerequisite dependency installation.
func (b *ExtendedBundle) BuildPrerequisiteInstallationName(installation string, dependency string) string {
	return fmt.Sprintf("%s-%s", installation, dependency)
}

// this is all copied v2 stuff
// todo(schristoff): in the future, we should clean this up

// ResolveVersion returns the bundle name, its version and any error.
func (b *ExtendedBundle) ResolveVersionv2(name string, dep v2.Dependency) (OCIReference, error) {
	ref, err := ParseOCIReference(dep.Bundle)
	if err != nil {
		return OCIReference{}, fmt.Errorf("error parsing dependency (%s) bundle %q as OCI reference: %w", name, dep.Bundle, err)
	}

	if dep.Version == "" {
		// Check if they specified an explicit tag in referenced bundle already
		if ref.HasTag() {
			return ref, nil
		}

		tag, err := b.determineDefaultTagv2(dep)
		if err != nil {
			return OCIReference{}, err
		}

		return ref.WithTag(tag)
	}
	//I think this is going to need to be smarter
	if dep.Version != "" {
		return ref, nil
	}

	return OCIReference{}, fmt.Errorf("not implemented: dependency version range specified for %s: %w", name, err)
}

func (b *ExtendedBundle) determineDefaultTagv2(dep v2.Dependency) (string, error) {
	tags, err := crane.ListTags(dep.Bundle)
	if err != nil {
		return "", fmt.Errorf("error listing tags for %s: %w", dep.Bundle, err)
	}

	var hasLatest bool
	versions := make(semver.Collection, 0, len(tags))
	for _, tag := range tags {
		if tag == "latest" {
			hasLatest = true
			continue
		}

	}
	if len(versions) == 0 {
		if hasLatest {
			return "latest", nil
		} else {
			return "", fmt.Errorf("no tag was specified for %s and none of the tags defined in the registry meet the criteria: semver formatted or 'latest'", dep.Bundle)
		}
	}

	return versions[0].Original(), nil
}
