package main

import (
	"log"
	"sync"
	"time"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(1)
	go write(&wg)
	wg.Wait()
}

func write(wg *sync.WaitGroup) {
	for i := 0; i <= 10; i++ {
		log.Printf("Writing %d", i)
		time.Sleep(1 * time.Second)
	}
	wg.Done()
}
