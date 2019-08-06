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
	timeout      time.Duration = 30 * time.Second
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
	handle, err := pcap.OpenLive(device.Name, snapshot_len, promiscuous, timeout)
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
			fmt.Println("HTTP found!\n")
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
func capturePackets(handle *pcap.Handle, wg *sync.WaitGroup, dataChan chan<- dataMsg, stopChan <-chan int, name string) {
	defer wg.Done()

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	packetChan := packetSource.Packets()

	loop:
	for {
		select {

		case <-stopChan:
			break loop

		case packet := <-packetChan:
			if sniffHTTP(packet) {
				dataChan <- dataMsg{
					dataType:  dataHTTP,
					timestamp: time.Now(),
					device: name,
					body: string(packet.ApplicationLayer().Payload()),
				}
			}
		}
	}

	/*
	for packet := range packetSource.Packets() {
		// Process packet here
		// fmt.Println(packet)
		sniffHTTP(packet)
	}
	 */

	fmt.Println("Stop listening")
}


// Collector listens on all network devices for relevant traffic and sends packets to dataChan
func Collector(parameters *Parameters, dataChan chan<- dataMsg, syncChan <-chan int) {
	devices := findDevices()

	listDevices(devices)

	handles := openDevices(devices)
	defer closeDevices(handles)

	wg := sync.WaitGroup{}
	wg.Add(len(handles))

	stopChan := make(chan int)

	for i, h := range handles {
		fmt.Println("Capturing packets on", devices[i].Name)
		addFilter(h, "tcp and port 80")
		go capturePackets(h, &wg, dataChan, stopChan, devices[i].Name)
	}

	// Receive sync to stop
	<-syncChan

	// Inform goroutines to stop
	close(stopChan)

	// Wait for goroutines to stop
	wg.Wait()
}
