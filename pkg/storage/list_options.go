package storage

// ListOptions is the set of options available to the list operation
// on any storage provider.
type ListOptions struct {
	// Namespace in which the particular result list is defined.
	Namespace string

	// Name specifies whether the result list name contain the specified substring.
	Name string

	// Labels is used to filter result list based on a key-value pair.
	Labels map[string]string

	// Skip is the number of results to skip past and exclude from the results.
	Skip int64

	// Limit is the number of results to return.
	Limit int64
}
