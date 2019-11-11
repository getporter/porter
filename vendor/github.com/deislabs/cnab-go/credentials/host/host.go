package host

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/deislabs/cnab-go/credentials"
)

const (
	SourceEnv     = "env"
	SourceCommand = "command"
	SourcePath    = "path"
	SourceValue   = "value"
)

// Resolve looks up the credentials as described in Source, then copies
// the resulting value into the Value field of each credential strategy.
//
// The typical workflow for working with a credential set is:
//
//	- Load the set
//	- Validate the credentials against a spec
//	- Resolve the credentials
//	- Expand them into bundle values
func ResolveCredentials(c credentials.CredentialSet) (credentials.Set, error) {
	l := len(c.Credentials)
	res := make(map[string]string, l)
	for i := 0; i < l; i++ {
		cred := c.Credentials[i]
		src := cred.Source
		// Precedence is command, path, env, value
		switch strings.ToLower(src.Key) {
		case SourceCommand:
			data, err := execCmd(src.Value)
			if err != nil {
				return res, err
			}
			cred.Value = string(data)
		case SourcePath:
			data, err := ioutil.ReadFile(os.ExpandEnv(src.Value))
			if err != nil {
				return res, fmt.Errorf("credential %q: %s", c.Credentials[i].Name, err)
			}
			cred.Value = string(data)
		case SourceEnv:
			var ok bool
			cred.Value, ok = os.LookupEnv(src.Value)
			if ok {
				break
			}
			fallthrough
		case SourceValue:
			cred.Value = src.Value
		default:
			return nil, fmt.Errorf("invalid credential source: %s", src.Key)
		}
		res[c.Credentials[i].Name] = cred.Value
	}
	return res, nil
}

func execCmd(cmd string) ([]byte, error) {
	parts := strings.Split(cmd, " ")
	c := parts[0]
	args := parts[1:]
	run := exec.Command(c, args...)

	return run.CombinedOutput()
}
