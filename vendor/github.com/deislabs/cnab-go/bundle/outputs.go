package bundle

type OutputsDefinition struct {
	Fields   map[string]OutputDefinition `json:"fields" mapstructure:"fields"`
	Required []string                    `json:"required,omitempty" mapstructure:"required,omitempty"`
}

type OutputDefinition struct {
	DataType      string             `json:"type" mapstructure:"type"`
	Default       interface{}        `json:"default,omitempty" mapstructure:"default"`
	AllowedValues []interface{}      `json:"allowedValues,omitempty" mapstructure:"allowedValues"`
	MinValue      *int               `json:"minValue,omitempty" mapstructure:"minValue"`
	MaxValue      *int               `json:"maxValue,omitempty" mapstructure:"maxValue"`
	MinLength     *int               `json:"minLength,omitempty" mapstructure:"minLength"`
	MaxLength     *int               `json:"maxLength,omitempty" mapstructure:"maxLength"`
	Metadata      *ParameterMetadata `json:"metadata,omitempty" mapstructure:"metadata"`
	Path          string             `json:"path,omitemtpty" mapstructure:"path,omitempty"`
	ApplyTo       []string           `json:"apply-to,omitempty" mapstructure:"apply-to,omitempty"`
	Sensitive     bool               `json:"sensitive,omitempty" mapstructure:"sensitive,omitempty"`
}
