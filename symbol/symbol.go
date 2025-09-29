package symbol

type Symbol struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type Module struct {
	Name      string    `json:"name"`      // python module name
	Functions []*Symbol `json:"functions"` // package functions
	Variables []*Symbol `json:"variables"` // package variables
	// TODO: classes, etc.
}
