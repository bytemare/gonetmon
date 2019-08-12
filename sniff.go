package gonetmon

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"time"
)

var log = logrus.New()

// Init initialises Sniffing and Monitoring
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

	// Past this point, log to file
	file, err := os.OpenFile(defLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		log.Out = file
	} else {
		log.Info("Failed to log to file, using default stderr")
	}

	return params, devices, nil
}

// Sniff is an example use of the tool
func Sniff(testWait *sync.WaitGroup) {
	defer testWait.Done()
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
	syn.addRoutine() // add this main process

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
	syn.wg.Done()
	<-syn.syncChan
	log.Info("Waiting for all processes to stop.")
	syn.wg.Wait()
	log.Info("Monitoring successfully stopped.")
}

// SnifferTest is a test function for Sniffer use with a timeout
func SnifferTest(duration time.Duration) {

	testWait := sync.WaitGroup{}
	testWait.Add(1)
	go Sniff(&testWait)

	// Send interrupt signal after timeout
	p, _ := os.FindProcess(os.Getpid())
	select {
	case <-time.After(duration):
		_ = p.Signal(os.Interrupt)
	}
	testWait.Wait()
}
