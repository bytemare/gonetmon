package main

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
	"time"
)

// TODO : don't keep values here
var (
	snapshotLen int32 = 1024
	promiscuous       = false
	timeout           = defDisplayRefresh //10 * time.Second
)

// InitialiseCapture opens device interfaces and associated handles to listen on, returns a map of these.
// If the interfaces parameter is not nil, only open those specified.
func InitialiseCapture(interfaces []string) (map[string]*pcap.Handle, error) {

	var err error

	devices := findDevices(interfaces)

	if devices == nil {
		return nil, err
	}

	m := make(map[string]*pcap.Handle)
	err = nil
	for _, d := range devices {
		if h, err := openDevice(d); err != nil {
			// todo : error
		} else {
			m[d.Name] = h
		}
	}

	if len(m) == 0 {
		log.Error("Could not open any device interface.")
		return nil, errors.New("could not open any device interface")
	}

	return m, nil
}

// findDevices gathers the list of interfaces of the machine.
// If the interfaces parameter is not nil, only list those specified if present.
func findDevices(interfaces []string) []net.Interface {
	devices, err := net.Interfaces()

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error in finding network devices.")
		return nil
	}

	if len(devices) == 0 {
		log.Error("Could not find any network devices (but no error occurred).")
		return nil
	}

	// If we want a custom list of interfaces
	if interfaces != nil {
		var tailoredList []net.Interface

	interfacesLoop:
		for _, i := range interfaces {

			for index, d := range devices {
				if d.Name == i {
					tailoredList = append(tailoredList, d)

					// Remove the found element from array to avoid it on next iteration
					// Won't affect current loop since Go uses a copy
					devices[index] = devices[len(devices)-1]
					devices = devices[:len(devices)-1]

					log.Info("Found requested interface ", i)

					continue interfacesLoop
				}
			}

			// Here, the requested interface is not in the found set
			log.Error("Could not find requested interface : ", i)
		}

		if len(tailoredList) == 0 {
			log.Error("Could not find any requested network devices among : ", interfaces)
			return nil
		}

		devices = tailoredList
	}

	log.Info("Found devices ", len(devices))

	return devices
}

// openDevice opens a live listener on the interface designated by the device parameter and returns a corresponding handle
func openDevice(device net.Interface) (*pcap.Handle, error) {
	handle, err := pcap.OpenLive(device.Name, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.WithFields(log.Fields{
			"interface": device.Name,
			"error":     err,
		}).Error("Could not open device.")

		return nil, err
	}

	log.WithFields(log.Fields{
		"interface": device.Name,
	}).Info("Opened device interface.")

	return handle, nil
}

// Closes listening on a device
func closeDevice(h *pcap.Handle) {
	h.Close()
}

// Opens a list of devices
func openDevices(devs []net.Interface) []*pcap.Handle {
	var handlers []*pcap.Handle
	for _, d := range devs {
		h, err := openDevice(d)
		if h != nil && err == nil {
			handlers = append(handlers, h)
		}
	}
	return handlers
}

// Closes listening on given interfaces through their handle
func closeDevices(handles []*pcap.Handle) {
	log.Info("Closing devices.")
	for _, h := range handles {
		closeDevice(h)
	}
}

func closeMapDevices(devs map[string]*pcap.Handle) {
	for d, h := range devs {
		log.Info("Closing device on interface ", d)
		closeDevice(h)
	}
}

// addFilter adds a BPF filter to the handle to filter sniffed traffic
func addFilter(handle *pcap.Handle, filter string) error {
	return handle.SetBPFFilter(filter)
}

// sniffHTTP tells whether the packet contains HTTP
func sniffHTTP(packet gopacket.Packet) bool {
	var ishttp = false
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		payload := applicationLayer.Payload()
		if strings.Contains(string(payload), "HTTP") {
			/*
				fmt.Print("HTTP found!\n")
				fmt.Printf("\t     Payload : '%s'\n", payload)
				fmt.Printf("\t     Packet Data : '%s'\n", string(packet.Data()))
				fmt.Printf("%s%s\n", strings.Repeat("=", 20), strings.Repeat("\n", 4))
			*/
			ishttp = true
		}
	}

	return ishttp
}

// capturePacket continuously listens to a device interface managed by handle, and extracts relevant packets from traffic
// to send it to dataChan
func capturePackets(handle *pcap.Handle, wg *sync.WaitGroup, dataChan chan<- dataMsg, name string) {
	defer wg.Done()

	log.Info("Capturing packets on ", name)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// This will loop on a channel that will send packages, and will quit when the handle is closed by another caller
	for packet := range packetSource.Packets() {
		if sniffHTTP(packet) {
			dataChan <- dataMsg{
				dataType:  dataHTTP,
				timestamp: time.Now(),
				device:    name,
				payload:   string(packet.ApplicationLayer().Payload()),
			}
		}
	}

	log.Info("Stopping capture on ", name)
}

// Collector listens on all network devices for relevant traffic and sends packets to dataChan
func Collector(parameters *Parameters, devices map[string]*pcap.Handle, dataChan chan dataMsg, syncChan <-chan struct{}, syncwg *sync.WaitGroup) {

	wg := sync.WaitGroup{}

	for dev, h := range devices {
		wg.Add(1)
		if err := addFilter(h, parameters.Filter); err != nil {
			log.WithFields(log.Fields{
				"interface": dev,
				"error":     err,
			}).Error("Could not set filter on device. Closing.")
			closeDevice(h)
		}
		go capturePackets(h, &wg, dataChan, dev)
	}

	// Wait until sync to stop
	<-syncChan

	// Inform goroutines to stop
	closeMapDevices(devices)

	// Wait for goroutines to stop
	log.Info("Collector waiting for subs...")
	wg.Wait()
	log.Info("Collector terminating")
	syncwg.Done()
}
