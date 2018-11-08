package porter

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

func (p *Porter) Build() error {
	f, err := os.OpenFile("Dockerfile", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open Dockerfile: %s", err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	err = p.addDockerBaseImage(w)
	err = p.addCNAB(w)
	defer w.Flush()
	return err
}

func (p *Porter) addDockerBaseImage(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "FROM ubuntu:latest\n"); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}
	return nil
}

func (p *Porter) addCNAB(w io.Writer) error {
	if _, err := fmt.Fprintf(w, "ADD cnab/ cnab/\n"); err != nil {
		return fmt.Errorf("couldn't write docker base image: %s", err)
	}
	return nil
}
