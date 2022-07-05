package storage

import (
	"time"

	"get.porter.sh/porter/pkg/cnab"
	"github.com/cnabio/cnab-go/schema"
)

var _ Document = Result{}

type Result struct {
	// SchemaVersion of the document.
	SchemaVersion schema.Version `json:"schemaVersion"`

	// ID of the result.
	ID string `json:"_id"`

	// Created timestamp of the result.
	Created time.Time `json:"created"`

	// Namespace of the installation.
	Namespace string `json:"namespace"`

	// Installation name that owns this result.
	Installation string `json:"installation"`

	// RunID of the run that generated this result.
	RunID string `json:"runId"`

	// Message communicates the outcome of the operation.
	Message string `json:"message,omitempty"`

	// Status of the operation, for example StatusSucceeded.
	Status string `json:"status"`

	// OutputMetadata generated by the operation, mapping from the output names to
	// metadata about the output.
	OutputMetadata cnab.OutputMetadata `json:"outputs"`

	// Custom extension data applicable to a given runtime.
	Custom interface{} `json:"custom,omitempty"`
}

func (r Result) DefaultDocumentFilter() map[string]interface{} {
	return map[string]interface{}{"_id": r.ID}
}

func NewResult() Result {
	return Result{
		SchemaVersion: InstallationSchemaVersion,
		ID:            cnab.NewULID(),
		Created:       time.Now(),
	}
}

func (r Result) NewOutput(name string, data []byte) Output {
	return Output{
		SchemaVersion: InstallationSchemaVersion,
		Name:          name,
		Namespace:     r.Namespace,
		Installation:  r.Installation,
		RunID:         r.RunID,
		ResultID:      r.ID,
		Value:         data,
	}
}
