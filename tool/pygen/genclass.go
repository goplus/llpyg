package pygen

import (
	"github.com/goplus/gogen"
	"go/types"
	"go/token"
	"go/ast"
	"strings"
)



func (ctx *context) genClasses(pkg *gogen.Package, classes []*class, moduleName string) {
	toHandle := map[string]*class{}
	for _, cls := range classes {
		if cls.InitMethod == nil {
			// TODO: support class without __init__
			ctx.skips = append(ctx.skips, symbol{Name: cls.Name, Type: "class"})
			continue
		}
		if len(cls.Bases) > 1 {
			// TODO: support multiple inheritance
			ctx.skips = append(ctx.skips, symbol{Name: cls.Name, Type: "class"})
			continue
		}
		if cls.Bases[0].Name == "object" && cls.Bases[0].Module == "builtins" {
			ctx.genStruct(pkg, cls, false)
			continue
		}
		if cls.Bases[0].Module != moduleName {
			// TODO: support inheritance from other module
			ctx.genStruct(pkg, cls, false)
			continue
		}
		toHandle[cls.Name] = cls
	}
	// structs embed struct in the same package
	for len(toHandle) > 0 {
		for name, cls := range toHandle {
			baseName := cls.Bases[0].Name
			if _, ok := ctx.structs[baseName]; ok {
				ctx.genStruct(pkg, cls, true)
				delete(toHandle, name)
			}
		}
	}

	// generate methods for classes
	classMap := make(map[string]*class)
	for _, cls := range classes {
		classMap[cls.Name] = cls
	}
	for name, structType := range ctx.structs {
		cls := classMap[name]
		ctx.genConstructor(pkg, cls, structType)
		ctx.genMethods(pkg, cls, structType)
		ctx.genProperties(pkg, cls, structType)
	}
}



func (ctx *context) genStruct(pkg *gogen.Package, cls *class, hasParent bool) {
	var baseType types.Type
	if hasParent {
		baseType = ctx.structs[cls.Bases[0].Name]
	} else {
		baseType = ctx.obj
	}
	structType := pkg.NewTypeDefs().NewType(ctx.genName(cls.Name, -1))
	structType.InitType(pkg, types.NewStruct(
		[]*types.Var{
			types.NewVar(0, pkg.Types, "", baseType),
		},
		nil,
	))
	ctx.structs[cls.Name] = structType.Type()
}



func (ctx *context) genConstructor(pkg *gogen.Package, cls *class, structType types.Type) {
	funcName := "New" + ctx.genName(cls.Name, -1)
	sym := cls.InitMethod
	// signature
	params, variadic := ctx.genParams(pkg, sym.Sig, true, false)
	ret := types.NewTuple(pkg.NewParam(0, "", types.NewPointer(structType)))
	sig := types.NewSignatureType(nil, nil, nil, params, ret, variadic)
	fn := pkg.NewFuncDecl(token.NoPos, funcName, sig)
	// linkname
	docList := make([]*ast.Comment, 0, 2)
	goLinkname := "//go:linkname " + funcName + " py." + cls.Name
	docList = append(docList, &ast.Comment{Text: goLinkname})
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}


func (ctx *context) genMethods(pkg *gogen.Package, cls *class, structType types.Type) {
	recv := types.NewVar(0, pkg.Types, "", types.NewPointer(structType))
	for _, method := range cls.InstanceMethods {
		ctx.genMethod(pkg, cls.Name, method.Name, method, recv, true, false, true)
	}
	for _, method := range cls.ClassMethods {
		ctx.genMethod(pkg, cls.Name, method.Name, method, recv, false, true, true)
	}
	for _, method := range cls.StaticMethods {
		ctx.genMethod(pkg, cls.Name, method.Name, method, recv, false, false, true)
	}
}


func (ctx *context) genMethod(pkg *gogen.Package, clsName string, methodName string, method *symbol, recv *types.Var, self bool, cls bool, ret bool) {
	name := method.Name
	if strings.HasPrefix(name, "__") && strings.HasSuffix(name, "__") {
		name = name[2:len(name)-2]
	}
	funcName := ctx.genName(name, -1)
	// signature
	params, variadic := ctx.genParams(pkg, method.Sig, self, cls)
	var retType *types.Tuple
	if ret {
		retType = ctx.ret
	}
	sig := types.NewSignatureType(recv, nil, nil, params, retType, variadic)
	fn,err := pkg.NewFuncWith(token.NoPos, funcName, sig, nil)
	if err != nil {
		return
	}
	if ret {
		fn.BodyStart(pkg).ZeroLit(ctx.objPtr).Return(1).End()
	}else {
		fn.BodyStart(pkg).End()
	}
	// doc
	docList := ctx.genDoc(method.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	//llgo:link
	link := "//llgo:link " + funcName + " py." + clsName + "." + methodName
	docList = append(docList, &ast.Comment{Text: link})
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}


func (ctx *context) genProperties(pkg *gogen.Package, cls *class, structType types.Type) {
	recv := types.NewVar(0, pkg.Types, "", types.NewPointer(structType))
	for _, property := range cls.Properties {
		name := property.Name
		if property.Getter != "" {
			sym := &symbol{Name: name, Sig: property.Getter}
			methodName := name + ".__get__"
			ctx.genMethod(pkg, cls.Name, methodName, sym, recv, true, false, true)
		}
		if property.Setter != "" {
			sym := &symbol{Name:"Set_" + name, Sig: property.Setter}
			methodName := name + ".__set__"
			ctx.genMethod(pkg, cls.Name, methodName, sym, recv, true, false, false)
		}
	}
}