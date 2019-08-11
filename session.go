package main

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

// Session is a placeholder for current analysis and report, and Watchdog reference
type Session struct {
	analysis *Analysis // Current ongoing analysis
	watchdog *Watchdog // Surveil traffic behaviour and raise alert if need
}

// NewSession initialises a new monitoring session and launches a Watchdog goroutine
func NewSession(parameters *Parameters, alertChan chan<- alertMsg, syn *Sync) *Session {
	return &Session{
		analysis: NewAnalysis(),
		watchdog: NewWatchdog(parameters, alertChan, syn),
	}
}

// BuildReport calls for a final analysis and collects the resulting report
func (s *Session) BuildReport(t time.Time) *Report {
	return NewReport(s.analysis, t)
}

// readRequest is a wrapper around http.ReadRequest
func readRequest(b *bufio.Reader) (*http.Request, error) {
	req, err := http.ReadRequest(b)
	if err == io.EOF {
		log.Error("HTTP Request reading hit EOF : ", err)
		return nil, err
	}
	if err != nil {
		log.Error("HTTP Request reading error : ", err)
		return nil, err
	}

	return req, nil
}

// readResponse is a wrapper around http.ReadResponse
func readResponse(b *bufio.Reader) (*http.Response, error) {

	resp, err := http.ReadResponse(b, nil)

	if err == io.EOF {
		log.Error("HTTP Response reading hit EOF : ", err)
		return nil, err
	}

	if err != nil {
		log.Error("HTTP Response reading error : ", err)
		return nil, err
	}

	return resp, nil
}

// DataToHTTP transforms the raw payload into a MetaPacket struct.
// Returns nil wth an error if data does not contain a valid http payload
func DataToHTTP(data *packetMsg) (*MetaPacket, error) {

	packet := NewMetaPacket(data)

	appPayload := string(data.rawPacket.ApplicationLayer().Payload())
	// In order to use the /net/http functions to interpret http packets,
	// we have to present *bufio.Reader containing the payload
	b := []byte(appPayload)
	bufReader := bufio.NewReader(bytes.NewReader(b))

	// If it is a Response, it starts with 'HTTP/'
	if strings.HasPrefix(appPayload, "HTTP/") {

		response, err := readResponse(bufReader)

		if err != nil {
			return nil, err
		}

		packet.messageType = httpResponse
		packet.response = response
		return packet, nil
	}

	// If not, it may be a Request
	request, err := readRequest(bufReader)

	if err != nil {
		return nil, err
	}

	packet.messageType = httpRequest
	packet.request = request
	return packet, nil
}
