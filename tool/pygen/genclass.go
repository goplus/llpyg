package pygen

import (
	"github.com/goplus/gogen"
	"github.com/goplus/llpyg/symbol"
	"go/ast"
	"go/token"
	"go/types"
	"log"
	"strings"
)

func (ctx *context) genClasses(pkg *gogen.Package, classes []*symbol.Class) {
	for _, cls := range classes {
		if len(cls.Bases) == 0 {
			log.Fatalf("Class %s has no base", cls.Name)
		}
		// now inherit from object only
		// TODO: support other bases
		ctx.genStruct(pkg, cls.Name)
		if cls.InitMethod != nil {
			ctx.genConstructor(pkg, cls)
		}
		ctx.genMethods(pkg, cls)
		ctx.genProperties(pkg, cls)
	}
}

func (ctx *context) genStruct(pkg *gogen.Package, name string) {
	structType := pkg.NewTypeDefs().NewType(ctx.genName(name, -1))
	// TODO: support parents embed
	structType.InitType(pkg, types.NewStruct(
		[]*types.Var{
			types.NewVar(0, pkg.Types, "", ctx.obj),
		},
		nil,
	))
	ctx.structs[name] = structType.Type()
}

func (ctx *context) genConstructor(pkg *gogen.Package, cls *symbol.Class) {
	// TODO: convert to XGO struct.New
	funcName := "New" + ctx.genName(cls.Name, -1)
	params, variadic := ctx.genParams(pkg, cls.InitMethod.Sig)
	// return struct pointer
	ret := types.NewTuple(pkg.NewParam(0, "", types.NewPointer(ctx.structs[cls.Name])))
	sig := types.NewSignatureType(nil, nil, nil, params, ret, variadic)
	fn := pkg.NewFuncDecl(token.NoPos, funcName, sig)
	// doc
	docList := ctx.genDoc(cls.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	// linkname
	goLinkname := "//go:linkname " + funcName + " py." + cls.Name
	docList = append(docList, &ast.Comment{Text: goLinkname})
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}

func (ctx *context) genMethods(pkg *gogen.Package, cls *symbol.Class) {
	for _, method := range cls.InstanceMethods {
		name := method.Name
		// TODO: XGO magic method style
		if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
			name = name[2 : len(name)-2]
		}
		funcName := ctx.genName(name, -1)
		// full name
		method.Name = cls.Name + "." + method.Name
		ctx.genMethod(pkg, method, cls.Name, funcName, true) // return *py.Object
	}
	// TODO: class methods, static methods to XGO static methods
}

func (ctx *context) genProperties(pkg *gogen.Package, cls *symbol.Class) {
	for _, property := range cls.Properties {
		name := cls.Name + property.Name
		if property.Getter != "" {
			sym := &symbol.Symbol{Name: name + ".__get__", Sig: property.Getter}
			funcName := ctx.genName(property.Name, -1)
			ctx.genMethod(pkg, sym, cls.Name, funcName, true)
		}
		if property.Setter != "" {
			sym := &symbol.Symbol{Name: name + ".__set__", Sig: property.Setter}
			funcName := ctx.genName("set_"+property.Name, -1)
			ctx.genMethod(pkg, sym, cls.Name, funcName, false)
		}
	}
}
func (ctx *context) genMethod(pkg *gogen.Package, method *symbol.Symbol, clsName, funcName string, ret bool) {
	cls := ctx.structs[clsName]
	recv := types.NewVar(0, pkg.Types, "", types.NewPointer(cls))
	// signature
	params, variadic := ctx.genParams(pkg, method.Sig)
	var retType *types.Tuple
	if ret {
		retType = ctx.ret
	}
	sig := types.NewSignatureType(recv, nil, nil, params, retType, variadic)
	fn, err := pkg.NewFuncWith(token.NoPos, funcName, sig, nil)
	if err != nil {
		log.Fatal(err)
	}
	if ret {
		fn.BodyStart(pkg).ZeroLit(ctx.objPtr).Return(1).End()
	} else {
		fn.BodyStart(pkg).End()
	}
	// doc
	docList := ctx.genDoc(method.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	//llgo:link
	funcName = "(*" + cls.Obj().Name() + ")." + funcName
	link := "//llgo:link " + funcName + " py." + method.Name
	docList = append(docList, &ast.Comment{Text: link})
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}
