package main

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed schema.json
var schema string

//go:embed version.txt
var version string

func main() {
	if len(os.Args) < 2 {
		return
	}

	command := os.Args[1]
	switch command {
	case "version":
		fmt.Println(version)
	case "schema":
		// This is a mixin that helps us test out our schema command
		fmt.Println(schema)
	case "lint":
		// The test mixin does not implement lint
		fmt.Fprintln(os.Stderr, `unknown command "lint" for "testmixin"`)
		os.Exit(1)
	case "build":
		fmt.Println("# testmixin")
	case "run":
		fmt.Println("running testmixin...")
	default:
		fmt.Println("unsupported command:", command)
		os.Exit(1)
	}
}
