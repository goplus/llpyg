package demo

import (
	"github.com/goplus/lib/py"
	_ "unsafe"
)

const LLGoPackage = "py.demo"

//go:linkname IntVar py.int_var
var IntVar *py.Object

//go:linkname IntVar_ py.Int_var
var IntVar_ *py.Object

//go:linkname FloatVar py.float_var
var FloatVar *py.Object

//go:linkname ComplexVar py.complex_var
var ComplexVar *py.Object

//go:linkname BoolVar py.bool_var
var BoolVar *py.Object

//go:linkname StrVar py.str_var
var StrVar *py.Object

//go:linkname ListVar py.list_var
var ListVar *py.Object

//go:linkname TupleVar py.tuple_var
var TupleVar *py.Object

//go:linkname SetVar py.set_var
var SetVar *py.Object

//go:linkname DictVar py.dict_var
var DictVar *py.Object

//go:linkname NoneVar py.none_var
var NoneVar *py.Object
