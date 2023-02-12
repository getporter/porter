package cnab

import (
	cnabclaims "github.com/cnabio/cnab-go/claim"
)

// IDGenerator is a test friendly interface for swapping out how we generate IDs.
type IDGenerator interface {
	// NewID returns a new unique ID.
	NewID() string
}

// ULIDGenerator creates IDs that are ULIDs.
type ULIDGenerator struct{}

func (g ULIDGenerator) NewID() string {
	return cnabclaims.MustNewULID()
}
