package main

import (
	"errors"
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
	section string
	nbHits  int
}

type hostStats struct {
	host      string
	sections  map[string]*sectionStats
	// Statistics about requests and responses
	requests	requestStats
	responses	responseStats
}

type analysis struct {
	nbHosts int
	hosts   map[string]*hostStats
}

// Report holds the packets and the result of a recording window
type Report struct {
	packets  []*MetaPacket // A set of packets to be analysed
	analysis analysis      // Final analysis of data
}

/*
// Update statistics of a section with new data
func (r *Report) updateStats(hostname string, sectionName string, p *MetaPacket) {

	host := r.analysis.hosts[hostname]
	section := host.sections[sectionName]

	// Update Hits
	section.nbHits++

	// Update Request/Response counters
	if p.messageType == httpRequest {

		host.requests.nbReqs++
		section.nbHits++
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
}*/

// NewSectionStats returns an empty set of statistics about a section
func NewSectionStats(section string) *sectionStats {
	return &sectionStats{
		section: section,
		nbHits:  0,
	}
}

func NewHostStats(host string) *hostStats {
	return &hostStats{
		host:      host,
		sections:  nil,
	}
}


func getHost(p *MetaPacket) (string, error) {

	if p.messageType == httpRequest {
		return p.request.Host, nil
	}

	return "fAiL.c0m", errors.New("could not extract hostname from packet")
}

func getSection(p *MetaPacket) (string, error) {
	if p.messageType == httpRequest {
		return p.request.RequestURI, nil
	}

	return "dummy section", errors.New("could not extract section from packet")
}


// updateAnalysis update's the report's current analysis with the new incoming packet information
func (r *Report) updateAnalysis(p *MetaPacket){
	host, _ := getHost(p)
	section, _ := getSection(p)
	hosts := r.analysis.hosts

	// If if host not registered, create new
	if _, ok := r.analysis.hosts[host]; ok == false {
		// Register new host and section
		hosts[host] = NewHostStats(host)
		hosts[host].sections[section] = NewSectionStats(section)
	} else {
		// If the section is not registered, create new
		if _, ok := hosts[host].sections[section]; ok == false {
			// Register new section
			hosts[host].sections[section] = NewSectionStats(section)
		}
	}

	// Update statistics on section
	//updateStats(host, section, p)
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
