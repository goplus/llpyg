package pygen

type symbol struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type module struct {
	Name      string    `json:"name"`      // python module name
	Functions []*symbol `json:"functions"` // package functions
	Variables []*symbol `json:"variables"` // package variables
	// TODO: classes, etc.
}
