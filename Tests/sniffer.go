// gonetmon is a network monitoring tool.
// It captures packets on the wire from devices based on given criteria, and displays statistics about traffic.
package main

import (
	. "github.com/bytemare/gonetmon"
	log "github.com/sirupsen/logrus"
	"os"
	"time"
)
import "flag"

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
