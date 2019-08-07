// Params loads and holds configuration for runtime
package main

// Parameters holds the application's parameters it runs on
type Parameters struct {

	// Raw data parameters
	Filter			string	// BPF filter to filter traffic sniffing

	// Analysis related parameters
	AlertSpan			int	// Time (seconds) span to monitor for alert trigger
	AlertThreshold		int	// Traffic (hits/span) threshold triggering an alert

	// Display related parameters
	DisplayRefresh	int		// Time (seconds) the display will be updated
	DisplayType		string	// Type of display (for now it's CLI only)
}

// Default values for Parameter object
const (
	defFilter			=	"tcp and port 80"
	defAlertSpan		=	120
	defAlertThreshold	=	500
	defDisplayRefresh	=	10
	defDisplayType		=	"console"
)

// LoadParams loads the application's parameters it should run on into an object and returns it
func LoadParams() *Parameters{
	// Todo : There should be a better way of doing this + argument validation

	return &Parameters{
		Filter:			defFilter,
		AlertSpan:		defAlertSpan,
		AlertThreshold:	defAlertThreshold,
		DisplayRefresh:	defDisplayRefresh,
		DisplayType:	defDisplayType,
	}
}