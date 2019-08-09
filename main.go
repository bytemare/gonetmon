package main

import (
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

func main() {

	//TODO : do proper prompt to ask for privileges
	if os.Geteuid() != 0 {
		log.Fatal("You must run this program with elevated privileges in order to capture traffic. Try running with sudo.")
	}

	// Load parameters
	params := LoadParams()

	// Check whether we can run
	devices, err := InitialiseCapture(params.Interfaces)
	if err != nil {
		log.Fatal("Initialising capture failed. ", err)
	}

	log.Info("Found devices ", len(devices))

	// IPCs
	var nbReceivers = 1
	var wg sync.WaitGroup
	dataChan := make(chan dataMsg, 1000)
	reportChan := make(chan reportMsg, 1)
	alertChan := make(chan alertMsg, 1)
	syncChan := make(chan struct{})

	// Run Sniffer/Collector
	nbReceivers++
	wg.Add(1)
	go Collector(params, devices, dataChan, syncChan, &wg)

	// Run monitoring
	nbReceivers += 2 // Todo : change that
	wg.Add(1)
	go Monitor(params, dataChan, reportChan, alertChan, syncChan, &wg)

	// Run display to print result
	nbReceivers++
	wg.Add(1)
	go Display(params, reportChan, alertChan, syncChan, &wg)

	// Run command
	wg.Add(1)
	go command(syncChan, nbReceivers, &wg)

	log.Info("Capturing set up.")

	// Shutdown
	<-syncChan
	log.Info("Waiting for all processes to stop.")
	wg.Wait()
	log.Info("Monitoring successfully stopped.")
}
