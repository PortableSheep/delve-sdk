package sdk

// RegisterRequest defines the structure for the registration request.
// The JSON tags MUST match what the host application expects.
type RegisterRequest struct {
	Name             string `json:"name"`
	Description      string `json:"description"`
	UiComponentPath  string `json:"uiComponentPath"`
	CustomElementTag string `json:"customElementTag"`
}
