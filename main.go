package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	toWrite := make(chan int)
	timer := time.Tick(1 * time.Second)
	go generate(toWrite)
	go write(&wg, toWrite, timer)
	wg.Wait()
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

func write(wg *sync.WaitGroup, in chan int, ticker <-chan time.Time) {
	var next int
LOOP:
	for {
		select {
		case i, ok := <-in:
			if !ok {
				break LOOP
			}
			next = i
		case <-ticker:
			if next > 0 {
				log.Printf("Writing %d", next)
			} else {
				log.Printf("Nothing to write")
			}
			next = -1
		}
	}
	wg.Done()
}
