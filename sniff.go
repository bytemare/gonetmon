package gonetmon

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// Init initialises Sniffing and Monitoring
// TODO: Load configuration from file or command line to initialise parameters
func Init() (*devices, error) {

	// Check whether we can capture packets
	devices, err := InitialiseCapture()
	if err != nil {
		return nil, fmt.Errorf("initialising capture failed : %s", err)
	}

	// Past this point, log to file
	file, err := os.OpenFile(defLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	return devices, nil
}

// Sniff holds examples of initialising a session and manage different routines to perform monitoring
func Sniff(testWait *sync.WaitGroup, result chan<- error) error {
	if testWait != nil {
		defer testWait.Done()
	}

	// Initialise, and fail if conditions are not met
	devices, err := Init()
	if err != nil {
		log.Error(err)
		if result != nil {
			result <- err
		}
		return err
	}

	// IPCs
	syn := &synchronisation{
		wg:          sync.WaitGroup{},
		syncChan:    make(chan struct{}),
		nbReceivers: 0,
	}
	syn.addRoutine() // add this main process

	packetChan := make(chan packetMsg, 1000)
	reportChan := make(chan *report, 1)
	alertChan := make(chan alertMsg, 1)

	// Run Sniffer/Collector
	syn.addRoutine()
	go Collector(devices, packetChan, syn)

	// Run monitoring
	syn.addRoutine()
	go Monitor(packetChan, reportChan, alertChan, syn)

	// Run display to print result
	syn.addRoutine()
	go Display(reportChan, alertChan, syn)

	// Run CLI
	syn.addRoutine()
	go CLI(syn)

	log.Info("Capturing set up.")

	// synchronisation and shutdown
	syn.wg.Done()
	<-syn.syncChan
	log.Info("Waiting for all processes to stop.")
	syn.wg.Wait()
	log.Info("Monitoring successfully stopped.")
	if result != nil {
		result <- nil
	}
	return nil
}

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
