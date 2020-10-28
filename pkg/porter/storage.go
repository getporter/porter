package porter

import (
	"errors"
	"fmt"
)

func (p *Porter) MigrateStorage() error {
	logfilePath, err := p.Storage.Migrate()

	fmt.Fprintf(p.Out, "\nSaved migration logs to %s\n", logfilePath)

	if err != nil {
		// The error has already been printed, don't return it otherwise it will be double printed
		return errors.New("Migration failed!")
	}

	fmt.Fprintln(p.Out, "Migration complete!")
	return nil
}
