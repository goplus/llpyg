llpyg - LLGo bindings autogen tool for Python libraries
====

[![Build Status](https://github.com/goplus/llpyg/actions/workflows/llgo.yml/badge.svg)](https://github.com/goplus/llpyg/actions/workflows/llgo.yml)
[![GitHub release](https://img.shields.io/github/v/tag/goplus/llpyg.svg?label=release)](https://github.com/goplus/llpyg/releases)
[![Coverage Status](https://codecov.io/gh/goplus/llpyg/branch/main/graph/badge.svg)](https://codecov.io/gh/goplus/llpyg)
[![GoDoc](https://pkg.go.dev/badge/github.com/goplus/llpyg.svg)](https://pkg.go.dev/github.com/goplus/llpyg)
[![LLGo](https://img.shields.io/badge/powered_by-LLGo-green.svg)](https://github.com/xgo-dev/llgo)
[![XGo](https://img.shields.io/badge/project-XGo-blue.svg)](https://github.com/goplus/xgo)

llpyg aims to be a tool for automatically generating [LLGo](https://github.com/xgo-dev/llgo) bindings for Python libraries, enhancing the experience of integrating LLGo with Python.

With llpyg, you point the tool at a Python library, and it produces a Go package that LLGo can compile and link against the real Python library. You can then call that library from LLGo code as if it were a native Go package, without writing the binding layer by hand.
