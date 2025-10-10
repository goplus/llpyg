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
	"go/ast"
	"go/types"
	"github.com/goplus/gogen"
	"github.com/goplus/llpyg/tool/pysig"
	"github.com/goplus/llpyg/symbol"
)


type context struct {
	pkg    *gogen.Package
	obj    *types.Named
	objPtr *types.Pointer
	ret    *types.Tuple
	py     gogen.PkgRef
	skips  []symbol.Symbol
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

func pydump(moduleName string) (mod symbol.Module, err error) {
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

func createGoPackage(mod symbol.Module) (ctx *context) {
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
	ctx = &context{pkg, obj, objPtr, ret, py, nil}
	return ctx
}

func (ctx *context) genMod(pkg *gogen.Package, mod *symbol.Module) {
	// global functions
	funcMap := make(map[string]bool)
	for _, sym := range mod.Functions {
		if funcMap[sym.Name] {
			continue
		}
		funcMap[sym.Name] = true
		ctx.genFunc(pkg, sym)
	}
	// variables
	ctx.genVars(pkg, mod.Variables)
	// TODO: class, etc.
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
		if name == "/" || name == "" || name == "," {
			continue
		}
        // go keyword
        if goKeywords[name] {
            name = name + "_"
        }
		if name[0] == '(' {
			// (a1, a2, ...) -> list_0
			name = "list_" + strconv.Itoa(listNum)
			listNum++
		}
		if name == "*" || name == "\\*" {
			// TODO: support kwargs
			break
		}
		if strings.HasPrefix(name, "*") {
			if name[1] != '*' {
				list = append(list, vArgs)  // *args
				return types.NewTuple(list...), true
			}
			// TODO: support **kwargs
			return types.NewTuple(list...), false
		}
		list = append(list, pkg.NewParam(0, ctx.genName(name, 0), objPtr))
	}
	return types.NewTuple(list...), false
}

// python name to go name
func (ctx *context) genName(name string, idxDontTitle int) string {
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

// Generate documentation comments from the symbol's doc string
func (ctx *context) genDoc(doc string) []*ast.Comment {
	if doc == "" {
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
