package main

import (
	"errors"
	"github.com/google/gopacket"
	log "github.com/sirupsen/logrus"
	"net/http"
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

type requestStats struct {
	nbReqs    uint            // Sum of all the elements
	nbMethods map[string]uint // Map request methods to the number of times they were encountered
}

type responseStats struct {
	nbResp   uint         // Sum of all registered elements
	nbStatus map[int]uint // Map status codes to the number of times they were encountered
}

type sectionStats struct {
	section  string
	nbHits   int
	requests requestStats
}

type hostStats struct {
	host     string
	ips      []string
	sections map[string]*sectionStats
	// Statistics about responses on that host
	responses responseStats
}

// analysis holds the packets and the result of a recording window
type analysis struct {
	packets      []*MetaPacket // A set of packets to be analysed
	nbHosts      int
	hosts        map[string]*hostStats
	lastSeenHost *hostStats
}

// Report holds the final result of an analysis, to be sent out to display()
type Report struct {
	topDomain string
	responses responseStats
	sections  sectionStats
	analysis  analysis
}

// Update statistics of a section with new data
func (a *analysis) updateSectionStats(hostname string, sectionName string, req *http.Request) {

	host := a.hosts[hostname]
	a.lastSeenHost = host
	section := host.sections[sectionName]

	// Update Hits
	section.nbHits++
	section.requests.nbReqs++

	method := req.Method

	// If method was not yet registered, do it
	if _, ok := section.requests.nbMethods[method]; ok == false {
		section.requests.nbMethods[method] = 0
	}
	section.requests.nbMethods[method]++
}

// updateResponseStats updates data for hostname with relevant data
func (a *analysis) updateResponseStats(hostname string, res *http.Response) {

	host := a.hosts[hostname]

	a.lastSeenHost = host
	host.responses.nbResp++

	status := res.StatusCode
	// If status code has not yet been encountered, add it
	if _, ok := host.responses.nbStatus[status]; ok == false {
		host.responses.nbStatus[status] = 0
	}
	host.responses.nbStatus[status]++
}

// NewSectionStats returns an empty set of statistics about a section
func NewSectionStats(section string) *sectionStats {
	return &sectionStats{
		section: section,
		nbHits:  0,
		requests: requestStats{
			nbReqs:    0,
			nbMethods: make(map[string]uint),
		},
	}
}

func NewHostStats(host string) *hostStats {
	return &hostStats{
		host:     host,
		ips:      []string{},
		sections: make(map[string]*sectionStats),
		responses: responseStats{
			nbResp:   0,
			nbStatus: make(map[int]uint),
		},
	}
}

func getHost(p *MetaPacket, a *analysis) (string, error) {

	// If it's a request, it's in the header
	if p.messageType == httpRequest {
		return p.request.Host, nil
	}

	// Verify if the ip corresponds to the last encountered host
	for _, ip := range a.lastSeenHost.ips {
		if strings.Compare(ip, p.remoteIP) == 0 {
			return a.lastSeenHost.host, nil
		}
	}

	// Iterate over all encountered hosts
	for host, stat := range a.hosts {
		for _, ip := range stat.ips {
			if strings.Compare(ip, p.remoteIP) == 0 {
				return host, nil
			}
		}
	}

	return "nil", errors.New("error : http response remote IP matches to no known host")
}

func getSection(p *MetaPacket) (string, error) {
	if p.messageType == httpRequest {
		uri := p.request.RequestURI
		if idx := strings.IndexByte(uri[1:], '/'); idx >= 0 {
			uri = uri[:idx+1]
		}
		return uri, nil
	}

	return "", errors.New("could not extract section from packet")
}

// updateAnalysis update's the report's current analysis with the new incoming packet information
func (a *analysis) updateAnalysis(p *MetaPacket) {

	// If it is a response, we must have seen the corresponding host before, or we cannot work with it
	if p.messageType == httpResponse {
		host, err := getHost(p, a)
		if err != nil {
			log.WithFields(log.Fields{
				"remote IP": p.remoteIP,
			}).Error(err)
			return
		}
		a.updateResponseStats(host, p.response)
	} else {

		// Here, it is a request
		host, _ := getHost(p, a)
		section, _ := getSection(p)

		hosts := a.hosts

		// If host not registered, create new
		if _, ok := a.hosts[host]; ok == false {
			// Register new host and section
			hosts[host] = NewHostStats(host)
			hosts[host].ips = append(hosts[host].ips, p.remoteIP)
			hosts[host].sections[section] = NewSectionStats(section)
		} else {

			// Verify if remote IP was registered for this host
			b := false
			for _, ip := range hosts[host].ips {
				if strings.Compare(ip, p.remoteIP) == 0 {
					b = true
				}
			}
			if b == false {
				hosts[host].ips = append(hosts[host].ips, p.remoteIP)
			}

			// If the section is not registered, create new
			if _, ok := hosts[host].sections[section]; ok == false {
				// Register new section
				hosts[host].sections[section] = NewSectionStats(section)
			}
		}

		// Update statistics
		a.updateSectionStats(host, section, p.request)
	}
}

// AddPacket adds a packet to the report
func (a *analysis) AddPacket(p *MetaPacket) {
	a.packets = append(a.packets, p)

	a.updateAnalysis(p)
}

func (a *analysis) build() {
	// TODO : finish analysis of the report and build the final thing

	/*
		1. for each domain, count the number of total hits
		2. for the domain with most hits, sort sections per hits
		3. build report
	*/
}

func buildReportMsg(r *analysis) reportMsg {
	// TODO : build a report message from the report
	msg := reportMsg{
		report:    r,
		timestamp: time.Now(),
	}

	return msg
}

// NewAnalysis returns a new and empty Analysis struct
func NewAnalysis() *analysis {
	return &analysis{
		packets:      nil,
		nbHosts:      0,
		hosts:        make(map[string]*hostStats),
		lastSeenHost: nil,
	}
}

func NewReport() *Report {
	return &Report{
		topDomain: "",
		responses: nil,
		sections:  nil,
		analysis:  nil,
	}
}
