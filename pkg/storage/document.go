package storage

// Document represents a stored Porter document with
// accessor methods to make persistence more straightforward.
type Document interface {
	// DefaultDocumentFilter is the default filter to match the current document.
	DefaultDocumentFilter() map[string]interface{}
}
