package exec

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

func (e *Exec) Install(commandFile string) error {
	contents, err := getCommandFile(commandFile, e.Out)
	if err != nil {
		return fmt.Errorf("there was an error getting commands: %s", err)
	}
	var payload Instruction
	err = yaml.Unmarshal(contents, &payload)
	if err != nil {
		return err
	}
	return e.Execute(payload)
}

func getCommandFile(commandFile string, w io.Writer) ([]byte, error) {
	if commandFile == "" {
		reader := bufio.NewReader(os.Stdin)
		return ioutil.ReadAll(reader)
	}
	return ioutil.ReadFile(commandFile)
}
