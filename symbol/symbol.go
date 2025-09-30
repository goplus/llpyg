package symbol

// represents the origin of a function signature
type SigSource int

const (
	// signature extracted from docstring
	SigSourceDoc SigSource = iota
	// signature from Python inspect module
	SigSourceInspect
	// generic fallback signature
	SigSourceParadigm
)

// represents a function signature with its source information
type Signature struct {
	Str    string    `json:"str"`    // The signature string
	Source SigSource `json:"source"` // Where the signature was obtained
}

type Symbol struct {
	Name string    `json:"name"`
	Type string    `json:"type"`
	Doc  string    `json:"doc"`
	Sig  Signature `json:"sig"`
}

type Module struct {
	Name      string    `json:"name"`      // python module name
	Functions []*Symbol `json:"functions"` // package functions
	// TODO: variables, classes, etc.
}
