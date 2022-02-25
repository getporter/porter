//go:build docs
// +build docs

package main

func init() {
	// This is a sad hack, if we really want conditional builds of the CLI, there are other ways
	includeDocsCommand = true
}
