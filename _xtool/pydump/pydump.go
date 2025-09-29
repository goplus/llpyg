package main

import (
	"os"
	"fmt"
	"strings"
	"encoding/json"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/py/inspect"
	"github.com/goplus/llpyg/symbol"
)

var pyFuncTypes = map[string]bool{
	"ufunc":                      true,
	"method":                     true,
	"function":                   true,
	"method-wrapper":             true,
	"builtin_function_or_method": true,
	"_ArrayFunctionDispatcher":   true,
}

var pyVarTypes = map[string]struct{}{
	"int":      {},
	"float":    {},
	"complex":  {},
	"bool":     {},
	"str":      {},
	"list":     {},
	"tuple":    {},
	"set":      {},
	"dict":     {},
	"NoneType": {},
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

func getSignature(val *py.Object, sym *symbol.Symbol) string {
	// function, method, class, or implement __call__
	if val.Callable() == 0 {
		return ""
	}
	// get signature from inspect
	sigFromInspect := inspect.Signature(val)
	if sigFromInspect != nil {
		sig := c.GoString(sigFromInspect.Str().CStr())
		if sig != "(*args, **kwargs)" {
			return sig
		}
	}
	// get signature from doc
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

// moduleName: Python module name
func pydump(moduleName string) (*symbol.Module, error) {
	// import module
	mod := py.ImportModule(c.AllocaCStr(moduleName))
	if mod == nil {
		return nil, fmt.Errorf("failed to import module %s", moduleName)
	}
	// get dict, python list Object
	keys := mod.ModuleGetDict().DictKeys()
	if keys == nil {
		return nil, fmt.Errorf("failed to get dict keys of %s", moduleName)
	}
	// create module instance
	modInstance := &symbol.Module{
		Name: moduleName,
	}
	// get symbols
	for i, n := 0, keys.ListLen(); i < n; i++ {
		key := keys.ListItem(i)
		val := mod.GetAttr(key)
		if val == nil {
			continue
		}
		// define symbol
		sym := &symbol.Symbol{}
		sym.Name = c.GoString(key.CStr())
		sym.Type = c.GoString(val.Type().TypeName().CStr())
		doc := val.GetAttrString(c.Str("__doc__"))
		if doc != nil && doc.IsTrue() == 1 {
			sym.Doc = c.GoString(doc.Str().CStr())
		}
		// functions
		if pyFuncTypes[sym.Type] {
			sym.Sig = getSignature(val, sym)
			modInstance.Functions = append(modInstance.Functions, sym)
			continue
		}
		// variables
		if _, ok := pyVarTypes[sym.Type]; ok {
			modInstance.Variables = append(modInstance.Variables, sym)
			continue
		}
		// TODO: classes, etc.
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
