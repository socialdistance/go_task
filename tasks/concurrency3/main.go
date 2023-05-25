package main

import (
	"fmt"
)

func NewRingBuffer(inCh, outCh chan int) *ringBuffer {
	return &ringBuffer{
		inCh:  inCh,
		outCh: outCh,
	}
}

type ringBuffer struct {
	inCh  chan int
	outCh chan int
}

func (r *ringBuffer) Run() {
	size := cap(r.outCh)
	for v := range r.inCh {
		if len(r.outCh) == size {
			<-r.outCh
		}
		r.outCh <- v
	}

	close(r.outCh)
}

func main() {
	inCh := make(chan int)
	outCh := make(chan int, 4)
	doneCh := make(chan struct{})
	rb := NewRingBuffer(inCh, outCh)

	go rb.Run()

	max := 100

	for i := 0; i < max; i++ {
		inCh <- i
	}
	close(inCh)

	resSlice := make([]int, 0)
	go func() {
		for res := range outCh {
			resSlice = append(resSlice, res)
		}
		doneCh <- struct{}{}
	}()

	<-doneCh

	fmt.Println(resSlice)
}
