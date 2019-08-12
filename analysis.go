package gonetmon

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/sirupsen/logrus"
	"net/http"
	"sort"
	"strings"
	"time"
)

const (
	httpResponse = "response"
	httpRequest  = "request"
)

// MetaPacket is a wrapper around a captured packet with some additional information :
// /net/http Request or Response struct
// on which interface the packet was captured
type MetaPacket struct {
	messageType string // Either request or response
	device      string // Interface on which the packet was recorded
	deviceIP    string // IP address of local network device interface
	remoteIP    string // IP address or remote peer

	// Request information
	request *http.Request

	// Response information
	response *http.Response

	// Associated Captured Packet
	packet gopacket.Packet
}

// NewMetaPacket returns a new struct initialised with values from the packetMsg
func NewMetaPacket(data *packetMsg) *MetaPacket {
	return &MetaPacket{
		messageType: "",
		device:      data.device,
		deviceIP:    data.deviceIP,
		remoteIP:    data.remoteIP,
		request:     nil,
		response:    nil,
		packet:      data.rawPacket,
	}
}

type requestStats struct {
	nbReqs    uint            // Sum of all the elements
	nbMethods map[string]uint // Map request methods to the number of times they were encountered
}

type responseStats struct {
	nbResp   uint         // Sum of all registered elements
	nbStatus map[int]uint // Map status codes to the number of times they were encountered
}

type sectionStats struct {
	section  string       // Section of a website
	nbHits   int          // Number of requests that were made for that section
	requests requestStats // Associated statistics
}

// SortedSections implements sort.Interface based on the hit field
type SortedSections []*sectionStats

func (s SortedSections) Len() int           { return len(s) }
func (s SortedSections) Less(i, j int) bool { return s[i].nbHits > s[j].nbHits }
func (s SortedSections) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// hostStats holds information about traffic with a host
type hostStats struct {
	host     string                   // Domain name
	ips      []string                 // IP addresses that were encountered for that host (sort of a local DNS cache)
	hits     int                      // Number of successfully recognised packets associated with that host
	sections map[string]*sectionStats // Statistics about requested sections of that host
	// Statistics about responses on that host
	responses responseStats // Statistics about responses from that hosts
}

// Analysis holds the packets and the result of a recording window
type Analysis struct {
	packets []*MetaPacket      // A set of packets to be analysed
	traffic map[string]int64 // maps device name and corresponding amount of bits
	nbHosts int
	hosts   map[string]*hostStats
	//lastSeenHost *hostStats
}

// Report holds the final result of an analysis, to be sent out to display()
type Report struct {
	topHost        *hostStats
	sortedSections []*sectionStats
	traffic 		map[string]int64
	timestamp 		time.Time
}

// Update statistics of a section with new data
func (a *Analysis) updateSectionStats(hostname string, sectionName string, req *http.Request) {

	host := a.hosts[hostname]
	host.hits++
	//a.lastSeenHost = host
	section := host.sections[sectionName]

	// Update Hits
	section.nbHits++
	section.requests.nbReqs++

	method := req.Method

	// If method was not yet registered, do it
	if _, ok := section.requests.nbMethods[method]; !ok {
		section.requests.nbMethods[method] = 0
	}
	section.requests.nbMethods[method]++
}

// updateResponseStats updates data for hostname with relevant data
func (a *Analysis) updateResponseStats(hostname string, res *http.Response) {

	host := a.hosts[hostname]
	host.hits++
	//a.lastSeenHost = host
	host.responses.nbResp++

	status := res.StatusCode
	// If status code has not yet been encountered, add it
	if _, ok := host.responses.nbStatus[status]; !ok {
		host.responses.nbStatus[status] = 0
	}
	host.responses.nbStatus[status]++
}

// newSectionStats returns an empty set of statistics about a section
func newSectionStats(section string) *sectionStats {
	return &sectionStats{
		section: section,
		nbHits:  0,
		requests: requestStats{
			nbReqs:    0,
			nbMethods: make(map[string]uint),
		},
	}
}

// newHostStats returns an empty set of statistics about a host
func newHostStats(host string) *hostStats {
	return &hostStats{
		host:     host,
		ips:      []string{},
		hits:     0,
		sections: make(map[string]*sectionStats),
		responses: responseStats{
			nbResp:   0,
			nbStatus: make(map[int]uint),
		},
	}
}

// getHost returns the domain name from a http request, and attempts to do so for a http response.
// There's no standard trace of the remote host in the Response header,
// so the only way that's left is to see if we can match the remote address with a host's address we've already seen
// before with a request
func getHost(p *MetaPacket, a *Analysis) (string, error) {

	// If it's a request, it's in the header
	if p.messageType == httpRequest {
		return p.request.Host, nil
	}

	// Verify if the ip corresponds to the last encountered host
	/*for _, ip := range a.lastSeenHost.ips {
		if strings.Compare(ip, p.remoteIP) == 0 {
			return a.lastSeenHost.host, nil
		}
	}*/

	// Iterate over all encountered hosts
	for host, stat := range a.hosts {
		for _, ip := range stat.ips {
			if strings.Compare(ip, p.remoteIP) == 0 {
				return host, nil
			}
		}
	}

	// If no previous host was found, we don't yet have a way to reliably return a host
	return "nil", errors.New("error : http response remote IP matches no known host")
}

