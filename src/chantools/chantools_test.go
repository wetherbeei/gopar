package chantools

import "testing"

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
	data, length := ChanRead(c, 10)
	if length != 3 {
		t.Fail()
	}
	if data != nil {
		t.Fail()
	}
}

func TestRecv(t *testing.T) {
	c := make(chan int, 10)
	c <- 1
	c <- 2
	c <- 3
	data, _ := ChanRead(c, 0)
	if data == nil {
		t.Fail()
	}
}
