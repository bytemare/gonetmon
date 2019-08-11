// Params loads and holds configuration for runtime
package main

import (
	"sync"
	"time"
)

const (
	// dataTypes
	dataHTTP = "http"
)

type CaptureConfig struct {
	SnapshotLen     int32         // Maximum size to read for each packet
	PromiscuousMode bool          // Whether to ut the interface in promiscuous mode
	CaptureTimeout  time.Duration // Period to listen for traffic before sending out captured traffic
}

type Filter struct {
	Network     string // BPF filter to filter traffic at data layer
	Application string // String to look for in Application Layer
	Type        string // Monitor filter in case further development adds other traffic analysis
}

// Sync is a placeholder for synchronisation tools across goroutines
type Sync struct {
	wg          sync.WaitGroup
	syncChan    chan struct{}
	nbReceivers uint
}

// addRoutine increments the number of goroutines to be synced and waiting for a message on the channel
func (s *Sync) addRoutine() {
	s.wg.Add(1)
	s.nbReceivers++
}

// Parameters holds the application's parameters it runs on
type Parameters struct {

	// Raw data parameters
	PacketFilter  Filter
	CaptureConfig CaptureConfig
	Interfaces    []string // Array of interfaces to specifically listen on. If nil, listen on all devices.

	// Display related parameters
	DisplayRefresh time.Duration // Period (seconds) to renew display print, thus also used for capture and reporting
	DisplayType    string        // Type of display output

	// Analysis related parameters
	AlertSpan       time.Duration // Time (seconds) frame to monitor (and retain) traffic behaviour
	AlertThreshold  uint          // Number of request over time frame (hits/span) that will trigger an alert
	WatchdogTick    time.Duration // Period (milliseconds, preferably) over which to check for alerts
	WatchdogBufSize uint          // Size of the channel used to receive hit notification. Make it arbitrarily high. TODO: There may be a better way to do this
}

// Default values for Parameter object
const (
	// Capture default
	defNetworkFilter           = "tcp and port 80"
	defApplicationFilter       = "HTTP"
	defApplicationType         = dataHTTP
	defSnapshotLen       int32 = 1024
	defPromiscuousMode         = false
	defCaptureTimeout          = defDisplayRefresh

	// Display Parameters
	defDisplayRefresh = 5 * time.Second
	defDisplayType    = "console" // Default output destination

	// Format strings for display
	defAlertFormat    = "High traffic generated an alert - hits = %d, triggered at %s"
	defRecoveryFormat = "Alert recovered at %s"

	// Watchdog defaults
	defAlertSpan        = 10 * time.Second
	defAlertThreshold   = 4
	defaultWatchdogTick = 500 * time.Millisecond
	defaultBufSize      = 1000
)

// LoadParams loads the application's parameters it should run on into an object and returns it
func LoadParams() *Parameters {
	// Todo : There should be a better way of doing this + argument validation

	return &Parameters{
		PacketFilter: Filter{
			Network:     defNetworkFilter,
			Application: defApplicationFilter,
			Type:        defApplicationType,
		},
		CaptureConfig: CaptureConfig{
			SnapshotLen:     defSnapshotLen,
			PromiscuousMode: defPromiscuousMode,
			CaptureTimeout:  defCaptureTimeout,
		},
		Interfaces:      nil,
		DisplayRefresh:  defDisplayRefresh,
		DisplayType:     defDisplayType,
		AlertSpan:       defAlertSpan,
		AlertThreshold:  defAlertThreshold,
		WatchdogTick:    defaultWatchdogTick,
		WatchdogBufSize: defaultBufSize,
	}
}
