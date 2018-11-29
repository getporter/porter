package porter

import (
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"github.com/deislabs/porter/pkg/config"

	"github.com/mitchellh/reflectwalk"

	"github.com/pkg/errors"
)

func (p *Porter) resolveSourcedValues(s *config.Step) error {
	return reflectwalk.Walk(s, p)
}

func (p *Porter) Map(val reflect.Value) error {
	return nil
}

func (p *Porter) MapElem(m, k, v reflect.Value) error {
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	// If the value is is a map, check to see if it's a
	// single entry map with the key "source".
	if kind := v.Kind(); kind == reflect.Map {
		if len(v.MapKeys()) == 1 {
			sk := v.MapKeys()[0]
			if sk.Kind() == reflect.Interface {
				sk = sk.Elem()
			}
			//if the key is a string, and the string is source, then we should try
			//and replace this
			if sk.Kind() == reflect.String && sk.String() == "source" {
				kv := v.MapIndex(sk)
				if kv.Kind() == reflect.Interface {
					kv = kv.Elem()
					value := kv.String()
					replacement, err := p.resolveValue(value)
					if err != nil {
						return errors.Wrap(err, "unable to source value")
					}
					m.SetMapIndex(k, reflect.ValueOf(replacement))
				}
			}
		}
	}
	return nil
}

func (p *Porter) Slice(val reflect.Value) error {
	return nil
}

func (p *Porter) SliceElem(index int, val reflect.Value) error {
	v, ok := val.Interface().(string)
	if ok {
		//if the array entry is a string that matches source:...., we should replace it
		re := regexp.MustCompile("source:\\s?(.*)")
		matches := re.FindStringSubmatch(v)
		if len(matches) > 0 {
			source := matches[1]
			r, err := p.resolveValue(source)
			if err != nil {
				return errors.Wrap(err, "unable to source value")
			}
			val.Set(reflect.ValueOf(r))
		}
	}
	return nil
}

func (p *Porter) resolveValue(key string) (interface{}, error) {
	source := strings.Split(key, ".")
	var replacement interface{}
	if source[1] == "parameters" {
		for _, param := range p.Config.Manifest.Parameters {
			if param.Name == source[2] {
				if param.Destination == nil {
					// Porter by default sets CNAB params to name.ToUpper()
					pe := strings.ToUpper(source[2])
					replacement = os.Getenv(pe)
				} else if param.Destination.EnvironmentVariable != "" {
					replacement = os.Getenv(param.Destination.EnvironmentVariable)
				} else if param.Destination == nil && param.Destination.Path != "" {
					replacement = param.Destination.Path
				} else {
					return nil, errors.New("unknown parameter definition")
				}
			}
		}
	} else if source[1] == "credentials" {
		for _, cred := range p.Config.Manifest.Credentials {
			if cred.Name == source[2] {
				if cred.Path != "" {
					replacement = cred.Path
				} else if cred.EnvironmentVariable != "" {
					replacement = os.Getenv(cred.EnvironmentVariable)
				} else {
					return nil, errors.New("unknown credential definition")
				}
			}
		}
	} else {
		return nil, errors.New(fmt.Sprintf("unknown parameter source: %s", source[1]))
	}
	if replacement == nil {
		return nil, errors.New("unable to find parameter")
	}
	return replacement, nil
}
