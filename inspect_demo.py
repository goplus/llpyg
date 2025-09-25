#!/usr/bin/env python3
"""
inspect模块获取函数签名结构化信息的Demo
"""

import inspect
from typing import List, Dict, Optional, Union


def example_function(
    a: int,
    b: str = "default",
    c: Optional[List[str]] = None,
    *args: int,
    d: bool = True,
    **kwargs: Union[str, int]
) -> Dict[str, any]:
    """示例函数，用于演示签名解析"""
    return {"a": a, "b": b, "c": c, "args": args, "d": d, "kwargs": kwargs}


class ExampleClass:
    def __init__(self, name: str, age: int = 0):
        self.name = name
        self.age = age

    def method(self, x: float, y: float = 1.0) -> float:
        return x * y

    @staticmethod
    def static_method(data: List[int]) -> int:
        return sum(data)


def analyze_signature(func, name: str = None):
    """分析函数签名的详细信息"""
    print(f"\n{'='*50}")
    print(f"分析函数: {name or func.__name__}")
    print(f"{'='*50}")

    # 获取签名对象
    sig = inspect.signature(func)
    print(f"完整签名: {sig}")

    # 返回类型注解
    if sig.return_annotation != inspect.Signature.empty:
        print(f"返回类型: {sig.return_annotation}")
    else:
        print("返回类型: 未指定")

    print(f"\n参数详情:")
    print(f"{'参数名':<15} {'类型':<20} {'默认值':<15} {'注解':<25} {'种类'}")
    print("-" * 90)

    for param_name, param in sig.parameters.items():
        # 参数种类的中文映射
        kind_map = {
            param.POSITIONAL_ONLY: "仅位置",
            param.POSITIONAL_OR_KEYWORD: "位置或关键字",
            param.VAR_POSITIONAL: "*args",
            param.KEYWORD_ONLY: "仅关键字",
            param.VAR_KEYWORD: "**kwargs"
        }

        kind_str = kind_map.get(param.kind, str(param.kind))
        default_str = str(param.default) if param.default != param.empty else "无默认值"
        annotation_str = str(param.annotation) if param.annotation != param.empty else "无注解"

        print(f"{param_name:<15} {kind_str:<20} {default_str:<15} {annotation_str:<25}")


def main():
    print("Python inspect模块签名解析Demo")

    # 分析普通函数
    analyze_signature(example_function)

    # 分析类构造函数
    analyze_signature(ExampleClass.__init__, "ExampleClass.__init__")

    # 分析实例方法
    analyze_signature(ExampleClass.method, "ExampleClass.method")

    # 分析静态方法
    analyze_signature(ExampleClass.static_method, "ExampleClass.static_method")

    # 分析内置函数
    try:
        analyze_signature(len)
    except ValueError as e:
        print(f"\n内置函数len无法获取签名: {e}")

    # 演示绑定参数
    print(f"\n{'='*50}")
    print("参数绑定演示")
    print(f"{'='*50}")

    sig = inspect.signature(example_function)
    try:
        # 正确的参数绑定
        bound = sig.bind(1, "hello", ["a", "b"], 2, 3, d=False, extra="value")
        bound.apply_defaults()
        print("绑定参数:", dict(bound.arguments))

        # 调用函数验证
        result = example_function(*bound.args, **bound.kwargs)
        print("函数执行结果:", result)

    except TypeError as e:
        print(f"参数绑定错误: {e}")


if __name__ == "__main__":
    main()