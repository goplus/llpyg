package main

import (
	"os"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/py/inspect"
)


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


var pyFuncTypes = map[string]bool{
	"ufunc": true,
	"method": true,
	"function": true,
	"method-wrapper": true,
	"builtin_function_or_method": true,
	"_ArrayFunctionDispatcher": true,
}

func extractSignatureFromDoc(doc, funcName string) string {
    lines := strings.SplitN(doc, "\n\n", 2)
    if len(lines) == 0 {
        return ""
    }
    firstLine := strings.TrimSpace(lines[0])
    if !strings.HasPrefix(firstLine, funcName+"(") {
        return ""
    }
	idx := strings.Index(firstLine, "(")
    if idx == -1 {
        return ""
    }
    params := firstLine[idx:]
    fields := strings.Fields(params)
    return strings.Join(fields, " ")
}


func getSignature(val *py.Object, sym *symbol) string {
	if val.Callable() == 0 {
		return ""
	}
	// inspect
	sigFromInspect := inspect.Signature(val)
	if sigFromInspect != nil {
		sig := c.GoString(sigFromInspect.Str().CStr())
		if sig != "(*args, **kwargs)" {
			return sig
		}
	}
	// doc
	sigFromDoc := extractSignatureFromDoc(sym.Doc, sym.Name)
	if sigFromDoc != "" {
		return sigFromDoc
	}
	// Paradigms
	if pyFuncTypes[sym.Type] {
		return "(*args, **kwargs)"
	}
	return ""
}

// parse class methods
func parseMethod(val *py.Object, name string, typeName string) *symbol {
	sym := &symbol{Name: name, Type: typeName}
	sym.Sig = getSignature(val, sym)
	docObj := inspect.Getdoc(val)
	if docObj != nil {
		sym.Doc = c.GoString(docObj.Str().CStr())
	}
	return sym
}


func parseProperty(val *py.Object, name string) *property {
	property := &property{Name: name}
	fget := val.GetAttrString(c.Str("fget"))
	if fget != nil {
		sig := inspect.Signature(fget)
		if sig != nil {
			property.Getter = c.GoString(sig.Str().CStr())
		}
	}
	fset := val.GetAttrString(c.Str("fset"))
	if fset != nil {
		sig := inspect.Signature(fset)
		if sig != nil {
			property.Setter = c.GoString(sig.Str().CStr())
		}
	}
	return property
}


// parse class symbol
func parseClass(clsObj *py.Object, moduleName string) *class {
	cls := &class{}
	cls.Name = c.GoString(clsObj.GetAttrString(c.Str("__name__")).CStr())
	clsModule := c.GoString(clsObj.GetAttrString(c.Str("__module__")).CStr())
	if clsModule != moduleName {  // TODO: parse other module's class
		return nil
	}
	bases := clsObj.GetAttrString(c.Str("__bases__"))   // tuple
	for i, n := 0, bases.TupleLen(); i < n; i++ {
		baseObj := bases.TupleItem(i)
		base := &base{
			Name: c.GoString(baseObj.GetAttrString(c.Str("__name__")).CStr()),
			Module: c.GoString(baseObj.GetAttrString(c.Str("__module__")).CStr()),
		}
		cls.Bases = append(cls.Bases, base)
	}
	dict := clsObj.GetAttrString(c.Str("__dict__"))
	dictTypeName := c.GoString(dict.Type().TypeName().CStr())
	if dictTypeName != "mappingproxy" {
		return nil
	}
	builtins := py.ImportModule(c.AllocaCStr("builtins"))
	dictFunc := builtins.GetAttrString(c.Str("dict"))
	realDict := dictFunc.CallOneArg(dict)
	items := realDict.DictItems()
	for i, n := 0, items.ListLen(); i < n; i++ {
		item := items.ListItem(i)
		name := c.GoString(item.TupleItem(0).CStr())
		val := item.TupleItem(1)
		typeName := c.GoString(val.Type().TypeName().CStr())
		typeName = strings.TrimSpace(typeName)
		if name == "__init__" {
			sym := parseMethod(val, name, typeName)
			cls.InitMethod = sym
			continue
		}
		switch typeName {
		case "function":
			sym := parseMethod(val, name, typeName)
			cls.InstanceMethods = append(cls.InstanceMethods, sym)
		case "classmethod":
			val = val.GetAttrString(c.Str("__func__"))
			sym := parseMethod(val, name, typeName)
			cls.ClassMethods = append(cls.ClassMethods, sym)
		case "staticmethod":
			val = val.GetAttrString(c.Str("__func__"))
			sym := parseMethod(val, name, typeName)
			cls.StaticMethods = append(cls.StaticMethods, sym)
		case "property":
			property := parseProperty(val, name)
			cls.Properties = append(cls.Properties, property)
		default:
			// TODO: attributes
		}
	}
	return cls
}


// extract symbols from module
func pydump(moduleName string) (*module, error) {
	// import module
	mod := py.ImportModule(c.AllocaCStr(moduleName))
	if mod == nil {
		return nil, fmt.Errorf("failed to import module %s", moduleName)
	}
	// get dict keys
	keys := mod.ModuleGetDict().DictKeys()
	if keys == nil {
		return nil, fmt.Errorf("failed to get dict keys of %s", moduleName)
	}
	// create module instance
	modInstance := &module{
		Name: moduleName,
	}
	// get symbols
	for i, n := 0, keys.ListLen(); i < n; i++ {
		key := keys.ListItem(i)
		val := mod.GetAttr(key)
		if val == nil {
			continue
		}
		// name, type, doc
		sym := &symbol{}
		sym.Name = c.GoString(key.CStr())
		sym.Type = c.GoString(val.Type().TypeName().CStr())
		doc := val.GetAttrString(c.Str("__doc__"))
		if doc != nil {
			sym.Doc = c.GoString(doc.Str().CStr())
		}
		// functions
		if pyFuncTypes[sym.Type] {
			sym.Sig = getSignature(val, sym)
			modInstance.Functions = append(modInstance.Functions, sym)
			continue
		}
		// classes
		if sym.Type == "type" {
			cls := parseClass(val, moduleName)
			if cls != nil {
				cls.Doc = sym.Doc
				modInstance.Classes = append(modInstance.Classes, cls)
			}
			continue
		}
		// TODO: global variables
	}
	return modInstance, nil
}


func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: pydump <py_module_name>")
		return
	}
	moduleName := os.Args[1]
	mod, err := pydump(moduleName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)		
		os.Exit(1)
	}
	// print module information
	data, err := json.MarshalIndent(mod, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to marshal json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(data))
}

