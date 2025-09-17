import outer

class Animal:
    def __init__(self, name):
        self.name = name

    def speak(self):
        print(f"{self.name} is speaking")

# inherit in same package
class Dog(Animal):

    def __init__(self, name, age):
        super().__init__(name)
        self._age = age

    def run(self):
        print(f"{self.name} is running")
    
    @property
    def age(self):
        return self._age
    
    @age.setter
    def age(self, age):
        self._age = age
    
    @classmethod
    def get_dog_name(cls):
        return "Dog"
    
    @staticmethod
    def static_method():
        print("This is a static method")
    
    def __str__(self):
        return f"Dog(name={self.name}, age={self._age})"

class A:
    def __init__(self):
        self.a = 1

# multiple inheritance
class B(A, Dog):
    def __init__(self, name, age):
        super().__init__(name, age)
        self.b = 2

# inherit class which is multiple inheritance
class C(B):
    def __init__(self, name, age):
        super().__init__(name, age)
        self.c = 3

# same name
class Outer:
    def __init__(self):
        self.outer = 1

# inheritance from other module
class D(outer.Outer):
    def __init__(self):
        super().__init__()
        self.c = 3


class E(D):
    def __init__(self):
        super().__init__()
        self.d = 4

# Multi-level inheritance
class F(E):
    def __init__(self):
        super().__init__()
        self.f = 5
