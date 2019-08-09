package main

import "time"

type sectionStats struct {
	domain  string
	section string
	nbHits  int
	nbReqs  int
	nbResp  int
	nbSucc  int
	nbErrs  int
}

type domainStats struct {
	domain    string
	nbResults int
	stats     []*sectionStats
}

type analysis struct {
	nbDomains int
	stats     []*domainStats
}

// Report holds the packets and the result of a recording window
type Report struct {
	packets  []*HTTPPacket // A set of packets to be analysed
	analysis analysis      // Final analysis of data
}

// AddPacket adds a packet to the report
func (r *Report) AddPacket(p *HTTPPacket) {
	r.packets = append(r.packets, p)
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
		packets:  nil,
		analysis: analysis{},
	}
}
