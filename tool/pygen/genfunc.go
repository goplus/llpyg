package pygen

import (
	"go/ast"
	"go/token"
	"go/types"
	"github.com/goplus/gogen"
	"github.com/goplus/llpyg/symbol"
)

func (ctx *context) genFunc(pkg *gogen.Package, sym *symbol.Symbol) {
	name, symSig := sym.Name, sym.Sig
	if len(name) == 0 || name[0] == '_' {
		return
	}
	if symSig.Str == "" { // no signature
		ctx.skips = append(ctx.skips, *sym)
		return
	}
	// signature
	params, variadic := ctx.genParams(pkg, symSig.Str)
	goName := ctx.genName(name, -1)
	sig := types.NewSignatureType(nil, nil, nil, params, ctx.ret, variadic) // ret: *py.Object
	fn := pkg.NewFuncDecl(token.NoPos, goName, sig)
	// doc
	docList := ctx.genDoc(sym.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	docList = append(docList, ctx.genLinkname(goName, sym))
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}

func (ctx *context) genLinkname(name string, sym *symbol.Symbol) *ast.Comment {
	return &ast.Comment{Text: "//go:linkname " + name + " py." + sym.Name}
}
