// Type tests
package main

func main() {
	var x = &Struct1{}
	// s.a.WriteAccess -> x.a.WriteAccess
	x.Method1()
	var y = &Struct2{}
	// s.a.WriteAccess -> y.b.a.WriteAccess
	y.b.Method1()

	var z = &Struct1{}
	ModifyStruct1(z)

	var w = Struct1{}
	DontModifyStruct1(w)
	panic(z)

	var Func = func() {
		// write to a global variable
	}

	Func() // must assume all functions that we cannot see directly inside cannot be parallelized
}

type Struct1 struct {
	a       int
	b       string
	c, d, e float32
}

func ModifyStruct1(s *Struct1) {
	s.a = 2
	return
}

func DontModifyStruct1(s Struct1) {
	s.a = 2
	return
}

func (s *Struct1) Method1() {
	s.a = 1
	return
}

func (s Struct1) Method2() {
	return
}

type Struct2 struct {
	b Struct1
}

type Struct3 struct {
	Struct2
}

type Iface1 interface {
	Method1(arg1 int, arg2 int) (int32, int64)
}

type Iface2 interface {
	Iface1
	Method2(arg2 int)
}

func Func1() bool {
	return false
}

func Func2() (int, string, float32) {
	return 0, "", 0.0
}

func Func3(a bool, b int) (c, d []int, e [4]int) {
	return
}
