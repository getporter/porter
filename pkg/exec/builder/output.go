package builder

import (
	"errors"
)

type Output interface {
	GetName() string
	GetMetadata() OutputMetadata
}

type StepWithOutputs interface {
	GetOutputs() []Output
}

type OutputMetadata struct {
	Sensitive bool
}

type StepOutputMeta map[string]OutputMetadata

func NewStepOutputMeta(swo StepWithOutputs) StepOutputMeta {
	metadata := make(StepOutputMeta)

	for _, o := range swo.GetOutputs() {
		metadata[o.GetName()] = o.GetMetadata()
	}

	return metadata
}

func (s StepOutputMeta) Get(name string) (OutputMetadata, bool) {
	o, ok := s[name]
	return o, ok
}

func (s StepOutputMeta) Add(o Output) error {
	if _, ok := s.Get(o.GetName()); ok {
		return errors.New("output metadata already exist")
	}
	s[o.GetName()] = o.GetMetadata()
	return nil
}
