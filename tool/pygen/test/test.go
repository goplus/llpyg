package main

import (
	"os"
	"github.com/goplus/llpyg/tool/pyenv"
	"github.com/goplus/llpyg/tool/pygen"
)


func main() {
	os.Setenv("PYTHONHOME", "/Users/mac/work/xgo/python")
	pyenv.Prepare()
	pygen.GenLLGoBindings("sklearn.metrics", os.Stdout)
}