// getSection extracts the section from a HTTP Request's URI
func getSection(req *http.Request) string {
	uri := req.RequestURI
	if idx := strings.IndexByte(uri[1:], '/'); idx >= 0 {
		uri = uri[:idx+1]
	}
	// Sometimes requests will hold values, skip them
	if idx := strings.IndexByte(uri[1:], '?'); idx >= 0 {
		uri = uri[:idx+1]
	}
	return uri
}

// registerHostElements adds new remote IP and section to a host if they were not present
func (a *Analysis) registerHostElements(host string, section string, remoteIP string) {

	hosts := a.hosts

	// Verify if remote IP was registered for this host
	b := false
	for _, ip := range hosts[host].ips {
		if strings.Compare(ip, remoteIP) == 0 {
			b = true
		}
	}
	if !b {
		hosts[host].ips = append(hosts[host].ips, remoteIP)
	}

	// If the section is not registered, create new
	if _, ok := hosts[host].sections[section]; !ok {
		// Register new section
		hosts[host].sections[section] = newSectionStats(section)
	}
}

// updateTraffic adds size to calculate traffic speed
func (a *Analysis) updateTraffic(p *MetaPacket) {

	dev := p.device
	var bits int64
	if p.messageType == httpResponse {
		bits = p.response.ContentLength
	} else {
		bits = p.request.ContentLength
	}

	if _, ok := a.traffic[dev]; ok {
		a.traffic[dev] += bits
	} else {
		a.traffic[dev] = bits
	}
}

// updateAnalysis update's the report's current analysis with the new incoming packet information
func (a *Analysis) updateAnalysis(p *MetaPacket) {

	a.updateTraffic(p)

	// If it is a response, we must have seen the corresponding host before, or we cannot work with it
	if p.messageType == httpResponse {
		host, err := getHost(p, a)
		if err != nil {
			log.WithFields(logrus.Fields{
				"remote IP": p.remoteIP,
			}).Error(err)
			return
		}
		a.updateResponseStats(host, p.response)
	} else {

		// Here, it is a request
		host, _ := getHost(p, a)
		section := getSection(p.request)

		hosts := a.hosts

		// If host not registered, create new
		if _, ok := a.hosts[host]; !ok {
			// Register new host and section
			hosts[host] = newHostStats(host)
			hosts[host].ips = append(hosts[host].ips, p.remoteIP)
			hosts[host].sections[section] = newSectionStats(section)
		} else {
			a.registerHostElements(host, section, p.remoteIP)
		}

		// Update statistics
		a.updateSectionStats(host, section, p.request)
	}
}

// AddPacket adds a packet to the report
func (a *Analysis) AddPacket(p *MetaPacket) {
	a.packets = append(a.packets, p)

	a.updateAnalysis(p)
}

// NewAnalysis returns a new and empty Analysis struct
func NewAnalysis() *Analysis {
	return &Analysis{
		packets: nil,
		traffic: make(map[string]int64),
		nbHosts: 0,
		hosts:   make(map[string]*hostStats),
		//lastSeenHost: nil,
	}
}

// NewReport build a new report, containing the host with the most hits
func NewReport(a *Analysis, t time.Time) *Report {

	// If no hosts were registered, we have nothing to report
	if len(a.hosts) == 0 {
		log.Info("No hosts in analysis to build report on.")
		return &Report{
			topHost:        nil,
			sortedSections: nil,
			timestamp:      t,
		}
	}

	// Loop through all encountered hosts and find the first one with most hits
	var topHost *hostStats
	topHits := 0
	for _, stats := range a.hosts {
		if stats.hits > topHits {
			topHits = stats.hits
			topHost = stats
		}
	}

	// This should not happen, as we avoid the case above, but for the sake of it
	if topHost == nil {
		log.Error("Could not find a topHost on a non-empty set of Hosts. THIS SHOULD NOT HAPPEN.")
		return &Report{
			topHost:        nil,
			sortedSections: nil,
			timestamp:      t,
		}
	}

	// Copy sections of host into a slice
	sortedSections := make([]*sectionStats, len(topHost.sections))
	i := 0
	for _, stats := range topHost.sections {
		sortedSections[i] = stats
		i++
	}
	sort.Sort(SortedSections(sortedSections))

	log.Info("sections ", sortedSections)

	return &Report{
		topHost:        topHost,
		sortedSections: sortedSections,
		traffic: a.traffic,
		timestamp:      t,
	}
}
