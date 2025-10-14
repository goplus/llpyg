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
	Classes   []*Class  `json:"classes"`
}

// base class
type Base struct {
	Name   string `json:"name"`
	Module string `json:"module"`
}

// @property
type Property struct {
	Name   string `json:"name"`
	Getter string `json:"getter"`
	Setter string `json:"setter"`
}

// Python class
type Class struct {
	Name            string      `json:"name"`
	Doc             string      `json:"doc"`
	Bases           []*Base     `json:"bases"`
	InitMethod      *Symbol     `json:"initMethod"`
	InstanceMethods []*Symbol   `json:"instanceMethods"` // include override special methods
	ClassMethods    []*Symbol   `json:"classMethods"`
	StaticMethods   []*Symbol   `json:"staticMethods"`
	Properties      []*Property `json:"properties"`
	// TODO: attributes
}
