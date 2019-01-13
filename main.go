package main

import (
	"log"
	"math/rand"
	"time"
)

const (
	// MaxWritesToBuffer is the maximum number of writes we buffer before we start blocking.
	MaxWritesToBuffer = 100
	// MinWriteWaitMilliseconds is the minimum time we'll wait before we force a write.
	MinWriteWaitMilliseconds = 1000
	writeRateMilliseconds    = 50
)

// ToWrite is some struct you want to write.
type ToWrite struct {
	i int
}

func main() {
	toWrite := make(chan ToWrite, MaxWritesToBuffer)
	go generate(toWrite)
	go Write(toWrite)
	// block forever while the goroutines spin
	select {}
}

func generate(out chan ToWrite) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; true; i++ {
		time.Sleep(time.Duration(rnd.Intn(30)) * time.Millisecond)
		generated := &ToWrite{
			i: i,
		}
		out <- *generated
		generated.i = -1
	}
}

// Write will persist out the toWrite value passed to it. If it receives a
// single toWrite it will write that value immediately (select to read the
// value, then immediately select to the default case and write the value). If
// it's currently being flooded with toWrite values then it will busy loop on
// the in channel until it has stored the last (i.e. most recently sent) value,
// and then trigger a write through the default case. We force a write at least
// every writeRateMilliseconds, just to avoid a constant stream of toWrite
// values resulting in us never writing. We intentionally take ToWrite values,
// not pointers, just to make sure our client code isn't changing values around
// on us while we wait to write.
func Write(in chan ToWrite) {
	var toWrite *ToWrite
	ticker := time.Tick(MinWriteWaitMilliseconds * time.Millisecond)
	for {
		select {
		case <-ticker:
			// force a write at least every MinWriteWaitMilliseconds
			if toWrite != nil {
				doWrite(toWrite)
				toWrite = nil
			}
		case next := <-in:
			// read from channel until we've read the last (i.e. most recent) toWrite
			toWrite = &next
		default:
			// if there's a value to write out then write that value
			if toWrite != nil {
				doWrite(toWrite)
				toWrite = nil
			} else {
				// rate limit this case to max once per writeRateMilliseconds
				// otherwise we're just busy looping. we're waiting at least
				// this time before we check for another value to write.
				time.Sleep(writeRateMilliseconds * time.Millisecond)
			}
		}
	}
}

func doWrite(toWrite *ToWrite) {
	log.Printf("Writing %d", toWrite.i)
	time.Sleep(500 * time.Millisecond)
}
