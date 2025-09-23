package pygen

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"encoding/json"
	"strings"
	"log"
	"strconv"
	"go/token"
	"go/ast"
	"go/types"
	"github.com/goplus/gogen"
	"github.com/goplus/llpyg/tool/pysig"
)


type symbol struct {
	Name string `json:"name"` // python name
	Type string `json:"type"`
	Doc  string `json:"doc"`
	Sig  string `json:"sig"`
}

type module struct {
	Name  string    `json:"name"` // python module name
	Items []*symbol `json:"items"`
}

type context struct {
	pkg    *gogen.Package
	obj    *types.Named
	objPtr *types.Pointer
	ret    *types.Tuple
	skips  []element
	todos  []element
	py     gogen.PkgRef
}

type element struct {
	name   string
	pyType string
}


func GenLLGoBindings(moduleName string, outFile io.Writer) {
	// get module symbols info from pydump
	mod, err := pydump(moduleName)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	// create go package
	ctx := createGoPackage(mod)

	// generate go code
	ctx.genMod(ctx.pkg, &mod)

	skips := ctx.skips
	if n := len(skips); n > 0 {
		log.Printf("==> Skip %d symbols:\n%v\n", n, skips)
	}

	// write to file
	ctx.pkg.WriteTo(outFile)
}

func pydump(moduleName string) (mod module, err error) {
	var out bytes.Buffer
	cmd := exec.Command("pydump", moduleName)
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return mod, fmt.Errorf("pydump %s failed: %w", moduleName, err)
	}
	err = json.Unmarshal(out.Bytes(), &mod)
	if err != nil {
		return mod, fmt.Errorf("unmarshal %s failed: %w", moduleName, err)
	}
	if mod.Name != moduleName {
		return mod, fmt.Errorf("import module failed: %s", moduleName)
	}
	return mod, nil
}

func createGoPackage(mod module) (ctx *context) {
	parts := strings.Split(mod.Name, ".")
	pkgName := parts[len(parts)-1]
	if goKeywords[pkgName] {
		pkgName = pkgName + "_"
	}
	pkg := gogen.NewPackage("", pkgName, nil)
	pkg.Import("unsafe").MarkForceUsed(pkg)      // import _ "unsafe"
	py := pkg.Import("github.com/goplus/lib/py") // import "github.com/goplus/lib/py"
	f := func(cb *gogen.CodeBuilder) int {
		cb.Val("py." + mod.Name)
		return 1
	}
	defs := pkg.NewConstDefs(pkg.Types.Scope())
	defs.New(f, 0, 0, nil, "LLGoPackage") // const LLGoPackage = "py.moduleName"

	obj := py.Ref("Object").(*types.TypeName).Type().(*types.Named)
	objPtr := types.NewPointer(obj)
	ret := types.NewTuple(pkg.NewParam(0, "", objPtr)) // return *py.Object
	ctx = &context{pkg, obj, objPtr, ret, nil, nil, py}
	return ctx
}

var pyFuncTypes = map[string]bool{
	"ufunc": true,
	"method": true,
	"function": true,
	"method-wrapper": true,
	"builtin_function_or_method": true,
	"_ArrayFunctionDispatcher": true,
}

func (ctx *context) genMod(pkg *gogen.Package, mod *module) {
	funcMap := make(map[string]bool)
	for _, sym := range mod.Items {
		if funcMap[sym.Name] {
			continue
		}
		if pyFuncTypes[sym.Type] {
			funcMap[sym.Name] = true
			ctx.genFunc(pkg, sym)
		}
		// Todo: class, variable, etc.
		ctx.todos = append(ctx.todos, element{name: sym.Name, pyType: sym.Type})
	}
}

