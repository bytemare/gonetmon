package main

import (
	"fmt"
)

func main() {

	// Load parameters
	params := LoadParams()

	// IPC channels
	//comsChan := make(chan coms)
	dataChan := make(chan dataMsg)
	syncChan := make(chan int)

	// Run Sniffer/Collector
	Collector(params, dataChan, syncChan)

	// Run monitoring
	//go Monitor(params, comsChan, syncChan)

	// Run Interface
	//go Interface(params, comsChan, syncChan)

	// Run display to print result
	//go Display(params, comsChan, syncChan)

	// Shutdown
	fmt.Println("\n Sniffing stopped.")
}