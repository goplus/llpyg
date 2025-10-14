package main

import (
	"fmt"
	"strings"
	"github.com/goplus/lib/c"
	"github.com/goplus/lib/py"
	"github.com/goplus/lib/py/inspect"
	"github.com/goplus/llpyg/symbol"
)

func parseClass(class *py.Object, sym *symbol.Symbol) (*symbol.Class, error) {
	cls := &symbol.Class{
		Name: sym.Name,
		Doc:  sym.Doc,
	}
	// bases
	bases, err := parseBases(class, sym.Name)
	if err != nil {
		return nil, err
	}
	cls.Bases = bases
	// methods, properties, etc.
	cls, err = parseClassDict(class, cls)
	if err != nil {
		return nil, err
	}
	return cls, nil
}

// get class parents
func parseBases(class *py.Object, name string) ([]*symbol.Base, error) {
	basesObj := class.GetAttrString(c.Str("__bases__")) // tuple
	if basesObj == nil {
		return nil, fmt.Errorf("can't get __bases__ from %s", name)
	}
	bases := make([]*symbol.Base, 0)
	for i, n := 0, basesObj.TupleLen(); i < n; i++ {
		baseObj := basesObj.TupleItem(i)
		base := &symbol.Base{
			Name:   c.GoString(baseObj.GetAttrString(c.Str("__name__")).CStr()),
			Module: c.GoString(baseObj.GetAttrString(c.Str("__module__")).CStr()),
		}
		bases = append(bases, base)
	}
	return bases, nil
}

func parseClassDict(class *py.Object, cls *symbol.Class) (*symbol.Class, error) {
	items, err := getRealDictItems(class, cls.Name)
	if err != nil {
		return nil, err
	}
	for i, n := 0, items.ListLen(); i < n; i++ {
		item := items.ListItem(i)
		name := c.GoString(item.TupleItem(0).CStr())
		val := item.TupleItem(1)
		typeName := c.GoString(val.Type().TypeName().CStr())
		typeName = strings.TrimSpace(typeName)
		// init method
		if name == "__init__" {
			sym := &symbol.Symbol{
				Name: name,
				Type: typeName,
			}
			sig, err := getSignature(val, sym, true)
			if err != nil {
				return nil, err
			}
			sym.Sig = sig
			cls.InitMethod = sym
			continue
		}
		// instance method
		if inspect.Isfunction(val).IsTrue() == 1 {
			sym, err := parseMethod(val, name, typeName, true)
			if err != nil {
				return nil, err
			}
			cls.InstanceMethods = append(cls.InstanceMethods, sym)
			continue
		}
		// hard-code
		switch typeName {
		case "classmethod":
			val = val.GetAttrString(c.Str("__func__"))
			if val == nil {
				return nil, fmt.Errorf("can't get __func__ of %s", name)
			}
			sym, err := parseMethod(val, name, typeName, true)
			if err != nil {
				return nil, err
			}
			cls.ClassMethods = append(cls.ClassMethods, sym)
		case "staticmethod":
			val = val.GetAttrString(c.Str("__func__"))
			if val == nil {
				return nil, fmt.Errorf("can't get __func__ of %s", name)
			}
			sym, err := parseMethod(val, name, typeName, false)
			if err != nil {
				return nil, err
			}
			cls.StaticMethods = append(cls.StaticMethods, sym)
		case "property":
			property, err := parseProperty(val, name)
			if err != nil {
				return nil, err
			}
			cls.Properties = append(cls.Properties, property)
		default:
			// TODO: others
		}
	}
	return cls, nil
}

func parseMethod(val *py.Object, name string, typeName string, skipFirst bool) (*symbol.Symbol, error) {
	sym := &symbol.Symbol{
		Name: name,
		Type: typeName,
	}
	doc := val.GetAttrString(c.Str("__doc__"))
	if doc != nil && doc.IsTrue() == 1 {
		sym.Doc = c.GoString(doc.Str().CStr())
	}
	sig, err := getSignature(val, sym, skipFirst)
	if err != nil {
		return nil, err
	}
	sym.Sig = sig
	return sym, nil
}

func parseProperty(val *py.Object, name string) (*symbol.Property, error) {
	property := &symbol.Property{
		Name: name,
	}
	getter := val.GetAttrString(c.Str("fget"))
	if getter != nil {
		// (self) -> value
		property.Getter = "()"
	}
	setter := val.GetAttrString(c.Str("fset"))
	if setter != nil {
		sym := &symbol.Symbol{
			Name: name,
			Type: "property",
		}
		// (self, value) -> None
		sig, err := getSignature(setter, sym, true)
		if err != nil {
			return nil, err
		}
		property.Setter = sig
	}
	return property, nil
}

func getRealDictItems(class *py.Object, name string) (*py.Object, error) {
	dict := class.GetAttrString(c.Str("__dict__"))
	if dict == nil {
		return nil, fmt.Errorf("can't get __dict__ of %s", name)
	}
	dictTypeName := c.GoString(dict.Type().TypeName().CStr())
	if dictTypeName != "mappingproxy" {
		return nil, fmt.Errorf("__dict__ of %s is not a mappingproxy", name)
	}
	realDict := dict.CallMethod(c.Str("copy"), nil)
	if realDict == nil {
		return nil, fmt.Errorf("failed to copy real dict of %s", name)
	}
	return realDict.DictItems(), nil
}