func (ctx *context) genFunc(pkg *gogen.Package, sym *symbol) {
	name, symSig := sym.Name, sym.Sig
	if len(name) == 0 || name[0] == '_' {
		return
	}
	if symSig == "" { // no signature
		ctx.skips = append(ctx.skips, element{name: name, pyType: sym.Type})
		return
	}
	// signature
	params, variadic := ctx.genParams(pkg, symSig)
	name = genName(name, -1)
	sig := types.NewSignatureType(nil, nil, nil, params, ctx.ret, variadic) // ret: *py.Object
	fn := pkg.NewFuncDecl(token.NoPos, name, sig)
	// doc
	docList := ctx.genDoc(sym.Doc)
	if len(docList) > 0 {
		docList = append(docList, emptyCommentLine)
	}
	docList = append(docList, ctx.genLinkname(name, sym))
	fn.SetComments(pkg, &ast.CommentGroup{List: docList})
}

var goKeywords = map[string]bool{
    "package": true, "import": true, "var": true, "const": true, "func": true, "type": true,
    "if": true, "else": true, "switch": true, "case": true, "default": true, "for": true, "range": true,
    "break": true, "continue": true, "goto": true, "fallthrough": true,
    "return": true, "defer": true,
    "go": true, "chan": true, "select": true,
    "struct": true, "interface": true, "map": true,
}

func (ctx *context) genParams(pkg *gogen.Package, sig string) (*types.Tuple, bool) {
	args := pysig.Parse(sig)
	if len(args) == 0 {
		return nil, false
	}
	n := len(args)
	objPtr := ctx.objPtr
	list := make([]*types.Var, 0, n)
    listNum := 0
	for i := 0; i < n; i++ {
		name := strings.TrimSpace(args[i].Name)
        // go keyword
        if goKeywords[name] {
            name = name + "_"
        }
		if name[0] == '(' {				// (a1, a2, ...) -> list_0
			name = "list_" + strconv.Itoa(listNum)
			listNum++
		}
		if name == "/" || name == "" || name == "," {
			continue
		}
		if name == "*" || name == "\\*" {
			break
		}
		if strings.HasPrefix(name, "*") {
			if name[1] != '*' {
				list = append(list, vArgs)
				return types.NewTuple(list...), true		// *args
			}
			return types.NewTuple(list...), false		// **kwargs
		}
		list = append(list, pkg.NewParam(0, genName(name, 0), objPtr))
	}
	return types.NewTuple(list...), false
}

// python name to go name
func genName(name string, idxDontTitle int) string {
	lastIdx := len(name) - 1
	for lastIdx >= 0 && name[lastIdx] == '_' {
		lastIdx--
	}
	workingName := name[:lastIdx+1]
	trail := name[lastIdx+1:]
	parts := strings.Split(workingName, "_")
	// underline to camel case
	for i, part := range parts {
		if i != idxDontTitle && part != "" {
			if c := part[0]; c >= 'a' && c <= 'z' {
				part = string(c+'A'-'a') + part[1:]
			}
			parts[i] = part
		}
	}
	name = strings.Join(parts, "") + trail
	if goKeywords[name] {
		name = name + "_"
	}
	return name
}

func (ctx *context) genLinkname(name string, sym *symbol) *ast.Comment {
	return &ast.Comment{Text: "//go:linkname " + name + " py." + sym.Name}
}

// Generate documentation comments from the symbol's doc string
func (ctx *context) genDoc(doc string) []*ast.Comment {
	if doc == "" || doc == "None" {
		return make([]*ast.Comment, 0, 4)
	}
	lines := strings.Split(doc, "\n")
	list := make([]*ast.Comment, len(lines), len(lines)+4)
	for i, line := range lines {
		list[i] = &ast.Comment{Text: "// " + line}
	}
	return list
}

const (
	NameValist = "__llgo_va_list"
)

var (
	emptyCommentLine = &ast.Comment{Text: "//"}
	tyAny            = types.NewInterfaceType(nil, nil)
	vArgs            = types.NewParam(0, nil, NameValist, types.NewSlice(tyAny))
)
