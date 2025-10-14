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

func getSignature(val *py.Object, sym *symbol.Symbol, skipFirst bool) (string, error) {
	// which has __call__
	if val.Callable() == 0 {
		return "", fmt.Errorf("%s is not callable", sym.Name)
	}
	// use inspect
	sigObj := inspect.Signature(val)
	if sigObj != nil {
		sig := c.GoString(sigObj.Str().CStr())
		if skipFirst {
			sig = removeFirstParam(sig)
		}
		return sig, nil
	}
	// parse doc
	sig := extractSignatureFromDoc(sym.Doc, sym.Name)
	if sig != "" {
		return sig, nil
	}
	// Paradigms
	if pyFuncTypes[sym.Type] {
		return "(*args, **kwargs)", nil
	}
	return "", fmt.Errorf("failed to get signature of %s", sym.Name)
}

// (self, ...) -> (...)
func removeFirstParam(sig string) string {
	idx := strings.Index(sig, ",")
	if idx == -1 {
		return "()"
	}
	return "(" + strings.TrimSpace(sig[idx+1:])
}

// moduleName: Python module name
func pydump(moduleName string) (*symbol.Module, error) {
	// import module
	mod := py.ImportModule(c.AllocaCStr(moduleName))
	if mod == nil {
		return nil, fmt.Errorf("failed to import module %s", moduleName)
	}
	// get object dict
	keys := mod.ModuleGetDict().DictKeys()
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
			sig, err := getSignature(val, sym, false)
			if err != nil {
				return nil, err
			}
			sym.Sig = sig
			modInstance.Functions = append(modInstance.Functions, sym)
			continue
		}
		// classes
		if inspect.Isclass(val).IsTrue() == 1 {
			cls, err := parseClass(val, sym)
			if err != nil {
				return nil, err
			}
			modInstance.Classes = append(modInstance.Classes, cls)
			continue
		}
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
