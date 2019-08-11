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
	deviceIP  string // IP address of local network device interface
	remoteIP  string // IP address or remote peer

	// Request information
	request *http.Request

	// Response information
	response *http.Response

	// Associated Captured Packet
	packet		gopacket.Packet
}

type requestStats struct {
	nbReqs		uint				// Sum of all the elements
	nbMethods	map[string]uint	// Map request methods to the number of times they were encountered
}

type responseStats struct {
	nbResp		uint			// Sum of all registered elements
	nbStatus	map[int]uint	// Map status codes to the number of times they were encountered
}

type sectionStats struct {
	section string
	nbHits  int
	requests	requestStats
}

type hostStats struct {
	host      string
	ips		  []string
	sections  map[string]*sectionStats
	// Statistics about responses on that host
	responses	responseStats
}

type analysis struct {
	nbHosts int
	hosts   map[string]*hostStats
	lastSeenHost	*hostStats
}

// Report holds the packets and the result of a recording window
type Report struct {
	packets  []*MetaPacket // A set of packets to be analysed
	analysis analysis      // Final analysis of data
}


// Update statistics of a section with new data
func (r *Report) updateSectionStats(hostname string, sectionName string, p *MetaPacket) {

	// Update Request/Response counters
	if p.messageType == httpRequest {

		host := r.analysis.hosts[hostname]
		section := host.sections[sectionName]

		// Update Hits
		section.nbHits++
		section.requests.nbReqs++

		method := p.request.Method

		// If method was not yet registered, do it
		if _, ok := section.requests.nbMethods[method]; ok == false {
			section.requests.nbMethods[method] = 0
		}
		section.requests.nbMethods[method]++
	}
}

func (r *Report) updateResponseStats(hostname string, p *MetaPacket) {

	resp := r.analysis.hosts[hostname].responses

	resp.nbResp++

	status := p.response.StatusCode

	// If status code has not yet been encountered, add it
	if _, ok := resp.nbStatus[status]; ok == false {
		resp.nbStatus[status] = 0
	}
	resp.nbStatus[status]++
}

// NewSectionStats returns an empty set of statistics about a section
func NewSectionStats(section string) *sectionStats {
	return &sectionStats{
		section: section,
		nbHits:  0,
		requests:  requestStats{
			nbReqs:    0,
			nbMethods: make(map[string]uint),
		},
	}
}

func NewHostStats(host string) *hostStats {
	return &hostStats{
		host:      host,
		ips: []string{},
		sections:  make(map[string]*sectionStats),
		responses: responseStats{
			nbResp:   0,
			nbStatus: make(map[int]uint),
		},
	}
}

func getHost(p *MetaPacket, r *Report) (string, error) {

	// If it's a request, it's in the header
	if p.messageType == httpRequest {
		return p.request.Host, nil
	}

	// Verify if the ip corresponds to the last encountered host
	for _, ip := range r.analysis.lastSeenHost.ips {
		if strings.Compare(ip, p.remoteIP) == 0 {
			return r.analysis.lastSeenHost.host, nil
		}
	}

	// Iterate over all encountered hosts
	for host, stat := range r.analysis.hosts {
		for _, ip := range stat.ips {
			if strings.Compare(ip, p.remoteIP) == 0{
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
func (r *Report) updateAnalysis(p *MetaPacket){

	// If it is a response, we must have seen the corresponding host before, or we cannot work with it
	if p.messageType == httpResponse {
		host, err := getHost(p, r)
		if err != nil {
			log.WithFields(log.Fields{
				"remote IP": p.remoteIP,
			}).Error(err)
			return
		}
		r.updateResponseStats(host, p)
	} else {

		// Here, it is a request
		host, _ := getHost(p, r)
		section, _ := getSection(p)

		hosts := r.analysis.hosts

		// If host not registered, create new
		if _, ok := r.analysis.hosts[host]; ok == false {
			// Register new host and section
			hosts[host] = NewHostStats(host)
			hosts[host].ips = append(hosts[host].ips, p.remoteIP)
			hosts[host].sections[section] = NewSectionStats(section)
		} else {

			// Verify if remote IP was registered for this host
			b := false
			for _, ip := range hosts[host].ips {
				if strings.Compare(ip, p.remoteIP) == 0{
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
		r.updateSectionStats(host, section, p)
	}
}

// AddPacket adds a packet to the report
func (r *Report) AddPacket(p *MetaPacket) {
	r.packets = append(r.packets, p)

	r.updateAnalysis(p)
}

func (r *Report) build() {
	// TODO : finish analysis of the report and build the final thing

	1. for each domain, count the number of total hits
	2. for the domain with most hits, sort sections per hits
	3. build report
}

func buildReportMsg(r *Report) reportMsg {
	// TODO : build a report message from the report
	msg := reportMsg{
		report:    r,
		timestamp: time.Now(),
	}

	return msg
}

// NewReport returns a new and empty Report struct
func NewReport() *Report {
	return &Report{
		packets:  []*MetaPacket{},
		analysis: analysis{
			nbHosts: 0,
			hosts:   make(map[string]*hostStats),
		},
	}
}
