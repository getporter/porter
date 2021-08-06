package claims

// Provider is an interface for interacting with Porter's claim data.
type Provider interface {
	// InsertInstallation saves a new Installation document.
	InsertInstallation(installation Installation) error

	// InsertRun saves a new Run document.
	InsertRun(run Run) error

	// InsertResult saves a new Result document.
	InsertResult(result Result) error

	// InsertOutput saves a new Output document.
	InsertOutput(output Output) error

	// UpdateInstallation saves changes to an existing Installation document.
	UpdateInstallation(installation Installation) error

	// UpsertRun saves changes a Run document, creating it if it doesn't already exist.
	UpsertRun(run Run) error

	// UpsertInstallation saves an Installation document, creating it if it doesn't already exist.
	UpsertInstallation(installation Installation) error

	// GetInstallation retrieves an Installation document by name.
	GetInstallation(namespace string, name string) (Installation, error)

	// ListInstallations returns Installation documents sorted in ascending order by name.
	ListInstallations(namespace string) ([]Installation, error)

	// ListRuns returns Run documents sorted in ascending order by ID.
	ListRuns(namespace string, installation string) ([]Run, error)

	// ListResults returns Result documents sorted in ascending order by ID.
	ListResults(runID string) ([]Result, error)

	// ListOutputs returns Output documents sorted in ascending order by name.
	ListOutputs(resultID string) ([]Output, error)

	// GetRun returns a Run document by ID.
	GetRun(id string) (Run, error)

	// GetResult returns a Result document by ID.
	GetResult(id string) (Result, error)

	// GetLastRun returns the last run of an Installation.
	GetLastRun(namespace string, installation string) (Run, error)

	// GetLastOutput returns the most recent value (last) of the specified
	// Output associated with the installation.
	GetLastOutput(namespace string, installation string, name string) (Output, error)

	// GetLastOutputs returns the most recent (last) value of each Output
	// associated with the installation.
	GetLastOutputs(namespace string, installation string) (Outputs, error)

	// RemoveInstallation by its name.
	RemoveInstallation(namespace string, name string) error

	// GetLogs returns the logs from the specified Run.
	GetLogs(runID string) (logs string, hasLogs bool, err error)

	// GetLastLogs returns the logs from the last run of an Installation.
	GetLastLogs(namespace string, installation string) (logs string, hasLogs bool, err error)
}
