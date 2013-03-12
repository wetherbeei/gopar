// Type tests
package main

import "fmt"
import (
	"go/ast"
	"go/token"
)

func main() {
	{
		a := 1
		b := 0.2
		c := "hello"
		d := true
		e := false
		f := []int64{1, 2, 3}
		i := Struct1{}
		j := &Struct1{}
	}
	{
		var a *int64
		var b [4]int64
		var c *Struct1
	}
	{
		a := make([]int64, 10)
		b := make(map[int64]string)
		c := Func1()
	}
	{
		a, b := 1, 2
		c, d := a, b
		d, e, f := Func2()
		g, h, i := Func3(true, 1)
	}
	{
		a := make(chan int)
		b := <-a
	}
}

type Struct1 struct {
	a       int
	b       string
	c, d, e float32
}

func (s *Struct1) Method1() {
	return
}

func (s Struct1) Method2() {

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

func Func1() bool {
	return false
}

func Func2() (int, string, float32) {
	return 0, "", 0.0
}

func Func3(a bool, b int) (c, d []int, e [4]int) {
	return
}
