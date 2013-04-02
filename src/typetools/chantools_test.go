package chantools

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

func TestLock(t *testing.T) {
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	ChanDebug(c)
}

func TestNoRecv(t *testing.T) {
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	data, length, _ := ChanRead(c, 10)
	if length != 3 {
		t.Fail()
	}
	if data != 0 {
		t.Fail()
	}
}

func TestRecv(t *testing.T) {
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	data, len, size := ChanRead(c, 0)
	if data == 0 {
		t.Fail()
	}
	if size != 4*3 {
		t.Fail()
	}
	// cast to Slice
	tmp := &reflect.SliceHeader{
		Data: data,
		Cap:  len,
		Len:  len,
	}
	slice := *(*[]int)(unsafe.Pointer(tmp))

	if slice[0] != 1 {
		t.Fail()
	}
	if slice[1] != 2 {
		t.Fail()
	}
	if slice[2] != 3 {
		t.Fail()
	}
}

type TestStruct struct {
	A byte
	B uint32
	C uint64
}

func TestRecvStruct(t *testing.T) {
	var (
		one TestStruct
		two TestStruct
	)
	one.A = 10
	one.B = 20
	one.C = 30
	two.A = 40
	two.B = 50
	two.C = 60
	c := make(chan TestStruct, 10)
	c <- one
	c <- two
	data, len, _ := ChanRead(c, 0)
	if data == 0 {
		t.Fail()
	}
	// cast to Slice
	tmp := &reflect.SliceHeader{
		Data: data,
		Cap:  len,
		Len:  len,
	}
	slice := *(*[]TestStruct)(unsafe.Pointer(tmp))

	oneR := slice[0]
	fmt.Println(oneR)
	if oneR.A != 10 || oneR.B != 20 || oneR.C != 30 {
		t.Fail()
	}
	twoR := slice[1]
	if twoR.A != 40 || twoR.B != 50 || twoR.C != 60 {
		t.Fail()
	}
}
