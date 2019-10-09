package builder

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
)

// Flag represents a flag passed to a mixin command.
type Flag struct {
	Name   string
	Values []string
}

type Dashes struct {
	Long  string
	Short string
}

// NewFlag creates an instance of a Flag.
func NewFlag(name string, values ...string) Flag {
	f := Flag{
		Name: name,
	}
	if len(values) > 0 {
		f.Values = make([]string, len(values))
		copy(f.Values, values)
	}
	return f
}

// ToSlice converts to a string array suitable of command arguments suitable for passing to exec.Command
func (flag Flag) ToSlice(dashes Dashes) []string {
	var flagName string
	dash := dashes.Long
	if len(flag.Name) == 1 {
		dash = dashes.Short
	}
	flagName = fmt.Sprintf("%s%s", dash, flag.Name)

	result := make([]string, 0, len(flag.Values)+1)
	if len(flag.Values) == 0 {
		result = append(result, flagName)
	} else {
		for _, value := range flag.Values {
			result = append(result, flagName)
			result = append(result, value)
		}
	}
	return result
}

type Flags []Flag

// ToSlice converts to a string array suitable of command arguments suitable for passing to exec.Command
func (flags *Flags) ToSlice(dashes Dashes) []string {
	result := make([]string, 0, 2*len(*flags))

	sort.Sort(flags)
	for _, flag := range *flags {
		result = append(result, flag.ToSlice(dashes)...)
	}

	return result
}

// UnmarshalYAML takes input like this
// flags:
//   flag1: value
//   flag2: value
//   flag3:
//   - value1
//   - value2
//
// and turns it into this:
//
// []Flags{ {"flag1", []string{"value"}}, {"flag2", []string{"value"}}, {"flag3", []string{"value1", "value2"} }
func (flags *Flags) UnmarshalYAML(unmarshal func(interface{}) error) error {
	flagMap := map[interface{}]interface{}{}
	err := unmarshal(&flagMap)
	if err != nil {
		return errors.Wrap(err, "could not unmarshal yaml into Step.Flags")
	}

	*flags = make(Flags, 0, len(flagMap))
	for k, v := range flagMap {
		f := Flag{}
		f.Name = k.(string)

		switch t := v.(type) {
		case []interface{}:
			f.Values = make([]string, len(t))
			for i := range t {
				iv, ok := t[i].(string)
				if !ok {
					return errors.Errorf("invalid yaml type for flag %s: %T", f.Name, t[i])
				}
				f.Values[i] = iv
			}
		case nil:
			// do nothing
		default:
			f.Values = make([]string, 1)
			f.Values[0] = fmt.Sprintf("%v", v)
		}

		*flags = append(*flags, f)
	}

	return nil
}

// MarshalYAML writes out flags back into the proper format for mixin flags.
// Input:
// []Flags{ {"flag1", []string{"value"}}, {"flag2", []string{"value"}}, {"flag3", []string{"value1", "value2"} }
//
// Is turned into
//
// flags:
//   flag1: value
//   flag2: value
//   flag3:
//   - value1
//   - value2
func (flags Flags) MarshalYAML() (interface{}, error) {
	result := make(map[string]interface{}, len(flags))

	for _, flag := range flags {
		if len(flag.Values) == 1 {
			result[flag.Name] = flag.Values[0]
		} else {
			result[flag.Name] = flag.Values
		}
	}

	return result, nil
}

func (flags Flags) Len() int {
	return len(flags)
}

func (flags Flags) Swap(i, j int) {
	flags[i], flags[j] = flags[j], flags[i]
}

func (flags Flags) Less(i, j int) bool {
	return flags[i].Name < flags[j].Name
}
