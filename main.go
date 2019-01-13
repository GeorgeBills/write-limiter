package main

import (
	"log"
	"math/rand"
	"time"
)

func main() {
	toWrite := make(chan int)
	timer := time.Tick(1 * time.Second)
	go generate(toWrite)
	go write(toWrite, timer)
	// block forever
	select {}
}

func generate(out chan int) {
	source := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(source)
	for i := 0; true; i++ {
		time.Sleep(time.Duration(rnd.Intn(3000)) * time.Millisecond)
		out <- i
	}
	close(out)
}

func write(in chan int, ticker <-chan time.Time) {
	var next int
	var ready bool
	for {
		select {
		case i := <-in:
			next, ready = i, true
		case <-ticker:
			if ready {
				log.Printf("Writing %d", next)
			}
			ready = false
		}
	}
}
