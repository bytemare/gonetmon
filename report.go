package main

import "time"
import "github.com/jinzhu/copier"

type SectionStats struct {
	domain  string
	section string
	nbHits  int
	nbReqs  int
	nbResp  int
	nbSucc  int
	nbErrs  int
}

type DomainStats struct {
	domain    string
	nbResults int
	stats     []*SectionStats
}

type Analysis struct {
	nbDomains int
	stats     []*DomainStats
}

type Report struct {
	packets  []*httpPacket // A set of packets to be analysed
	analysis Analysis      // Final analysis of data
}

// addPacket adds a packet to the report
func (r *Report) addPacket(p *httpPacket) {
	r.packets = append(r.packets, p)
}

func (r *Report) build() {
	// TODO : finish analysis of the report and build the final thing
}

func buildReportMsg(r *Report) reportMsg {
	// TODO : build a report message from the report
	msg := reportMsg{
		report:    Report{},
		timestamp: time.Now(),
	}

	// TODO : handle error
	_ = copier.Copy(msg.report, r)

	return msg
}

func NewReport() *Report {
	return &Report{
		packets:  nil,
		analysis: Analysis{},
	}
}
