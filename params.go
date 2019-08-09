// Params loads and holds configuration for runtime
package main

import "time"

const (
	// dataTypes
	dataHTTP = "http"
)

type Filter struct {
	Network     string // BPF filter to filter traffic sniffing
	Application string // Filter to apply at application level : string to search for
	Type        string // Type of expected data
}

// Parameters holds the application's parameters it runs on
type Parameters struct {

	// Raw data parameters
	PacketFilter	Filter
	Interfaces		[]string // Array of interfaces to specifically listen on. If nil, listen on all devices.

	// Analysis related parameters
	AlertSpan      time.Duration // Time (seconds) span to monitor for alert trigger
	AlertThreshold uint          // Traffic (hits/span) threshold triggering an alert

	// Display related parameters
	DisplayRefresh time.Duration // Time (seconds) the display will be updated
	DisplayType    string        // Type of display (for now it's CLI only)
}

// Default values for Parameter object
const (
	defNetworkFilter     = "tcp and port 80"
	defApplicationFilter = "HTTP"
	defApplicationType	 = dataHTTP
	defAlertSpan         = 10 * time.Second
	defAlertThreshold    = 4
	defDisplayRefresh    = 2 * time.Second
	defDisplayType       = "console"

	defRecoveryFormat = "Alert recovered at %s"
	defAlertFormat    = "High traffic generated an alert - hits = %d, triggered at %s"
)

// LoadParams loads the application's parameters it should run on into an object and returns it
func LoadParams() *Parameters {
	// Todo : There should be a better way of doing this + argument validation

	return &Parameters{
		PacketFilter:Filter{
			Network:     defNetworkFilter,
			Application: defApplicationFilter,
			Type:        defApplicationType,
		},
		Interfaces:        nil,
		AlertSpan:         defAlertSpan,
		AlertThreshold:    defAlertThreshold,
		DisplayRefresh:    defDisplayRefresh,
		DisplayType:       defDisplayType,
	}
}
