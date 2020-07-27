package storage

import "github.com/cnabio/cnab-go/schema"

type Schema struct {
	Claims      schema.Version `json:"claims"`
	Credentials schema.Version `json:"credentials"`
}
