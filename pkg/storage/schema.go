package storage

import "github.com/cnabio/cnab-go/schema"

var _ Document = Schema{}

// TODO(carolynvs): do we want to use porter schema versions instead of relying on cnab?
type Schema struct {
	ID          string         `json:"_id"`
	Claims      schema.Version `json:"claims"`
	Credentials schema.Version `json:"credentials"`
	Parameters  schema.Version `json:"parameters"`
}

func NewSchema(claims schema.Version, creds schema.Version, params schema.Version) Schema {
	return Schema{
		ID:          "schema",
		Claims:      claims,
		Credentials: creds,
		Parameters:  params,
	}
}

func (s Schema) DefaultDocumentFilter() interface{} {
	return map[string]interface{}{"_id": "schema"}
}
