package storage

import (
	"github.com/cnabio/cnab-go/schema"
	"go.mongodb.org/mongo-driver/bson"
)

var _ Document = Schema{}

type Schema struct {
	ID string `json:"_id"`

	// Installations is the schema for the installation documents.
	Installations schema.Version `json:"installations"`

	// Claims is the schema for the old CNAB claim spec. DEPRECATED.
	Claims schema.Version `json:"claims,omitempty"`

	// Credentials is the schema for the credential spec documents.
	Credentials schema.Version `json:"credentials"`

	// Parameters is the schema for the parameter spec documents.
	Parameters schema.Version `json:"parameters"`
}

func NewSchema(installations schema.Version, creds schema.Version, params schema.Version) Schema {
	return Schema{
		ID:            "schema",
		Installations: installations,
		Credentials:   creds,
		Parameters:    params,
	}
}

func (s Schema) DefaultDocumentFilter() bson.M {
	return bson.M{"_id": "schema"}
}
