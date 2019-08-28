package action

import (
	"github.com/deislabs/cnab-go/driver"
)

type OperationConfigFunc func(op *driver.Operation) error

// ConfigurableAction is used to define actions which can configure operations before they are executed.
type ConfigurableAction struct {
	// OperationConfig is an optional handler that applies additional
	// configuration to an operation before it is executed.
	OperationConfig OperationConfigFunc
}

// ApplyConfig safely applies the configuration function to the operation, if defined,
// and returns any error.
func (a ConfigurableAction) ApplyConfig(op *driver.Operation) error {
	if a.OperationConfig != nil {
		return a.OperationConfig(op)
	}
	return nil
}
