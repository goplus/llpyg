class Animal:
    def __init__(self, name):
        self._name = name

    def speak(self):
        pass

class Dog(Animal):

    DOG_NAME = "Dog"

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
    def bark(cls, msg):
        print(f"Dog is barking {msg}")

    @staticmethod
    def sleep():
        print("Dog is sleeping")

    def __str__(self):
        return f"Dog {self._name} is {self._age} years old"
