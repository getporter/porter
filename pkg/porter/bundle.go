package porter

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
)

//Unmarshal unmarshals a Bundle that was not signed.
func Unmarshal(data []byte) (*Bundle, error) {
	b := &Bundle{}
	return b, json.Unmarshal(data, b)
}

// ParseReader reads CNAB metadata from a JSON string
func ParseReader(r io.Reader) (Bundle, error) {
	b := Bundle{}
	err := json.NewDecoder(r).Decode(&b)
	return b, err
}

// WriteFile serializes the bundle and writes it to a file as JSON.
func (p Porter) WriteFile(b Bundle, mode os.FileMode) error {
	f, err := p.Config.FileSystem.OpenFile("bundle.json", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	defer f.Close()
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "    ")
	return enc.Encode(b)
}

// LocationRef specifies a location within the invocation package
type LocationRef struct {
	Path  string `json:"path" toml:"path"`
	Field string `json:"field" toml:"field"`
}

// BaseImage contains fields shared across image types
type BaseImage struct {
	Digest    string `json:"digest,omitempty" toml:"digest"`
	Size      uint64 `json:"size,omitempty" toml:"size"`
	Platform  string `json:"platform,omitempty" toml:"platform"`
	MediaType string `json:"mediaType,omitempty" toml:"mediaType"`
}

// ImagePlatform indicates what type of platform an image is built for
type ImagePlatform struct {
	Architecture string `json:"architecture,omitempty" toml:"architecture"`
	OS           string `json:"os,omitempty" toml:"os"`
}

// Image describes a container image in the bundle
type Image struct {
	BaseImage
	// FIXME: Is this the same as "image" on InvocationImage? Which do we prefer?
	Name string        `json:"name" toml:"name"`
	URI  string        `json:"uri" toml:"uri"`
	Refs []LocationRef `json:"refs" toml:"refs"`
}

// InvocationImage contains the image type and location for the installation of a bundle
type InvocationImage struct {
	BaseImage
	ImageType string `json:"imageType" toml:"imageType"`
	Image     string `json:"image" toml:"image"`
}

// Location provides the location where a value should be written in
// the invocation image.
//
// A location may be either a file (by path) or an environment variable.
type Location struct {
	Path                string `json:"path" toml:"path"`
	EnvironmentVariable string `json:"env" toml:"env"`
}

// Maintainer describes a code maintainer of a bundle
type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name" toml:"name"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email" toml:"email"`
	// Url is an optional URL to an address for the named maintainer
	URL string `json:"url" toml:"url"`
}

// Action describes a custom (non-core) action.
type Action struct {
	// Modifies indicates whether this action modifies the release.
	//
	// If it is possible that an action modify a release, this must be set to true.
	Modifies bool
}

// Bundle is a CNAB metadata document
type Bundle struct {
	Name             string                         `json:"name" toml:"name"`
	Version          string                         `json:"version" toml:"version"`
	Description      string                         `json:"description" toml:"description"`
	Keywords         []string                       `json:"keywords,omitempty" toml:"keywords,omitempty"`
	Maintainers      []Maintainer                   `json:"maintainers,omitempty" toml:"maintainers,omitempty"`
	InvocationImages []InvocationImage              `json:"invocationImages" toml:"invocationImages"`
	Images           []Image                        `json:"images" toml:"images"`
	Actions          map[string]Action              `json:"actions,omitempty" toml:"actions,omitempty"`
	Parameters       map[string]ParameterDefinition `json:"parameters" toml:"parameters"`
	Credentials      map[string]Location            `json:"credentials" toml:"credentials"`
}

// // ValuesOrDefaults returns parameter values or the default parameter values
// func ValuesOrDefaults(vals map[string]interface{}, b *Bundle) (map[string]interface{}, error) {
// 	res := map[string]interface{}{}
// 	for name, def := range b.Parameters {
// 		if val, ok := vals[name]; ok {
// 			if err := def.ValidateParameterValue(val); err != nil {
// 				return res, fmt.Errorf("can't use %v as value of %s: %s", val, name, err)
// 			}
// 			typedVal := def.CoerceValue(val)
// 			res[name] = typedVal
// 			continue
// 		} else if def.Required {
// 			return res, fmt.Errorf("parameter %q is required", name)
// 		}
// 		res[name] = def.DefaultValue
// 	}
// 	return res, nil
// }

// Validate the bundle contents.
func (b Bundle) Validate() error {
	if len(b.InvocationImages) == 0 {
		return errors.New("at least one invocation image must be defined in the bundle")
	}

	for _, img := range b.InvocationImages {
		err := img.Validate()
		if err != nil {
			return err
		}
	}

	return nil
}

// Validate the image contents.
func (img InvocationImage) Validate() error {
	switch img.ImageType {
	case "docker", "oci":
		return validateDockerish(img.Image)
	default:
		return nil
	}
}

func validateDockerish(s string) error {
	if !strings.Contains(s, ":") {
		return errors.New("tag is required")
	}
	return nil
}
