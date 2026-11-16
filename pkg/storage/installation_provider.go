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

	// FindInstallationsUsingCredentialSet finds the installations that use the named credential set.
	// When a namespace is specified, only installations in that namespace are returned.
	FindInstallationsUsingCredentialSet(ctx context.Context, namespace string, name string) ([]Installation, error)

	// FindInstallationsUsingParameterSet finds the installations that use the named parameter set.
	// When a namespace is specified, only installations in that namespace are returned.
	FindInstallationsUsingParameterSet(ctx context.Context, namespace string, name string) ([]Installation, error)

	// FindInstallationsReferencing finds the installations that depend on the specified
	// installation, identified by its "namespace/name".
	FindInstallationsReferencing(ctx context.Context, installation string) ([]Installation, error)

	// FindDependencyCandidates finds the installations of a bundle repository that may be
	// shared with a dependent bundle, preferring those in the specified namespace over global ones.
	// The digest is optional.
	FindDependencyCandidates(ctx context.Context, namespace string, repository string, digest string) ([]Installation, error)

	// GetInstallation retrieves an Installation document by name.
	GetInstallation(ctx context.Context, namespace string, name string) (Installation, error)

	// ListInstallations returns Installations sorted in ascending order by the namespace and then name.
	ListInstallations(ctx context.Context, listOption ListOptions) ([]Installation, error)

	// ListRuns returns Run documents sorted in ascending order by ID.
	ListRuns(ctx context.Context, namespace string, installation string) ([]Run, map[string][]Result, error)

	// GetActiveRuns returns runs for the installation that have not yet
	// reached a terminal status (succeeded, failed, canceled), for example a
	// run with no result yet, or whose last result is pending/running/unknown.
	GetActiveRuns(ctx context.Context, namespace string, installation string) ([]Run, error)

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

	// GetOutputs returns all outputs associated with the specified run.
	GetOutputs(ctx context.Context, runID string) (Outputs, error)

	// GetOutput returns a specific output by name from the specified run.
	GetOutput(ctx context.Context, runID string, name string) (Output, error)

	// RemoveInstallation by its name.
	RemoveInstallation(ctx context.Context, namespace string, name string) error

	// GetLogs returns the logs from the specified Run.
	GetLogs(ctx context.Context, runID string) (logs string, hasLogs bool, err error)

	// GetLastLogs returns the logs from the last run of an Installation.
	GetLastLogs(ctx context.Context, namespace string, installation string) (logs string, hasLogs bool, err error)
}
