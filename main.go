package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	toWrite := make(chan int)
	go generate(toWrite)
	go write(&wg, toWrite)
	wg.Wait()
}

func generate(out chan int) {
	for i := 0; i <= 10; i++ {
		out <- i
		time.Sleep(1 * time.Second)
	}
}

func write(wg *sync.WaitGroup, in chan int) {
	for i := range in {
		log.Printf("Writing %d", i)
	}
	wg.Done()
}
