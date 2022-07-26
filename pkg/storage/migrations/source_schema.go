package migrations

import (
	"github.com/cnabio/cnab-go/schema"
)

// SourceSchema represents the file format of Porter's v0.38 schema document
type SourceSchema struct {
	Claims      schema.Version `json:"claims"`
	Credentials schema.Version `json:"credentials"`
	Parameters  schema.Version `json:"parameters"`
}
