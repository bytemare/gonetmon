// gonetmon is a network monitoring tool.
// It captures packets on the wire from devices based on given criteria, and displays statistics about traffic.
package main

import (
	"os"
	"sync"
	"time"
)
import "flag"

// SnifferTest is a wrapper function for Sniffer use with a timeout
func SnifferTest(duration time.Duration) error {

	testWait := sync.WaitGroup{}
	testWait.Add(1)
	result := make(chan error)
	go Sniff(&testWait, result)

	// Send interrupt signal after timeout
	p, _ := os.FindProcess(os.Getpid())

	var err error
loop:
	for {
		select {
		case err = <-result:
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
			err = res
			break loop
		}
	}

	testWait.Wait()
	return err
}

func main() {
	var err error
	timeout := flag.Int("timeout", 0, "monitoring time in seconds. 0 or none is infinite")
	flag.Parse()

	if *timeout > 0 {
		log.Info("Started with timeout : ", *timeout)
		err = SnifferTest(time.Duration(*timeout) * time.Second)

	} else {
		log.Info("Started without timeout : ", *timeout)
		err = Sniff(nil, nil)
	}

	if err == nil {
		os.Exit(0)
	}
	os.Exit(1)
}
