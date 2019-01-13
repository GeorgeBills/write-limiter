package main

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"
)

const (
	// MaxWritesToBuffer is the maximum number of writes we buffer before we start blocking.
	MaxWritesToBuffer = 100
	// MaxWriteWaitMilliseconds is the maximum time we'll wait before we force a write.
	MaxWriteWaitMilliseconds = 1000
	writeRateMilliseconds    = 100
	// CheckpointFile is the file we write to.
	CheckpointFile = "checkpoint.json"
)

// ToWrite is some struct you want to write.
type ToWrite struct {
	N int
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
			N: i,
		}
		out <- *generated
		generated.N = -1
	}
}

// Write will persist out the toWrite value passed to it. If it receives a
// single toWrite it will write that value immediately (select to read the
// value, then immediately select to the default case and write the value). If
// it's currently being flooded with toWrite values then it will busy loop on
// the in channel until it has stored the last (i.e. most recently sent) value,
// and then trigger a write through the default case. We force a write at least
// every MaxWriteWaitMilliseconds, just to avoid a constant stream of toWrite
// values resulting in us never writing. We intentionally take ToWrite values,
// not pointers, just to make sure our client code isn't changing values around
// on us while we wait to write.
func Write(in chan ToWrite) {
	var toWrite *ToWrite
	ticker := time.Tick(MaxWriteWaitMilliseconds * time.Millisecond)
	maybeWrite := func() {
		if toWrite != nil {
			err := writeToFile(toWrite)
			if err != nil {
				log.Fatalf("Error writing to checkpoint file")
			}
			toWrite = nil
		}
	}
	for {
		select {
		case <-ticker:
			// force a write at least every MaxWriteWaitMilliseconds
			maybeWrite()
		case next := <-in:
			// read from channel until we've read the last (i.e. most recent) toWrite
			toWrite = &next
		default:
			// if there's a toWrite to write out then write it
			maybeWrite()
			// rate limit this case to max once per writeRateMilliseconds
			// otherwise we're just busy looping. we're waiting at least
			// this time before we check for another value to write.
			time.Sleep(writeRateMilliseconds * time.Millisecond)
		}
	}
}

func writeToFile(toWrite *ToWrite) error {
	log.Printf("Writing %d", toWrite.N)

	file, err := os.OpenFile(CheckpointFile, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	encoder := json.NewEncoder(writer)
	encoder.Encode(toWrite)
	writer.Flush()
	file.Sync()

	return nil
}
