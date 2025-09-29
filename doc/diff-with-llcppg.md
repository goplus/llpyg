## llpyg 与 llcppg 的设计差异

> llcppg: https://github.com/goplus/llcppg

### 1. 目标语言

- llcppg 面向 C/C++ 库，处理静态类型语言
- llpyg 面向 Python 库，处理动态类型语言

### 2. 符号信息获取方式

- llcppg 静态分析：C/C++ 库有明确的头文件定义，使用 libclang 解析头文件，获取 AST，得到类型与函数声明

- llpyg 动态分析：Python 库没有统一的符号定义文件，目前借助 CPython API，利用反射获取符号信息

> llpyg 可以通过解析 AST 的方式获取符号信息，但工作量较大，且具有一定难度

### 3. 类型处理

llcppg 类型处理：将头文件声明与库符号匹配，处理复杂类型映射。

```go
// llgo:link (*BZFILE).Read C.BZ2_bzread
func (recv_ *BZFILE) Read(buf c.Pointer, len c.Int) c.Int {
	return 0
}
```

llpyg 类型处理：主要使用 `py.Object` 作为通用参数类型，返回类型为 `*py.Object`。
> 完善对类的支持后将会增加返回类型转换的功能。

```go
//go:linkname A py.A
var A *py.Object

//llgo:link (*Dog).Speak py.Dog.speak
func (d *Dog) Speak(msg py.Object) *py.Object {
	return nil
}
```

### 4. 配置文件

llcppg 配置文件 (llcppg.cfg)：
```json
{
	"name": "zlib",
	"cflags": "$(pkg-config --cflags zlib)",
	"libs": "$(pkg-config --libs zlib)",
	"include": [
		"zlib.h",
		"zconf.h"
	],
	"deps": ["c/os"],
}
```
llpyg 配置文件 (llpyg.cfg)：
```json
{
  "name": "numpy",
  "libName": "numpy", 
  "modules": ["numpy", "numpy.random"]
}
```
- `llcppg.cfg` 中的 `cflags` 和 `libs` 字段用于提供编译和链接 C/C++ 库所需的信息，而 `llpyg.cfg` 中的 `libName` 字段则指定了要处理的 Python 库名。两者都用于定位目标库。
- `include` 字段和 `modules` 字段作用等同，指出要处理的文件/模块有哪些。
- `deps` 字段用于声明库的依赖，在类型处理时需要用到。Python 也有类似的需求，但目前类型统一为了 `py.Object`，后续会添加依赖声明相关功能。


### 5. 架构设计

llcppg 与 llpyg 在生成 LLGo Bindings 的处理逻辑是接近的，大致包含三个阶段：符号信息提取 -> 函数签名解析/类型处理 -> Go 代码生成。

主要差异体现在处理方式上：
- llcppg 通过解析头文件的 AST 提取符号信息，llpyg 通过导入模块利用反射获取符号信息
- llcppg 利用符号表对各种类型都做了映射，llpyg 统一为了 `py.Object`


### 总结

由于目标语言的根本差异，它们在设计和实现上有显著不同。llcppg 更注重静态分析和复杂的类型系统处理，而 llpyg 更依赖运行时反射和简化的对象模型。


