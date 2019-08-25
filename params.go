package gonetmon

import (
	"sync"
	"time"
)

const (
	// dataTypes
	dataHTTP = "http"

	// output
	consoleOutput = "console"
	//fileOutput    = ""
)

// Default values for program parameters
const (
	// Capture default
	defNetworkFilter           = "tcp and port 80"
	defApplicationFilter       = "HTTP"
	defApplicationType         = dataHTTP
	defNbSection               = 3
	defSnapshotLen       int32 = 1024
	defPromiscuousMode         = false
	defCaptureTimeout          = defDisplayRefresh

	// Display configuration
	defDisplayRefresh = 10 * time.Second
	defDisplayType    = consoleOutput // Default output destination

	// Format strings for display
	defAlertFormat    = "High traffic generated an alert - hits = %d, triggered at %s"
	defRecoveryFormat = "Alert recovered at %s"

	// watchdog defaults
	defAlertSpan        = 120 * time.Second
	defAlertThreshold   = 7000
	defaultWatchdogTick = 500 * time.Millisecond
	defaultBufSize      = 1000

	// General
	defLogFile    = "./log-gonetmon.log"
	defTimeLayout = "2006-01-02 15:04:05.124"
)

// captureConfig holds configuration for capturing packets
type captureConfig struct {
	snapshotLen     int32         // Maximum size to read for each packet
	promiscuousMode bool          // Whether to ut the interface in promiscuous mode
	captureTimeout  time.Duration // Period to listen for traffic before sending out captured traffic
}

// filter holds different filters on different levels to apply and tag data
type filter struct {
	network     string // BPF filter to filter traffic at data layer
	application string // String to look for in application Layer
	dataType    string // Monitor filter in case further development adds other traffic analysis
	nbSections  int    // Number of sections to retain for top sections display
}

// synchronisation is a placeholder for synchronisation tools across goroutines
type synchronisation struct {
	wg          sync.WaitGroup
	syncChan    chan struct{}
	nbReceivers uint
}

// addRoutine increments the number of goroutines to be synced and waiting for a message on the channel
func (s *synchronisation) addRoutine() {
	s.wg.Add(1)
	s.nbReceivers++
}

// alertVars analysis related parameters
type alertVars struct {
	span            time.Duration // Time (seconds) frame to monitor (and retain) traffic behaviour
	threshold       int           // Number of request over time frame (hits/span) that will trigger an alert
	watchdogTick    time.Duration // Period (milliseconds, preferably) over which to check for alerts
	watchdogBufSize uint          // Size of the channel used to receive hit notification. Make it arbitrarily high. TODO: There may be a better way to do this
}

// configuration holds the application's parameters it runs on
type configuration struct {

	// Raw data parameters
	packetFilter        filter
	captureConf         captureConfig
	requestedInterfaces []string // Array of interfaces to specifically listen on. If nil, listen on all devices.

	// Display related parameters
	displayRefresh time.Duration // Period (seconds) to renew display print, thus also used for capture and reporting
	displayType    string        // Type of display output

	alert alertVars
}

// LoadParams loads the application's parameters it should run on into an object and returns it
func LoadParams() *configuration {
	// Todo : There should be a better way of doing this + argument validation

	return &configuration{
		packetFilter: filter{
			network:     defNetworkFilter,
			application: defApplicationFilter,
			dataType:    defApplicationType,
			nbSections:  defNbSection,
		},
		captureConf: captureConfig{
			snapshotLen:     defSnapshotLen,
			promiscuousMode: defPromiscuousMode,
			captureTimeout:  defCaptureTimeout,
		},
		requestedInterfaces: nil,
		displayRefresh:      defDisplayRefresh,
		displayType:         defDisplayType,
		alert: alertVars{
			span:            defAlertSpan,
			threshold:       defAlertThreshold,
			watchdogTick:    defaultWatchdogTick,
			watchdogBufSize: defaultBufSize,
		},
	}
}
