package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strings"
	"time"
)

var (
	//device       string = "eth0"
	snapshot_len int32  = 1024
	promiscuous  bool   = false
	err          error
	timeout      time.Duration = 30 * time.Second
	handle       *pcap.Handle
	// Will reuse these for each packet
	ethLayer layers.Ethernet
	ipLayer  layers.IPv4
	tcpLayer layers.TCP
)

func findDevices() []net.Interface {
	devs, err := net.Interfaces()
	// TODO handle error
	if err != nil {
		panic(err)
	}
	return devs
}

func listDevices(devices []net.Interface) {
	for _, f := range devices {
		fmt.Println(f.Name)
	}
}

func openDevice(device net.Interface) *pcap.Handle {
	handle, err = pcap.OpenLive(device.Name, snapshot_len, promiscuous, timeout)
	if err != nil {log.Fatal(err) }
	//defer handle.Close()
	return handle
}

func closeDevice(h *pcap.Handle) {
	h.Close()
}

func openDevices(devs []net.Interface) []*pcap.Handle {
	var handlers []*pcap.Handle
	for _, d := range devs {
		handlers = append(handlers, openDevice(d))
	}
	return handlers
}

func closeDevices(handles []*pcap.Handle) {
	for _, h := range handles {
		closeDevice(h)
	}
}

func addFilter(handle *pcap.Handle, filter string) {
	err = handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
}

func capturePackets(handle *pcap.Handle) {
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		// Process packet here
		// fmt.Println(packet)
		processPacket(packet)
	}
}

func processPacket(packet gopacket.Packet) {
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		fmt.Println("Application layer/Payload found.")
		fmt.Printf("%s\n", applicationLayer.Payload())

		// Search for a string inside the payload
		if strings.Contains(string(applicationLayer.Payload()), "HTTP") {
			fmt.Println("HTTP found!")
		}
	}
}

func Collector(parameters *Parameters, syncChan chan int) {
	devices := findDevices()

	listDevices(devices)

	handles := openDevices(devices)
	defer closeDevices(handles)

	for i, h := range handles {
		if devices[i].Name == "enp0s3" {
			fmt.Println("Capturing packets on enp0s3")

			addFilter(h, "tcp and port 443")
			fmt.Println("Only capturing TCP port 80 packets.")

			capturePackets(h)
		}
	}
}
