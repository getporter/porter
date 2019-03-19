package kubernetes

type Step struct {
	Description string             `yaml:"description"`
	Outputs     []KubernetesOutput `yaml:"outputs,omitempty"`
}

type KubernetesOutput struct {
	Name         string `yaml:"name"`
	ResourceType string `yaml:"resourceType"`
	ResourceName string `yaml:"resourceName"`
	Namespace    string `yaml:"namespace,omitempty"`
	JSONPath     string `yaml:"jsonPath"`
}
