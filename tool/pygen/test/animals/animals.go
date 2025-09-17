package animals

import (
	"github.com/goplus/lib/py"
	_ "unsafe"
)

const LLGoPackage = "py.animals"

type Animal struct {
	py.Object
}
type A struct {
	py.Object
}
type Outer struct {
	py.Object
}
type D struct {
	py.Object
}
type C struct {
	py.Object
}
type Dog struct {
	Animal
}
type E struct {
	D
}
type F struct {
	E
}

//go:linkname NewD py.D
func NewD() *D

//go:linkname NewC py.C
func NewC(name *py.Object, age *py.Object) *C

//go:linkname NewDog py.Dog
func NewDog(name *py.Object, age *py.Object) *Dog

//llgo:link (*Dog).Run py.Dog.run
func (*Dog) Run() *py.Object {
	return nil
}

// Return str(self).
//
//llgo:link (*Dog).Str py.Dog.__str__
func (*Dog) Str() *py.Object {
	return nil
}

//llgo:link (*Dog).GetDogName py.Dog.get_dog_name
func (*Dog) GetDogName() *py.Object {
	return nil
}

//llgo:link (*Dog).StaticMethod py.Dog.static_method
func (*Dog) StaticMethod() *py.Object {
	return nil
}

//llgo:link (*Dog).Age py.Dog.age.__get__
func (*Dog) Age() *py.Object {
	return nil
}

//llgo:link (*Dog).SetAge py.Dog.age.__set__
func (*Dog) SetAge(age *py.Object) {
}

//go:linkname NewE py.E
func NewE() *E

//go:linkname NewF py.F
func NewF() *F

//go:linkname NewAnimal py.Animal
func NewAnimal(name *py.Object) *Animal

//llgo:link (*Animal).Speak py.Animal.speak
func (*Animal) Speak() *py.Object {
	return nil
}

//go:linkname NewA py.A
func NewA() *A

//go:linkname NewOuter py.Outer
func NewOuter() *Outer
