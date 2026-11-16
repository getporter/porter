package storage

import "fmt"

// ErrNotFound indicates that the requested document was not found.
// You can test for this error using errors.Is(err, storage.ErrNotFound{})
type ErrNotFound struct {
	Collection string
	Item       string
}

func (e ErrNotFound) Error() string {
	var docType string
	switch e.Collection {
	case "installations":
		docType = "Installation"
	case "runs":
		docType = "Run"
	case "results":
		docType = "Result"
	case "output":
		docType = "Output"
	case "credentials", "parameters":
		if len(e.Item) > 0 {
			docType = e.Item
		}
	}

	return fmt.Sprintf("%s not found", docType)
}

func (e ErrNotFound) Is(err error) bool {
	_, ok := err.(ErrNotFound)
	return ok
}
