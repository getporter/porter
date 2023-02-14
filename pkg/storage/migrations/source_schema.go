package migrations

import (
	"get.porter.sh/porter/pkg/cnab"
)

// SourceSchema represents the file format of Porter's v0.38 schema document
type SourceSchema struct {
	Claims      cnab.SchemaVersion `json:"claims"`
	Credentials cnab.SchemaVersion `json:"credentials"`
	Parameters  cnab.SchemaVersion `json:"parameters"`
}
