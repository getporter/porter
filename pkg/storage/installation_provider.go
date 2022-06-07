package storage

import (
	"context"
)

// InstallationProvider is an interface for interacting with Porter's claim data.
type InstallationProvider interface {
	// InsertInstallation saves a new Installation document.
	InsertInstallation(ctx context.Context, installation Installation) error

	// InsertRun saves a new Run document.
	InsertRun(ctx context.Context, run Run) error

	// InsertResult saves a new Result document.
	InsertResult(ctx context.Context, result Result) error

	// InsertOutput saves a new Output document.
	InsertOutput(ctx context.Context, output Output) error

	// UpdateInstallation saves changes to an existing Installation document.
	UpdateInstallation(ctx context.Context, installation Installation) error

	// UpsertRun saves changes a Run document, creating it if it doesn't already exist.
	UpsertRun(ctx context.Context, run Run) error

	// UpsertInstallation saves an Installation document, creating it if it doesn't already exist.
	UpsertInstallation(ctx context.Context, installation Installation) error

	// FindInstallations applies the find operation against installations collection
	// using the specified options.
	FindInstallations(ctx context.Context, opts FindOptions) ([]Installation, error)

	// GetInstallation retrieves an Installation document by name.
	GetInstallation(ctx context.Context, namespace string, name string) (Installation, error)

	// ListInstallations returns Installations sorted in ascending order by the namespace and then name.
	ListInstallations(ctx context.Context, namespace string, name string, labels map[string]string, skip int64, limit int64) ([]Installation, error)

	// ListRuns returns Run documents sorted in ascending order by ID.
	ListRuns(ctx context.Context, namespace string, installation string) ([]Run, map[string][]Result, error)

	// ListResults returns Result documents sorted in ascending order by ID.
	ListResults(ctx context.Context, runID string) ([]Result, error)

	// ListOutputs returns Output documents sorted in ascending order by name.
	ListOutputs(ctx context.Context, resultID string) ([]Output, error)

	// GetRun returns a Run document by ID.
	GetRun(ctx context.Context, id string) (Run, error)

	// GetResult returns a Result document by ID.
	GetResult(ctx context.Context, id string) (Result, error)

	// GetLastRun returns the last run of an Installation.
	GetLastRun(ctx context.Context, namespace string, installation string) (Run, error)

	// GetLastOutput returns the most recent value (last) of the specified
	// Output associated with the installation.
	GetLastOutput(ctx context.Context, namespace string, installation string, name string) (Output, error)

	// GetLastOutputs returns the most recent (last) value of each Output
	// associated with the installation.
	GetLastOutputs(ctx context.Context, namespace string, installation string) (Outputs, error)

	// RemoveInstallation by its name.
	RemoveInstallation(ctx context.Context, namespace string, name string) error

	// GetLogs returns the logs from the specified Run.
	GetLogs(ctx context.Context, runID string) (logs string, hasLogs bool, err error)

	// GetLastLogs returns the logs from the last run of an Installation.
	GetLastLogs(ctx context.Context, namespace string, installation string) (logs string, hasLogs bool, err error)
}
