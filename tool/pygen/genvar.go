package pygen

import (
	"github.com/goplus/gogen"
	"go/ast"
	"go/token"
	"github.com/goplus/llpyg/symbol"
)


func (ctx *context) genVars(pkg *gogen.Package, syms []*symbol.Symbol) {
	names := make(map[string]struct{})
	for _, sym := range syms {
		if sym.Name == "" || sym.Name[0] == '_' {
			continue
		}
		name := ctx.genName(sym.Name, -1)
		_, exist := names[name]
		// avoid name conflict
		for exist {
			name = name + "_"
			_, exist = names[name]
		}
		names[name] = struct{}{}
		ctx.genVar(pkg, sym, name)
	}
}

func (ctx *context) genVar(pkg *gogen.Package, sym *symbol.Symbol, goName string) {
	def := pkg.NewVarDefs(pkg.Types.Scope())
	// linkname
	docList := make([]*ast.Comment, 0, 2)
	goLinkname := "//go:linkname " + goName + " py." + sym.Name
	docList = append(docList, &ast.Comment{Text: goLinkname})
	def.SetComments(&ast.CommentGroup{List: docList})
	// new
	def.New(token.NoPos, ctx.objPtr, goName)
}
