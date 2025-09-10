## 基本介绍

**[LLGo](https://github.com/goplus/llgo)**：一个基于 LLVM 的 Go 编译器，以便更好地与 C 和 Python 生态集成。

**LLGo Bindings**：为了能够在 Go 代码中调用其他语言，需要将其他语言的符号映射为 Go 符号，形成接口，LLGo 通过该接口来实现对其他语言的调用。我们将这种一一映射的接口称为 LLGo Bindings，即 LLGo 对其他语言的绑定代码。

以 Python 的 `numpy.add` 函数为例：
```Python
numpy.add(x1, x2, /, out=None, *, where=True, casting='same_kind', order='K', dtype=None, subok=True[, signature, extobj])
```

对应的 LLGo Bindings 为：
```Go
//go:linkname Add py.add
func Add(x1 *py.Object, x2 *py.Object) *py.Object
```

**[llpyg](https://github.com/goplus/llpyg)**：一个面向 Python 库的 LLGo Bindings 自动生成工具。

## 设计决策

### llpyg 是否需要脱离 LLGo ?
> https://github.com/goplus/llpyg/issues/5

llpyg 面向的是那些需要 LLGo Bindings 的用户，即 LLGo 开发者，因此可以依赖于 LLGo。

llpyg 依赖于 LLGo 的 Python 生态集成能力，该工具的一些子组件如 `pydump` 和 `pymodule` 需要 LLGo 进行编译和安装。

### llpyg 是否需要为用户提供与系统无关的 Python 环境？

> https://github.com/goplus/llpyg/issues/2#issuecomment-3200109475

llpyg 默认使用系统 Python 或用户指定的 Python，并不为用户提供 Python 安装及第三方库自动下载的服务。

### 要转换的 Python 库的版本是否可以指定？

> https://github.com/goplus/llpyg/issues/8

llpyg 使用的是用户已经安装好的 Python 库的版本。用户若想转换不同版本的 Python 库，需要手动更改已安装的库。

### llpyg 是否需要为用户提供指定 Python 路径的功能？

> https://github.com/goplus/llpyg/issues/9

用户可以通过 `PYTHONHOME` 环境变量来指定 Python 路径。 llpyg 并不会提供一个单独的环境变量。

## 架构设计

### 输入输出

- 输入: Python 库名称
- 输出: LLGo Bindings 代码

**命令输入执行**
```bash
llpyg [-o ouput_dir] [-mod mod_name] [-d module_depth] py_lib_name
```
参数说明：
- `-o`: 输出目录，默认值为 `./test`
- `-mod`: 生成 Go 模块名称，默认值为空
- `-d`: 获取 Python 库的模块的深度，默认值为 1
- `py_lib_name`: Python 库名称

**配置文件输入执行**
```bash
llpyg [-o output_dir] [-mod mod_name] llpyg.cfg
```
参数说明：
- `-o`: 输出目录，默认值为 `./test`
- `-mod`: 生成 Go 模块名称，默认值为空
- `llpyg.cfg`: 配置文件路径

配置文件示例：
```json
{
  "name": "numpy",
  "libName": "numpy",
  "modules": [
    "numpy"
  ]
}
```

**程序输出**
```go
numpy
├── numpy.go    // 主模块 LLGo Bindings 文件
├── random
│   └── random.go    // 子模块 LLGo Bindings 文件
├── go.mod
├── go.sum
└── llpyg.cfg    // 配置文件
```

### 项目结构
```go
llpyg
├── _xtool
│   ├── pydump
│   └── pymodule
├── cmd
│   └── llpyg
├── doc
├── tool
│   ├── pyenv
│   ├── pygen
│   └── pysig
├── go.mod
├── go.sum
├── install.sh
├── README.md
└── LICENSE
```

- `_xtool`: 需要使用 LLGo 进行编译和安装的子组件
- `cmd`: llpyg 可执行文件
- `tool`: 子模块，不依赖 LLGo
- `doc`: 存放项目文档


### 模块划分

```text
_xtool
├── pydump
└── pymodule
```
该目录下的 LLGo 程序安装后会放在 `$GOPATH/bin` 目录下。
- `pymodule`: 获取 Python 库的主模块及多级子模块名称
- `pydump`: 获取 Python 库的符号信息，包括函数、类等信息

```text
cmd
└── llpyg
```
- `llpyg`: llpyg 入口程序

```text
tool
├── pyenv
├── pygen
└── pysig
```

- `pyenv`: 设置 Python 路径，对环境进行检查
- `pysig`: 解析 Python 函数和方法签名
- `pygen`: 调用 `pydump` 和 `pysig`，生成 Go 代码

### 执行流程

llpyg 执行流程：

1. 执行 `pyenv`，设置 Python 路径，对环境进行检查
2. 执行 `pymodule`，获取模块名，生成 llpyg.cfg 配置文件
3. 根据 `llpyg.cfg` 配置文件，逐模块执行 `pygen`，生成 Go 代码

Go 代码生成流程：

1. 执行 `pydump`，获取 Python 库的符号信息
2. 执行 `pysig`，解析 Python 函数和方法签名
3. 调用 `gogen`，生成 Go 代码


### 模块接口
**`pymodule` 模块接口**：
```bash
pymodule [-d <depth>] <libraryName>
```
参数说明：
- `-d <depth>`: 获取模块的深度，默认值为 1
- `<libraryName>`: Python 库名称

返回值（以 `numpy` 库为例）：
```json
{
  "libName": "numpy",
  "libVersion": "1.26.0",
  "depth": 2,
  "modules": [
    "numpy",
    "numpy.array_api",
    "numpy.random",
  ]
}
```

**`pydump` 模块接口：**
```bash
pydump <moduleName>
```
参数说明：
- `<moduleName>`: Python 模块名称

返回值：
```json
{
  "name": "numpy",
  "items": [
    {
      "name": "__name__",
      "type": "str",
      "doc": "...",
      "sig": ""
    },
  ]
}
```


