package main

import (
	"fmt"
)

func main() {

	//TODO : verify if we are running with sudo in order to be able to work

	// Load parameters
	params := LoadParams()

	// IPCs
	//wg := sync.WaitGroup{}
	var nbReceivers = 1
	dataChan := make(chan dataMsg, 1000)
	reportChan := make(chan reportMsg, 1)
	alertChan := make(chan alertMsg, 1)
	syncChan := make(chan struct{})

	// Run Sniffer/Collector
	//wg.Add(1)
	nbReceivers++
	go Collector(params, dataChan, syncChan)

	fmt.Println("\n Collector launched.")

	// Run monitoring
	go Monitor(params, dataChan, reportChan, alertChan, syncChan)

	// Run display to print result
	go Display(params, reportChan, alertChan, syncChan)

	// Run command
	//wg.Add(1)
	go command(syncChan, nbReceivers)

	fmt.Println("\n Sniffing ready.")

	// Shutdown
	<-syncChan
	// TODO : proper synchronisation, this here ends before collector may shutdown properly
	fmt.Println("\n Sniffing stopped.")
}
