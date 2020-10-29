package builder

type Output interface {
	GetName() string
}

type StepWithOutputs interface {
	GetOutputs() []Output
}
