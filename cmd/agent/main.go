package main

import (
	"fmt"
	"os"

	"get.porter.sh/porter/pkg/agent"
)

// The porter agent wraps the porter cli,
// handling copying config files from a mounted
// volume into PORTER_HOME
func main() {
	porterHome := os.Getenv("PORTER_HOME")
	if porterHome == "" {
		porterHome = "/app/.porter"
	}
	porterConfig := os.Getenv("PORTER_CONFIG")
	if porterConfig == "" {
		porterConfig = "/porter-config"
	}
	err, run := agent.Execute(os.Args[1:], porterHome, porterConfig)
	if err != nil {
		if !run {
			fmt.Fprintln(os.Stderr, err)
		}

		os.Exit(1)
	}
}
