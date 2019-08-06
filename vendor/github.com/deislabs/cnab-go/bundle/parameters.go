package bundle

// Parameter defines a single parameter for a CNAB bundle
type Parameter struct {
	Definition  string    `json:"definition" mapstructure:"definition"`
	ApplyTo     []string  `json:"applyTo,omitempty" mapstructure:"applyTo,omitempty"`
	Description string    `json:"description,omitempty" mapstructure:"description"`
	Destination *Location `json:"destination,omitemtpty" mapstructure:"destination"`
	Required    bool      `json:"required,omitempty" mapstructure:"required,omitempty"`
}
