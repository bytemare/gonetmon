// sniffer_test launches a monitoring session for a given amount of time
package main

import (
	. "github.com/bytemare/gonetmon"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

// SnifferTest is a test function for Sniffer use with a timeout
func SnifferTest(duration time.Duration) {

	testWait := sync.WaitGroup{}
	testWait.Add(1)
	result := make(chan error)
	go Sniff(&testWait)

	// Send interrupt signal after timeout
	p, _ := os.FindProcess(os.Getpid())

loop:
	for {
		select {
		case err := <-result:
			if err != nil {
				log.Error("Sniffing returned with an error.")
				log.Error(err)
			}
			break loop

		case <-time.After(duration):
			_ = p.Signal(os.Interrupt)
			res := <-result
			if res != nil {
				log.Error("Sniffing returned with an error.")
				log.Error(res)
			}
			break loop
		}
	}

	testWait.Wait()
}

func main() {
	duration := 180 * time.Second
	SnifferTest(duration)
}
