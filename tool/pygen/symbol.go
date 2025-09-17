package pygen



type symbol struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type base struct {
	Name 	string `json:"name"`
	Module 	string `json:"module"`
}

type property struct {
	Name 	string `json:"name"`
	Getter 	string `json:"getter"`			// getAttr sig
	Setter 	string `json:"setter"`			// setAttr sig
}

type class struct {
	Name 			string 		`json:"name"`
	Doc 			string 		`json:"doc"`
	Bases 			[]*base 	`json:"base"`
	Properties 		[]*property `json:"properties"`
	InitMethod 		*symbol 	`json:"initMethod"`
	InstanceMethods []*symbol 	`json:"instanceMethods"`		// include override special methods (__name__)
	ClassMethods 	[]*symbol 	`json:"classMethods"`
	StaticMethods 	[]*symbol 	`json:"staticMethods"`
	// TODO: attributes
}

type module struct {
	Name  		string    	`json:"name"`
	Functions 	[]*symbol 	`json:"functions"`
	Classes 	[]*class 	`json:"classes"`
	//TODO: global variables
}

