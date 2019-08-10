package main

import (
	"github.com/google/gopacket"
	"net/http"
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

	//TODO somehow integrate interfaces in the stats

	// Request information
	request *http.Request

	// Response information
	response *http.Response

	// Associated Captured Packet
	packet		*gopacket.Packet
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
	domain  string
	section string
	nbHits  int

	// Statistics about requests and responses
	requests	requestStats
	responses	responseStats
}

type domainStats struct {
	domain    string
	nbResults int
	sections	map[string]*sectionStats
}

type analysis struct {
	nbDomains int
	domains   map[string]*domainStats
}

// Report holds the packets and the result of a recording window
type Report struct {
	packets  []*MetaPacket // A set of packets to be analysed
	analysis analysis      // Final analysis of data
}

// Update statistics of a section with new data
func updateSectionStat(section *sectionStats, p *MetaPacket) {

	// Update Hits
	section.nbHits++

	// Update Request/Response counters
	if p.messageType == httpRequest {

		section.requests.nbReqs++

		method := p.request.Method

		// If method was not yet registered, do it
		if _, ok := section.requests.nbMethods[method]; ok == false {
			section.requests.nbMethods[method] = 0
		}
		section.requests.nbMethods[method]++

	} else { // Response

		section.responses.nbResp++
		status := p.response.StatusCode

		// If status code has not yet been encountered, add it
		if _, ok := section.responses.nbStatus[status]; ok == false {
			section.responses.nbStatus[status] = 0
		}
		section.responses.nbStatus[status]++
	}
}

// NewSectionStats returns an empty set of statistics about a section
func NewSectionStats(domain string, section string) *sectionStats {
	return &sectionStats{
		domain:  domain,
		section: section,
		nbHits:  0,
		requests:requestStats{
			nbReqs:    0,
			nbMethods: make(map[string]uint),
		},
		responses:responseStats{
			nbResp:   0,
			nbStatus: make(map[int]uint),
		},
	}
}

func NewDomainStats(domain string) *domainStats {
	return &domainStats{
		domain:    domain,
		nbResults: 0,
		sections:  nil,
	}
}

// updateAnalysis update's the report's current analysis with the new incoming packet information
func (r *Report) updateAnalysis(p *MetaPacket){
	domain := getDomain(p)
	section := getSection(p)
	domains := r.analysis.domains

	// If if domain not registered, create new
	if _, ok := r.analysis.domains[domain]; ok == false {
		// Register new domain and section
		domains[domain] = NewDomainStats(domain)
		domains[domain].sections[section] = NewSectionStats(domain, section)
	} else {
		// If the section is not registered, create new
		if _, ok := domains[domain].sections[section]; ok == false {
			// Register new section
			domains[domain].sections[section] = NewSectionStats(domain, section)
		}
	}

	// Update statistics on section
	updateSectionStat(domains[domain].sections[section], p)
}

// AddPacket adds a packet to the report
func (r *Report) AddPacket(p *MetaPacket) {
	r.packets = append(r.packets, p)
	r.updateAnalysis(p)
}

func (r *Report) build() {
	// TODO : finish analysis of the report and build the final thing
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
		analysis: analysis{},
	}
}
