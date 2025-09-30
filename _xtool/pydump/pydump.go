package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

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

func getSignature(val *py.Object, sym *symbol.Symbol) (*symbol.Signature, error) {
	// which implement __call__
	if val.Callable() == 0 {
		return nil, fmt.Errorf("not callable")
	}
	// inspect
	sigFromInspect := inspect.Signature(val)
	if sigFromInspect != nil {
		sig := c.GoString(sigFromInspect.Str().CStr())
		return &symbol.Signature{
			Source: symbol.SigSourceInspect,
			Str:    sig,
		}, nil
	}
	// doc
	sigFromDoc := extractSignatureFromDoc(sym.Doc, sym.Name)
	if sigFromDoc != "" {
		return &symbol.Signature{
			Source: symbol.SigSourceDoc,
			Str:    sigFromDoc,
		}, nil
	}
	// Paradigms
	if pyFuncTypes[sym.Type] {
		return &symbol.Signature{
			Source: symbol.SigSourceParadigm,
			Str:    "(*args, **kwargs)",
		}, nil
	}
	return nil, fmt.Errorf("failed to get signature")
}

// moduleName: Python module name
func pydump(moduleName string) (*symbol.Module, error) {
	mod := py.ImportModule(c.AllocaCStr(moduleName))
	if mod == nil {
		return nil, fmt.Errorf("failed to import module %s", moduleName)
	}
	keys := mod.ModuleGetDict().DictKeys()
	if keys == nil {
		return nil, fmt.Errorf("failed to get dict keys of %s", moduleName)
	}
	modInstance := &symbol.Module{
		Name: moduleName,
	}
	for i, n := 0, keys.ListLen(); i < n; i++ {
		key := keys.ListItem(i)
		val := mod.GetAttr(key)
		if val == nil {
			continue
		}
		sym := &symbol.Symbol{}
		sym.Name = c.GoString(key.CStr())
		sym.Type = c.GoString(val.Type().TypeName().CStr())
		doc := val.GetAttrString(c.Str("__doc__"))
		if doc != nil && doc.IsTrue() == 1 {
			sym.Doc = c.GoString(doc.Str().CStr())
		}
		// functions
		if pyFuncTypes[sym.Type] {
			sig, err := getSignature(val, sym)
			if err != nil {
				return nil, err
			}
			sym.Sig = *sig
			modInstance.Functions = append(modInstance.Functions, sym)
		}
		// TODO: variables, classes, etc.
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
