package sdk

// RegisterRequest defines the structure for the registration request.
// The JSON tags MUST match what the host application expects.
type RegisterRequest struct {
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	UIComponentPath  string           `json:"uiComponentPath"`
	CustomElementTag string           `json:"customElementTag"`
	Contributions    *UIContributions `json:"contributions,omitempty"`
}
