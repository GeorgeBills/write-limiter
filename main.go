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
	// MaxWriteWait is the maximum time we'll wait before we force a write.
	MaxWriteWait = 1000 * time.Millisecond
	// WriteRate is the time we wait in between writes; we won't write more than this rate.
	WriteRate = 100 * time.Millisecond
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

// Write will persist out the latest toWrite value passed to it through the in
// channel.
//
// If it receives a single toWrite it will write that value immediately (select
// to read the value, then immediately select to the default case and write the
// value).
//
// If it's currently being flooded with toWrite values then it will busy loop on
// the in channel until it has stored the last (i.e. most recently sent) value,
// and then trigger a write through the default case.
//
// We force a write at least every MaxWriteWait to avoid a constant stream of
// toWrite values resulting in us never writing.
//
// We take ToWrite values, not pointers, to make sure our client code isn't
// changing values around on us while we wait to write.
func Write(in chan ToWrite) {
	var toWrite *ToWrite
	ticker := time.Tick(MaxWriteWait)
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
			// force a write attempt at least every MaxWriteWait
			maybeWrite()
		case next := <-in:
			// read from channel until we've read the last (i.e. most recent) toWrite
			toWrite = &next
		default:
			// if there's a toWrite to write out then write it
			maybeWrite()
			// rate limit this case to max once per WriteRate otherwise we're
			// just busy looping. we're waiting at least this time before we
			// check for another value to write.
			time.Sleep(WriteRate)
		}
	}
}

func writeToFile(toWrite *ToWrite) error {
	log.Printf("Writing %v", toWrite)

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
