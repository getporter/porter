package storage

import (
	"time"

	"github.com/cnabio/cnab-go/bundle"
	"github.com/cnabio/cnab-go/schema"
)

// https://cnab.io/schema/cnab-claim-1.0.0-DRAFT+d7ffba8/claim.schema.json
type claimd7ffba8 struct {
	SchemaVersion   schema.Version         `json:"schemaVersion"`
	Installation    string                 `json:"installation"`
	Revision        string                 `json:"revision"`
	Created         time.Time              `json:"created"`
	Modified        time.Time              `json:"modified"`
	Bundle          *bundle.Bundle         `json:"bundle"`
	BundleReference string                 `json:"bundleReference,omitempty"`
	Result          resultd7ffba8          `json:"result,omitempty"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
	Outputs         map[string]interface{} `json:"outputs,omitempty"`
	Custom          interface{}            `json:"custom,omitempty"`
}

type resultd7ffba8 struct {
	Message string `json:"message,omitempty"`
	Action  string `json:"action"`
	Status  string `json:"status"`
}
