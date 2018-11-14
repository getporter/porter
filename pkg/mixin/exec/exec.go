package exec

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"

	"gopkg.in/yaml.v2"
)

// Exec is the logic behind the exec mixin
type Exec struct {
	In  io.Reader
	Out io.Writer
	Err io.Writer

	instruction Instruction
}

type Instruction struct {
	Name       string            `yaml:"name"`
	Command    string            `yaml:"command"`
	Arguments  []string          `yaml:"arguments"`
	Parameters map[string]string `yaml:"parameters"`
}

func (e *Exec) LoadInstruction(commandFile string) error {
	contents, err := e.getCommandFile(commandFile, e.Out)
	if err != nil {
		return fmt.Errorf("there was an error getting commands: %s", err)
	}
	return yaml.Unmarshal(contents, &e.instruction)
}

func (e *Exec) Execute() error {
	cmd := exec.Command(e.instruction.Command, e.instruction.Arguments...)
	stderr, err := cmd.StderrPipe()
	if stderr == nil || err != nil {
		return fmt.Errorf("couldnt get stderr pipe")
	}
	stdout, err := cmd.StdoutPipe()
	if stdout == nil || err != nil {
		return fmt.Errorf("couldnt get stdout pipe")
	}
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start...%s", err)
	}
	_, err = io.Copy(e.Err, stderr)
	if err != nil {
		return fmt.Errorf("error copying stderr")
	}
	_, err = io.Copy(e.Out, stdout)
	if err != nil {
		return fmt.Errorf("error copying stdout")
	}
	return nil
}

func (e *Exec) getCommandFile(commandFile string, w io.Writer) ([]byte, error) {
	if commandFile == "" {
		reader := bufio.NewReader(e.In)
		return ioutil.ReadAll(reader)
	}
	return ioutil.ReadFile(commandFile)
}
