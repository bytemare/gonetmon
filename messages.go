// Messages defines format for some messages that are sent over channels between routines
package gonetmon

import (
	"github.com/google/gopacket"
	"time"
)

// packetMsg holds information and metadata about a captured packet after a filter was applied
type packetMsg struct {
	dataType  string          // Kind of data, for now just http packet
	device    string          // Interface on which the traffic was recorded
	deviceIP  string          // IP address of local network device interface
	remoteIP  string          // IP address or remote peer
	rawPacket gopacket.Packet // Actual packet payload
}

// alertMsg holds information about alert status updates
type alertMsg struct {
	recovery  bool   		// True if we recover from alert to no alert, false if not
	body      string 		// Message to display
	timestamp time.Time		// Date of alert or recovery
}
