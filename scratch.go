package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Type represents the type of a mount.
type Type string

// Type constants
const (
	// TypeBind is the type for mounting host dir
	TypeBind Type = "bind"
	// TypeVolume is the type for remote storage volumes
	TypeVolume Type = "volume"
	// TypeTmpfs is the type for mounting tmpfs
	TypeTmpfs Type = "tmpfs"
	// TypeNamedPipe is the type for mounting Windows named pipes
	TypeNamedPipe Type = "npipe"
)

// Mount represents a mount (volume).
type Mount struct {
	Type Type `json:""`
	// Source specifies the name of the mount. Depending on mount type, this
	// may be a volume name or a host path, or even ignored.
	// Source is not supported for tmpfs (must be an empty value)
	Source      string      `json:""`
	Target      string      `json:""`
	ReadOnly    bool        `json:""`
	Consistency Consistency `json:""`

	BindOptions   *BindOptions   `json:""`
	VolumeOptions *VolumeOptions `json:""`
	TmpfsOptions  *TmpfsOptions  `json:""`
}

// Propagation represents the propagation of a mount.
type Propagation string

const (
	// PropagationRPrivate RPRIVATE
	PropagationRPrivate Propagation = "rprivate"
	// PropagationPrivate PRIVATE
	PropagationPrivate Propagation = "private"
	// PropagationRShared RSHARED
	PropagationRShared Propagation = "rshared"
	// PropagationShared SHARED
	PropagationShared Propagation = "shared"
	// PropagationRSlave RSLAVE
	PropagationRSlave Propagation = "rslave"
	// PropagationSlave SLAVE
	PropagationSlave Propagation = "slave"
)

// Propagations is the list of all valid mount propagations
var Propagations = []Propagation{
	PropagationRPrivate,
	PropagationPrivate,
	PropagationRShared,
	PropagationShared,
	PropagationRSlave,
	PropagationSlave,
}

// Consistency represents the consistency requirements of a mount.
type Consistency string

const (
	// ConsistencyFull guarantees bind mount-like consistency
	ConsistencyFull Consistency = "consistent"
	// ConsistencyCached mounts can cache read data and FS structure
	ConsistencyCached Consistency = "cached"
	// ConsistencyDelegated mounts can cache read and written data and structure
	ConsistencyDelegated Consistency = "delegated"
	// ConsistencyDefault provides "consistent" behavior unless overridden
	ConsistencyDefault Consistency = "default"
)

// BindOptions defines options specific to mounts of type "bind".
type BindOptions struct {
	Propagation  Propagation `json:""`
	NonRecursive bool        `json:""`
}

// VolumeOptions represents the options for a mount of type volume.
type VolumeOptions struct {
	NoCopy       bool              `json:""`
	Labels       map[string]string `json:""`
	DriverConfig *Driver           `json:""`
}

// Driver represents a volume driver.
type Driver struct {
	Name    string            `json:""`
	Options map[string]string `json:""`
}

// TmpfsOptions defines options specific to mounts of type "tmpfs".
type TmpfsOptions struct {
	// Size sets the size of the tmpfs, in bytes.
	//
	// This will be converted to an operating system specific value
	// depending on the host. For example, on linux, it will be converted to
	// use a 'k', 'm' or 'g' syntax. BSD, though not widely supported with
	// docker, uses a straight byte value.
	//
	// Percentages are not supported.
	SizeBytes int64 `json:""`
	// Mode of the tmpfs upon creation
	Mode os.FileMode `json:""`

	// TODO(stevvooe): There are several more tmpfs flags, specified in the
	// daemon, that are accepted. Only the most basic are added for now.
	//
	// From https://github.com/moby/sys/blob/mount/v0.1.1/mount/flags.go#L47-L56
	//
	// var validFlags = map[string]bool{
	// 	"":          true,
	// 	"size":      true, X
	// 	"mode":      true, X
	// 	"uid":       true,
	// 	"gid":       true,
	// 	"nr_inodes": true,
	// 	"nr_blocks": true,
	// 	"mpol":      true,
	// }
	//
	// Some of these may be straightforward to add, but others, such as
	// uid/gid have implications in a clustered system.
}

type PortBinding struct {
	// HostIP is the host IP Address
	HostIP string `json:"HostIp"`
	// HostPort is the host port number
	HostPort string
}

// PortMap is a collection of PortBinding indexed by Port
type PortMap map[Port][]PortBinding

// PortSet is a collection of structs indexed by Port
type PortSet map[Port]struct{}

// Port is a string containing port number and protocol in the format "80/tcp"
type Port string

// RestartPolicy represents the restart policies of the container.
type RestartPolicy struct {
	Name              string
	MaximumRetryCount int
}

// Docker describes the set of custom extension metadata associated with the Docker extension
type Docker struct {
	// Privileged represents whether or not the Docker container should run as --privileged
	Privileged bool `json:"privileged"`
	// Mounts represent mounts to be attached to the host machine with all configurable options.
	Mounts []Mount `json:"mounts"`
	// Network represents the network type applied to the container "host,bridged,etc"
	Network string `json:"network"`
	// CapAdd represents the capabilities available to the container kernel
	CapAdd []string `json:"capadd"`
	// CapDrop represents capabilities to exclude from the container kernel
	CapDrop []string `json:"capdrop"`
	// Ports to bind between the host and the container
	PortBindings []PortMap `json:"portBindings"`
	// Restart policy to be used for the container
	// This may be useful in some rare cases
	RestartPolicy RestartPolicy `json:"restartPolicy"`
}

func PrettyStruct(data interface{}) (string, error) {
	val, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(val), nil
}

func main() {
	docker := Docker{
		Privileged: false,
		Mounts: []Mount{
			{
				Type:        "",
				Source:      "",
				Target:      "",
				ReadOnly:    false,
				Consistency: "",
				BindOptions: &BindOptions{
					Propagation:  "",
					NonRecursive: false,
				},
				VolumeOptions: &VolumeOptions{
					NoCopy: false,
					Labels: map[string]string{
						"": "",
					},
					DriverConfig: &Driver{
						Name: "",
						Options: map[string]string{
							"": "",
						},
					},
				},
				TmpfsOptions: &TmpfsOptions{
					SizeBytes: 0,
					Mode:      0,
				},
			},
		},
		Network: "",
		CapAdd: []string{
			"c1",
			"c2",
		},
		CapDrop: []string{
			"c3",
			"c4",
		},
		PortBindings: []PortMap{
			{
				"80": []PortBinding{
					{
						HostIP:   "0.0.0.0",
						HostPort: "8080",
					},
				},
			},
		},
		RestartPolicy: RestartPolicy{
			Name:              "",
			MaximumRetryCount: 0,
		},
	}

	fmt.Println(PrettyStruct(docker))
}
