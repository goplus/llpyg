# Problem Description

Currently, LLGo only supports calling Python functions and does not support handling classes and objects. The official repository (https://github.com/goplus/lib/py) also lacks LLGo Bindings code examples related to Python classes.

To expand LLGo's integration capabilities with the Python ecosystem, support for Python classes needs to be added. On one hand, llpyg needs to generate corresponding LLGo Bindings code, and on the other hand, LLGo needs to add corresponding processing logic.

## Symbol Information

Python class declarations contain the following symbol information:
- Name and inheritance relationships
- Methods (including class methods, instance methods, static methods)
- Properties (including class properties, instance properties)
- Special methods (methods starting and ending with `__`)

Python class example:
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
    def get_dog_name_static():
        return "Dog"
    
    def __str__(self):
        return f"Dog {self._name} is {self._age} years old"
```

Usage example:
```Python
dog = Dog("Buddy", 3)
dog.speak()
age = dog.age
dog.age = 4
dog.get_dog_name()
dog.get_dog_name_static()
print(dog)
```

## LLGo Bindings Design

A good approach is to convert Python classes to Go structs and convert all symbols declared in the class to struct methods.

### Classes and Instances
**For classes**, convert to structs. Single inheritance is implemented through embedding. For class instances, create through NewClassName functions and link to Python class symbols through `go:linkname`:

> The parameters of the NewClassName function are obtained from the `__init__` method.

```Go
type Animal struct {
	py.Object
}

//go:linkname NewAnimal py.Animal
func NewAnimal(name *py.Object) *Animal
```

### Methods
**For methods**, convert class methods, static methods, instance methods, and special methods to struct methods uniformly, using `llgo:link` directives to link method symbols:

```Go
//llgo:link (*Animal).Speak py.Animal.speak
func (a *Animal) Speak(name *py.Object) *py.Object {
    return nil
}

//llgo:link (*Animal).Str py.Animal.__str__
func (a *Animal) Str() *py.Object {
    return nil
}
```

### Properties
**For properties**, split them into Get and Set methods:
- Methods for getting property values do not need a Get prefix
- Methods for setting property values need a Set prefix and have no return value

```Go
//llgo:link (*Animal).Age py.Animal.age.__get__
func (a *Animal) Age() *py.Object {
    return nil
}

//llgo:link (*Animal).SetAge py.Animal.age.__set__
func (a *Animal) SetAge(age *py.Object) {
}
```

### Usage
Corresponding usage:
```Go
animal := NewAnimal(py.Str("Animal"))
animal.Speak(py.Str("msg"))
str := animal.Str()
age := animal.Age()
animal.SetAge(py.Long(10))
```

## Existing Problems

### Multiple Inheritance
In Python, a class can inherit from multiple parent classes:

```Python
class Child(Parent1, Parent2):
    pass
```

In Go, multiple inheritance is implemented by embedding multiple structs in a struct:

```Go
type Child struct {
	Parent1
	Parent2
}
```

However, embedding multiple structs has **memory layout problems**:

In Python, all objects have the same basic structure, and properties from multiple inheritance are merged. But in Go, structs implement inheritance through embedded fields, which results in continuous memory layout. Due to the different memory layouts, objects returned by CPython cannot be directly converted to structs that embed multiple `py.Object` fields.

### Initialization Methods

For pure Python classes, instances are created and initialized by calling the `__init__` method. Therefore, the `__init__` method can be converted to a NewClassName function to implement Python class creation. The linked symbol at this time is `py.ClassName`.

```Go
//go:linkname NewAnimal py.Animal
func NewAnimal(name *py.Object) *Animal

//go:linkname NewAnimal py.Animal.__init__
func NewAnimal(name *py.Object) *Animal
```

However, when testing the numpy library, it was found that C-implemented extension types (such as `numpy.ndarray`) do not have an `__init__` method, but create instances through the `__new__` method. In this scenario, does the linked symbol need to be changed?

```Go
//go:linkname NewNdarray py.Ndarray
func NewNdarray(shape *py.Object) *Ndarray
```

```Go
//go:linkname NewNdarray py.Ndarray.__new__
func NewNdarray(shape *py.Object) *Ndarray
```