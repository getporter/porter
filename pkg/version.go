package pkg

// These are build-time values, set during an official release
var (
	Commit  string
	Version string
)

// UserAgent returns a string that can be used as a user agent for porter.
func UserAgent() string {
	product := "porter"

	if Commit == "" && Version == "" {
		return product
	}

	v := Version
	if v == "" {
		v = Commit
	}

	return product + "/" + v
}
