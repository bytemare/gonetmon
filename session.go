package gonetmon

import (
	"bufio"
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"
)

// session is a placeholder for current analysis and watchdog reference
type session struct {
	analysis *analysis // Current ongoing analysis
	watchdog *watchdog // Surveil traffic behaviour and raise alert if need
}

// NewSession initialises a new monitoring session and launches a watchdog goroutine
func NewSession(alertChan chan<- alertMsg, syn *synchronisation) *session {
	return &session{
		analysis: NewAnalysis(),
		watchdog: NewWatchdog(alertChan, syn),
	}
}

// BuildReport calls for a final analysis and returns the resulting report
func (s *session) BuildReport(watchdogHits int, t time.Time) *report {
	return NewReport(s.analysis, watchdogHits, t)
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

	// If it is a Response, it should start with 'HTTP/'
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
