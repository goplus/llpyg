package demo

import (
	"github.com/goplus/lib/py"
	_ "unsafe"
)

const LLGoPackage = "py.demo"

type Animal struct {
	py.Object
}

//go:linkname NewAnimal py.Animal
func NewAnimal(name *py.Object) *Animal

//llgo:link (*Animal).Speak py.Animal.speak
func (*Animal) Speak() *py.Object {
	return nil
}

type Dog struct {
	py.Object
}

//go:linkname NewDog py.Dog
func NewDog(name *py.Object, age *py.Object) *Dog

//llgo:link (*Dog).Speak py.Dog.speak
func (*Dog) Speak() *py.Object {
	return nil
}

//llgo:link (*Dog).Str py.Dog.__str__
func (*Dog) Str() *py.Object {
	return nil
}

//llgo:link (*Dog).Age py.Dogage.__get__
func (*Dog) Age() *py.Object {
	return nil
}

//llgo:link (*Dog).SetAge py.Dogage.__set__
func (*Dog) SetAge(age *py.Object) {
}
