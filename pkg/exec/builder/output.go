package builder

type Output interface {
	GetName() string
	IsSensitive() bool
}

type StepWithOutputs interface {
	GetOutputs() []Output
}

type StepMetadata struct {
	sensitiveOutputs []string
}

func (s *StepMetadata) AddSensitiveOutput(o Output) {
	if s.sensitiveOutputs == nil {
		s.sensitiveOutputs = make([]string, 0)
	}

	if !o.IsSensitive() {
		return
	}

	s.sensitiveOutputs = append(s.sensitiveOutputs, o.GetName())
}
