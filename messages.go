package gonetmon

import (
	"github.com/google/gopacket"
	"time"
)

type packetMsg struct {
	dataType  string          // Kind of data, for now just http packet
	device    string          // Interface on which the traffic was recorded
	deviceIP  string          // IP address of local network device interface
	remoteIP  string          // IP address or remote peer
	rawPacket gopacket.Packet // Actual packet payload
}

type alertMsg struct {
	recovery  bool   // True if we recover from alert to no alert, false if not
	body      string // Message to display
	timestamp time.Time
}
