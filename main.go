package main

import (
	"fmt"
)

func main() {

	// Load parameters
	params := LoadParams()

	// IPCs
	//wg := sync.WaitGroup{}
	var nbReceivers = 1
	dataChan := make(chan dataMsg)
	syncChan := make(chan struct{})

	// Run Sniffer/Collector
	//wg.Add(1)
	nbReceivers++
	go Collector(params, dataChan, syncChan)

	fmt.Println("\n Collector launched.")

	// Run monitoring
	//go Monitor(params, comsChan, syncChan)

	// Run display to print result
	//go Display(params, comsChan, syncChan)

	// Run Interface
	//wg.Add(1)
	go Interface(syncChan, nbReceivers)

	fmt.Println("\n Sniffing ready.")

	// Shutdown
	<-syncChan
	// TODO : proper synchronisation, this here ends before collector shutsdown
	fmt.Println("\n Sniffing stopped.")
}