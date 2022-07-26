package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cnabio/cnab-go/bundle"
)

var _ json.Marshaler = BundleDocument{}
var _ json.Unmarshaler = &BundleDocument{}

// BundleDocument is the storage representation of a Bundle in mongo (as a string containing quoted json).
type BundleDocument bundle.Bundle

// MarshalJSON converts the bundle to a json string.
func (b BundleDocument) MarshalJSON() ([]byte, error) {
	// Write the bundle definition to a string, so we can
	// store it in mongo
	var bunData bytes.Buffer
	bun := bundle.Bundle(b)
	if _, err := bun.WriteTo(&bunData); err != nil {
		return nil, fmt.Errorf("error marshaling Bundle into its storage representation: %w", err)
	}

	bunStr := strconv.Quote(bunData.String())
	return []byte(bunStr), nil
}

// UnmarshalJSON converts the bundle from a json string.
func (b *BundleDocument) UnmarshalJSON(data []byte) error {
	jsonData, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("error unquoting Bundle from its storage representation: %w", err)
	}
	var rawBun bundle.Bundle
	if err := json.Unmarshal([]byte(jsonData), &rawBun); err != nil {
		return fmt.Errorf("error unmarshaling Bundle from its storage representation: %w", err)
	}

	*b = BundleDocument(rawBun)
	return nil
}
