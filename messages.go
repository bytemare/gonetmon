package main

import (
	"github.com/google/gopacket"
	"time"
)

type packetMsg struct {
	dataType  string // Kind of data, for now just http packet
	device    string // Interface on which the traffic was recorded
	rawPacket gopacket.Packet // Actual packet payload
	timestamp time.Time
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