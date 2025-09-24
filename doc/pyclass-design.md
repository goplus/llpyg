# 问题描述

目前，LLGo 仅支持调用 Python 函数，不支持处理类和对象。官方仓库 [goplus/lib/py](https://github.com/goplus/lib/py) 也缺少与 Python 类相关的 LLGo Bindings 代码示例。

为了扩展 LLGo 与 Python 生态系统的集成能力，需要添加对 Python 类的支持。一方面，llpyg 需要生成相应的 LLGo Bindings 代码，另一方面，LLGo 需要添加相应的处理逻辑。

## 符号信息

Python 类声明包含以下符号信息：
- 名称和继承关系
- 方法（包括类方法、实例方法、静态方法和特殊方法）
- 属性（包括类属性、实例属性）

Python 类示例：
```Python
class Animal:
    def __init__(self, name):
        self._name = name

    def speak(self):
        pass

class Dog(Animal):

    dog_name = "Dog"

    def __init__(self, name, age):
        super().__init__(name)
        self._age = age

    def speak(self):
        print(f"Dog {self._name} is speaking")
    
    @property
    def age(self):
        return self._age

    @age.setter
    def age(self, age):
        self._age = age
    
    @classmethod
    def get_dog_name(cls):
        return cls.dog_name
    
    @staticmethod
    def get_name():
        return "Dog"
    
    def __str__(self):
        return f"Dog {self._name} is {self._age} years old"
```

使用示例：
```Python
dog = Dog("Buddy", 3)
dog.speak()
age = dog.age
dog.age = 4
Dog.dog_name = "Dog1"
print(Dog.dog_name)   # Dog1
Dog.get_dog_name()
Dog.get_name()
print(dog)
```

## LLGo Bindings 设计

### 类和实例
将 Python 类转换为 Go 结构体，通过嵌入实现类似继承的功能。对于类实例，通过声明的 `New[ClassName]` 构造函数进行创建。

> NewClassName 函数的参数从类的 `__init__` 或 `__new__` 方法获得。

Python 类：
```Python
class Dog(Animal):
    def __init__(self, name, age):
        super().__init__(name)
        self._age = age

dog = Dog("Buddy", 3)
```

Go 结构体：

```Go
type Dog struct {
	Animal
}

//go:linkname NewDog py.Dog
func NewDog(name *py.Object, age *py.Object) *Dog
```

LLGo 使用方式：

```Go
dog := NewDog(py.Str("Buddy"), py.Long(3))
```

### 方法
Python 类中声明的方法包括类方法、实例方法、静态方法和特殊方法 （魔法方法）。

```Python
def speak(self):
    pass

@classmethod
def get_dog_name(cls):
    pass

@staticmethod
def get_name():
    pass

def __str__(self):
    pass
```
```Python
dog = Dog("Buddy", 3)
dog.speak()
Dog.get_dog_name()
Dog.get_name()
str(dog)
```

类方法、实例方法和特殊方法都与类或实例相关联，因此将它们转为 Go 结构体方法。

> 对于特殊方法，去除前后下划线，使其更符合命名规范与用户使用习惯。

```Go
//llgo:link (*Dog).Speak py.Dog.speak
func (d *Dog) Speak() *py.Object {
    return nil
}

//llgo:link (*Dog).GetDogName py.Dog.get_dog_name
func (d *Dog) GetDogName() *py.Object {
    return nil
}

//llgo:link (*Dog).Str py.Dog.__str__
func (d *Dog) Str() *py.Object {
    return nil
}
```

对于静态方法，与类和实例无关，因此将它们转为 Go 函数。为了防止命名冲突以及更符合用户使用习惯，添加类名作为函数名的前缀。

```Go
//go:linkname DogGetName py.Dog.get_name
func DogGetName() *py.Object {
    return nil
}
```

LLGo 使用方式：

```Go
dog := NewDog(py.Str("Buddy"), py.Long(3))
dog.Speak()
dog.GetDogName()
dog.Str()
DogGetName()
```

### 属性
Python 类中声明的属性包括类属性和实例属性。
```Python
class Dog(Animal):
    dog_name = "Dog"

    @property
    def age(self):
        return self._age

    @age.setter
    def age(self, age):
        self._age = age
```
```Python
dog = Dog("Buddy", 3)
dog.age
dog.age = 4
Dog.dog_name
```

在 Python 中可以对属性执行获取和设置操作，因此将每个属性拆分为 `get` 和 `set` 方法。

> 对于属性的获取操作，为了符合用户使用习惯，不添加 Get 前缀。

```Go
//llgo:link (*Dog).Age py.Dog.age.__get__
func (d *Dog) Age() *py.Object {
    return nil
}

//llgo:link (*Dog).SetAge py.Dog.age.__set__
func (d *Dog) SetAge(age *py.Object) {
}
```

LLGo 使用方式：

```Go
dog := NewDog(py.Str("Buddy"), py.Long(3))
dog.Age()
dog.SetAge(py.Long(4))
```

## 现有问题

### 多重继承
在 Python 中，一个类可以继承多个父类：

```Python
class Child(Parent1, Parent2):
    pass
```

在 Go 中，类似多重继承的功能通过在结构体中嵌入多个结构体来实现：

```Go
type Child struct {
	Parent1
	Parent2
}
```

然而，嵌入多个结构体存在**内存布局问题**：

在 Python 中，所有对象都有相同的基本结构，多重继承的属性会被合并。但在 Go 中，结构体通过嵌入字段实现类似继承的功能，这导致连续的内存布局。由于内存布局不同，CPython 返回的对象无法直接转换为嵌入多个 `py.Object` 字段的结构体。

### 类属性

在 Python 中，类属性属于类，而不是实例。通过实例修改类属性，实际上是创建了一个同名的实例属性，并不会影响类和其他实例：

```Python
class Dog:
    dog_name = "dog"

    def __init__(self, name):
        self.name = name

dog1 = Dog("Buddy")
dog2 = Dog("Max")

dog1.dog_name = "dog1"
print(dog1.dog_name)   # dog1
print(dog2.dog_name)   # dog
print(Dog.dog_name)    # dog
```

但通过类名修改类属性，会影响到不存在同名实例属性的所有实例，包括未创建的实例：
```Python
Dog.dog_name = "Dog"
print(Dog.dog_name)     # Dog
print(dog2.dog_name)    # Dog   
```

Python 支持通过类名修改类属性，那么在 Go 这边如何实现？

