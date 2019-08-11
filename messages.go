package main

import (
	"github.com/google/gopacket"
	"net"
	"time"
)

type packetMsg struct {
	dataType  string // Kind of data, for now just http packet
	device    string // Interface on which the traffic was recorded
	deviceIP  net.IP // IP address of local network device interface
	remoteIP  net.IP // IP address or remote peer
	rawPacket gopacket.Packet // Actual packet payload
}

type reportMsg struct {
	report    *Report // Message body to display
	timestamp time.Time
}

type alertMsg struct {
	recovery  bool   // True if we recover from alert to no alert, false if not
	body      string // Message to display
	timestamp time.Time
}