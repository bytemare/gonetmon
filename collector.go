package gonetmon

import (
	"errors"
	"fmt"
	"github.com/google/gopacket"
	//_ "github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
)

// devices is a couple of arrays to hold corresponding devices with their handles
type devices struct {
	devices []net.Interface
	handles []*pcap.Handle
}

// InitialiseCapture opens device interfaces and associated handles to listen on, returns a map of these.
// If the interfaces parameter is not nil, only open those specified.
func InitialiseCapture() (*devices, error) {

	interfaceDevices := findDevices(config.requestedInterfaces)
	if interfaceDevices == nil {
		return nil, errors.New("could not find any devices")
	}

	devs := &devices{
		devices: []net.Interface{},
		handles: []*pcap.Handle{},
	}

	for _, d := range interfaceDevices {
		// Try to open all devices for capture
		if h, err := openDevice(d, &config.captureConf); err != nil {
			log.WithFields(logrus.Fields{
				"error": err,
			}).Error("Could not open device for capture.")
		} else {
			devs.devices = append(devs.devices, d)
			devs.handles = append(devs.handles, h)
		}
	}

	if len(devs.devices) == 0 {
		log.Error("Could not open any device interface.")
		return nil, errors.New("could not open any device interface")
	}

	return devs, nil
}

// selectDevices returns an array of requested interfaces among those available in the devices argument
func selectDevices(requestedInterfaces []string, devices []net.Interface) ([]net.Interface, error) {
	var tailoredList []net.Interface

interfacesLoop:
	for _, i := range requestedInterfaces {

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
		log.Error("Could not find requested interface among activated interfaces : ", i)
	}

	if len(tailoredList) == 0 {
		return nil, fmt.Errorf("could not find any requested network devices among : %s", requestedInterfaces)
	}

	return tailoredList, nil
}

// findDevices gathers the list of interfaces of the machine that have their state flage UP.
// If the interfaces parameter is not nil, only list those specified if present.
func findDevices(requestedInterfaces []string) []net.Interface {
	devices, err := net.Interfaces()

	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err,
		}).Error("Error in finding network devices.")
		return nil
	}

	if devices == nil || len(devices) == 0 {
		log.Error("Could not find any network devices (but no error occurred).")
		return nil
	}

	// Purge interfaces that don't have their state flag UP
	cpy := devices[:0]
	for _, d := range devices {
		if d.Flags&(net.FlagUp) == 1 {
			// Flag is up, Interface is activated, keep element
			cpy = append(cpy, d)
		}
	}

	// If we want a custom list of interfaces
	if requestedInterfaces != nil {
		devices, err = selectDevices(requestedInterfaces, devices)
		if err != nil {
			log.Error(err)
			return nil
		}
	}

	return devices
}

// openDevice opens a live listener on the interface designated by the device parameter and returns a corresponding handle
func openDevice(device net.Interface, config *captureConfig) (*pcap.Handle, error) {
	handle, err := pcap.OpenLive(device.Name, config.snapshotLen, config.promiscuousMode, config.captureTimeout)
	if err != nil {
		log.WithFields(logrus.Fields{
			"interface": device.Name,
			"error":     err,
		}).Error("Could not open device.")

		return nil, err
	}

	log.WithFields(logrus.Fields{
		"interface": device.Name,
	}).Info("Opened device interface.")

	return handle, nil
}

// closeDevice closes listening on a device
func closeDevice(h *pcap.Handle) {
	h.Close()
}

// closeDevices closes all devices given
func closeDevices(devices *devices) {
	for index, dev := range devices.devices {
		log.Info("Closing device on interface ", dev.Name)
		closeDevice(devices.handles[index])
	}
}

// addFilter adds a BPF filter to the handle to filter sniffed traffic
func addFilter(handle *pcap.Handle, filter string) error {
	return handle.SetBPFFilter(filter)
}

// sniffApplicationLayer tells whether the packet contains the filter string in its application layer
func sniffApplicationLayer(packet gopacket.Packet, filter string) bool {
	var isApp = false
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		payload := applicationLayer.Payload()
		if strings.Contains(string(payload), filter) {
			isApp = true
		}
	}

	return isApp
}

// getRemoteIP extracts the IP address of the remote peer from packet
func getRemoteIP(packet gopacket.Packet, deviceIP string) string {
	src, dst := packet.NetworkLayer().NetworkFlow().Endpoints()

	var rip string

	// The deviceIP is among these two, so we return the other
	if strings.Compare(deviceIP, src.String()) == 0 {
		rip = dst.String()
	} else {
		rip = src.String()
	}

	log.Info("Remote peer address ", rip)

	return rip
}

// getDeviceIP extracts the interface's local IP address
func getDeviceIP(device *net.Interface) (string, error) {
	add, err := device.Addrs()
	if err != nil {
		return "", err
	}
	// Don't keep the network mask
	address := add[0].String()[:strings.IndexByte(add[0].String(), '/')]
	return address, nil
}

// capturePacket continuously listens to a device interface managed by handle, and extracts relevant packets from traffic
// to send it to packetChan
func capturePackets(device net.Interface, handle *pcap.Handle, filter *filter, wg *sync.WaitGroup, packetChan chan<- packetMsg) {
	defer wg.Done()

	log.Info("Capturing packets on ", device.Name)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// This will loop on a channel that will send packages, and will quit when the handle is closed by another caller
	for packet := range packetSource.Packets() {
		if sniffApplicationLayer(packet, filter.application) {

			ip, err := getDeviceIP(&device)
			if err != nil {
				log.WithFields(logrus.Fields{
					"interface": device.Name,
					"error":     err,
				}).Error("Could not extract IP from local network interface")
			}

			packetChan <- packetMsg{
				dataType:  filter.dataType,
				device:    device.Name,
				deviceIP:  ip,
				remoteIP:  getRemoteIP(packet, ip),
				rawPacket: packet,
			}
		}
	}

	log.Info("Stopping capture on ", device.Name)
}

// Collector listens on all network devices for relevant traffic and sends packets to packetChan
// Behaviour and filters can be given as argument with parameters
func Collector(devices *devices, packetChan chan packetMsg, syn *synchronisation) {
	defer syn.wg.Done()

	collWG := sync.WaitGroup{}

	for index, dev := range devices.devices {
		collWG.Add(1)
		h := devices.handles[index]
		if err := addFilter(h, config.packetFilter.network); err != nil {
			log.WithFields(logrus.Fields{
				"interface": dev.Name,
				"error":     err,
			}).Error("Could not set filter on device. Closing.")
			closeDevice(h)
		}
		go capturePackets(dev, h, &config.packetFilter, &collWG, packetChan)
	}

	// Wait until sync to stop
	<-syn.syncChan

	// Inform goroutines to stop by closing their handles
	closeDevices(devices)

	// Wait for goroutines to stop
	log.Info("Collector waiting for subs...")
	collWG.Wait()
	log.Info("Collector terminating")
}
