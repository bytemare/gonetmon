package main

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

// Initialises Sniffing and Monitoring
// TODO: Load configuration from file or command line to initialise parameters
func Init() (*Parameters, *Devices, error) {

	// Must be root or sudo
	if os.Geteuid() != 0 {
		log.Error("Geteuid is not 0 : not running with elevated privileges.")
		return nil, nil, errors.New("you must run this program with elevated privileges in order to capture traffic. Try running with sudo")
	}

	// Load default parameters
	params := LoadParams()

	// Check whether we can capture packets
	devices, err := InitialiseCapture(params)
	if err != nil {
		return nil, nil, fmt.Errorf("initialising capture failed : %s", err)
	}

	return params, devices, nil
}

func main() {

	params, devices, err := Init()
	if err != nil {
		log.Fatal(err)
	}

	// IPCs
	syn := &Sync{
		wg:          sync.WaitGroup{},
		syncChan:    make(chan struct{}),
		nbReceivers: 0,
	}

	//var nbReceivers = 1
	//var wg sync.WaitGroup
	packetChan := make(chan packetMsg, 1000)
	reportChan := make(chan *Report, 1)
	alertChan := make(chan alertMsg, 1)

	// Run Sniffer/Collector
	syn.addRoutine()
	go Collector(params, devices, packetChan, syn)

	// Run monitoring
	syn.addRoutine()
	go Monitor(params, packetChan, reportChan, alertChan, syn)

	// Run display to print result
	syn.addRoutine()
	go Display(params, reportChan, alertChan, syn)

	// Run command
	syn.addRoutine()
	go command(syn)

	log.Info("Capturing set up.")

	// Shutdown
	<-syn.syncChan
	log.Info("Waiting for all processes to stop.")
	syn.wg.Wait()
	log.Info("Monitoring successfully stopped.")
}
