package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/deislabs/porter/pkg/mixin/exec"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var (
	commandFile string
)

func buildInstallCommand(e *exec.Exec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Execute the install functionality of this mixin",
		RunE: func(cmd *cobra.Command, args []string) error {
			contents, err := getCommandFile(commandFile, e.Out)
			if err != nil {
				fmt.Printf("there was an error: %s", err)
				return err
			}
			var payload exec.Instruction
			err = yaml.Unmarshal(contents, &payload)
			if err != nil {
				return err
			}
			return e.Execute(payload)
		},
	}
	flags := cmd.Flags()
	flags.StringVarP(&commandFile, "file", "f", "", "file to install")
	return cmd
}

func getCommandFile(commandFile string, w io.Writer) ([]byte, error) {
	if commandFile == "" {
		reader := bufio.NewReader(os.Stdin)
		return ioutil.ReadAll(reader)
	}
	return ioutil.ReadFile(commandFile)
}
