package pkg

import "os"

// These are build-time values, set during an official release
var (
	Commit  string
	Version string
)

const (
	// FileModeDirectory is the FileMode that should be used when creating a
	// directory. It ensures that both the user and the group have the same
	// permissions.
	FileModeDirectory os.FileMode = 0770

	// FileModeWritable is the FileMode that should be used when creating a file
	// that should have read/write permissions. It ensures that both the user and
	// the group have the same permissions.
	FileModeWritable os.FileMode = 0660

	// FileModeExecutable is the FileMode that should be used when creating an
	// executable file, such as a script or binary. It ensures that both the user
	// and the group have the same permissions.
	FileModeExecutable os.FileMode = 0770
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
