## 安装与使用

### Dependencies

- [LLGo](https://github.com/goplus/llgo)
- [Python 3.12+](https://www.python.org/)

需要设置以下环境变量：
```bash
export LLGO_ROOT=/path/to/llgo
export PYTHONHOME=/path/to/python
```
### How to install
Install from source:
```bash
git clone https://github.com/toaction/llpyg.git
cd llpyg
bash install.sh
```

### Usage
你可以选择两种不同的方式来执行命令，分别是：
- 命令行参数
- llpyg.cfg 配置文件

**1. 命令行参数**

```bash
llpyg [-o output_dir] [-mod mod_name] [-d module_depth] py_lib_name
```

- `-o`: LLGo Bindings output dir, default `./test`.
- `-mod`: Output Go module name, default `py_lib_name`.
- `-d`: Extract Python module max depth, default `1`.

**2. llpyg.cfg 文件**

llpyg.cfg 是配置文件，可以对内容进行修改，llpyg将会根据该文件执行程序。示例：

```json
{
  "name": "numpy",
  "libName": "numpy",
  "modules": [
    "numpy"
  ]
}
```

- `name`: Go package name.
- `libName`: Python library name.
- `modules`: Extract Python modules.

修改好后，执行命令：
```bash
llpyg [-o output_dir] [-mod mod_name] cfg_path
```