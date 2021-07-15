package claims

// Provider is an interface for interacting with Porter's claim data.
type Provider interface {
	InsertInstallation(installation Installation) error
	InsertRun(run Run) error
	InsertResult(result Result) error
	InsertOutput(output Output) error

	UpdateInstallation(installation Installation) error
	UpsertRun(run Run) error
	UpsertInstallation(installation Installation) error

	GetInstallation(namespace string, name string) (Installation, error)

	// ListInstallations returns Installations sorted in ascending order.
	ListInstallations(namespace string) ([]Installation, error)

	ListRuns(namespace string, installation string) ([]Run, error)
	ListResults(runID string) ([]Result, error)
	ListOutputs(resultID string) ([]Output, error)

	GetRun(id string) (Run, error)
	GetResult(id string) (Result, error)

	/*
		// ListClaims returns ToCNAB IDs associated with an Installation sorted in ascending order.
		ListClaims(installation string) ([]string, error)

		// ListResults returns Result IDs associated with a ToCNAB, sorted in ascending order.
		ListResults(claimID string) ([]string, error)

		// ListOutputs returns the names of outputs associated with a result that
		// have been persisted. It is possible for results to have outputs that were
		// generated but not persisted see
		// https://github.com/cnabio/cnab-spec/blob/cnab-claim-1.0.0-DRAFT+b5ed2f3/400-claims.md#outputs
		// for more information.
		ListOutputs(resultID string) ([]string, error)

		// ReadInstallation returns the specified Installation with all Claims and their Results loaded.
		ReadInstallation(installation string) (Installation, error)

		// ReadInstallationStatus returns the specified Installation with the last ToCNAB and its last Result loaded.
		ReadInstallationStatus(installation string) (Installation, error)

		// ReadAllInstallationStatus returns all Installations with the last ToCNAB and its last Result loaded.
		ReadAllInstallationStatus() ([]Installation, error)

		// ReadClaim returns the specified ToCNAB.
		ReadClaim(claimID string) (ToCNAB, error)

		// ReadAllClaims returns all claims associated with an Installation, sorted in ascending order.
		ReadAllClaims(installation string) ([]ToCNAB, error)
	*/
	// GetLastRun returns the last run of an Installation.
	GetLastRun(namespace string, installation string) (Run, error)

	// GetLastOutput returns the most recent value (last) of the specified
	// Output associated with the installation.
	GetLastOutput(namespace string, installation string, name string) (Output, error)

	// GetLastOutputs returns the most recent (last) value of each Output
	// associated with the installation.
	GetLastOutputs(namespace string, installation string) (Outputs, error)

	/*
		// ReadResult returns the specified Result.
		ReadResult(resultID string) (Result, error)

		// ReadAllResult returns all results associated with a ToCNAB, sorted in ascending order.
		ReadAllResults(claimID string) ([]Result, error)

		// ReadLastResult returns the last result associated with a ToCNAB.
		ReadLastResult(claimID string) (Result, error)

		// ReadAllOutputs returns the most recent (last) value of each Output associated
		// with the installation.
		ReadLastOutputs(installation string) (Outputs, error)

		// GetLastOutput returns the most recent value (last) of the specified Output associated
		// with the installation.
		GetLastOutput(installation string, name string) (Output, error)

		// ReadOutput returns the contents of the named output associated with the specified Result.
		ReadOutput(claim ToCNAB, result Result, outputName string) (Output, error)

		// SaveClaim persists the specified claim.
		// Associated results, ToCNAB.Results, must be persisted separately with SaveResult.
		SaveClaim(claim ToCNAB) error

		// SaveResult persists the specified result.
		SaveResult(result Result) error

		// SaveOutput persists the output, encrypting the value if defined as
		// sensitive (write-only) in the bundle.
		SaveOutput(output Output) error

		// DeleteInstallation removes all data associated with the specified installation.
		DeleteInstallation(installation string) error

		// DeleteClaim removes all data associated with the specified claim.
		DeleteClaim(claimID string) error

		// DeleteResult removes all data associated with the specified result.
		DeleteResult(resultID string) error

		// DeleteOutput removes an output persisted with the specified result.
		DeleteOutput(resultID string, outputName string) error
	*/

	RemoveInstallation(namespace string, name string) error
	GetLogs(runID string) (logs string, hasLogs bool, err error)
	GetLastLogs(namespace string, installation string) (logs string, hasLogs bool, err error)
}
