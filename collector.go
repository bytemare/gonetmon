package main

import (
	"fmt"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var (
	snapshot_len int32  = 1024
	promiscuous  bool   = false
	timeout      time.Duration = 10 * time.Second
)

// findDevices gathers the list of interfaces of the machine
func findDevices() []net.Interface {
	devs, err := net.Interfaces()
	// TODO handle error
	if err != nil {
		panic(err)
	}
	return devs
}

// Print the list of devices
func printDevices(devices []net.Interface) {
	for _, f := range devices {
		fmt.Println(f.Name)
	}
}

// openDevice opens a live listener on the interface designated by the device parameter and returns a corresponding handle
func openDevice(device net.Interface) *pcap.Handle {
	handle, err := pcap.OpenLive(device.Name, snapshot_len, promiscuous, timeout)
	if err != nil {log.Fatal(err) }
	//defer handle.Close()
	return handle
}

// Closes listening on a device
func closeDevice(h *pcap.Handle) {
	h.Close()
}

// Opens a list of devices
func openDevices(devs []net.Interface) []*pcap.Handle {
	var handlers []*pcap.Handle
	for _, d := range devs {
		handlers = append(handlers, openDevice(d))
	}
	return handlers
}

// Closes listening on given interfaces through their handle
func closeDevices(handles []*pcap.Handle) {
	for _, h := range handles {
		closeDevice(h)
	}
}

// addFilter adds a BPF filter to the handle to filter sniffed traffic
func addFilter(handle *pcap.Handle, filter string) {
	err := handle.SetBPFFilter(filter)
	if err != nil {
		log.Fatal(err)
	}
}

// sniffHTTP tells whether the packet contains HTTP
func sniffHTTP(packet gopacket.Packet) bool {
	var ishttp = false
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {

		payload := applicationLayer.Payload()
		if strings.Contains(string(payload), "HTTP") {
			fmt.Print("HTTP found!\n")
			fmt.Printf("\t     Payload : '%s'\n", payload)
			fmt.Printf("\t     Packet Data : '%s'\n", string(packet.Data()))
			fmt.Printf("%s%s\n", strings.Repeat("=", 20), strings.Repeat("\n", 4))
			ishttp = true
		}
	}

	return ishttp
}

// capturePacket continuously listens to a device interface managed by handle, and extracts relevant packets from traffic
// to send it to dataChan
func capturePackets(handle *pcap.Handle, wg *sync.WaitGroup, dataChan chan<- dataMsg, name string) {
	defer wg.Done()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// This will loop on a channel that will send packages, and will quit when the handle is closed by another caller
	for packet := range packetSource.Packets() {
				if sniffHTTP(packet) {
					dataChan <- dataMsg{
						dataType:  dataHTTP,
						timestamp: time.Now(),
						device: name,
						body: string(packet.ApplicationLayer().Payload()),
					}
				}
	}

	fmt.Printf("Stop listening on %s\n", name)
}


// Collector listens on all network devices for relevant traffic and sends packets to dataChan
func Collector(parameters *Parameters, dataChan chan dataMsg, syncChan <-chan struct{}) {

	devices := findDevices()

	printDevices(devices)

	handles := openDevices(devices)

	wg := sync.WaitGroup{}

	for i, h := range handles {
		fmt.Println("Capturing packets on", devices[i].Name)
		wg.Add(1)
		addFilter(h, parameters.Filter)
		go capturePackets(h, &wg, dataChan, devices[i].Name)
	}

	// Wait until sync to stop
	fmt.Println("\nCollector waiting for signal...")
	<-syncChan

	// Inform goroutines to stop
	closeDevices(handles)

	// Wait for goroutines to stop
	fmt.Println("Collector waiting for subs...")
	wg.Wait()
	fmt.Println("Collector terminating")
}
