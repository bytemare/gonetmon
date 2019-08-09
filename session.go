package main

import (
	"bufio"
	"bytes"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
)

type session struct {
	report   *Report   // Current report
	watchdog *Watchdog // Surveil traffic behaviour and raise alert if need
	alert    bool      // Current alert status
}

// HTTPPacket is a wrapper around a /net/http Request or Response struct, with additional information on which interface
// the packet was captured
type HTTPPacket struct {
	messageType string // Either request or response
	device      string // Interface on which the packet was recorded

	//TODO somehow integrate interfaces in the stats

	// Request information
	request *http.Request

	// Response information
	response *http.Response
}

const (
	httpResponse = "response"
	httpRequest  = "request"
)

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

// DataToHTTP transforms the raw payload into a HTTPPacket struct.
// Returns nil wth an error if data does not contain a valid http payload
func DataToHTTP(data dataMsg) (*HTTPPacket, error) {

	// In order to use the /net/http functions to interpret http packets,
	// we have to present *bufio.Reader containing the payload
	b := []byte(data.payload)
	bufReader := bufio.NewReader(bytes.NewReader(b))

	packet := &HTTPPacket{
		messageType: "",
		device:      data.device,
		request:     nil,
		response:    nil,
	}

	// If it is a Response, it starts with 'HTTP/'
	if strings.HasPrefix(data.payload, "HTTP/") {

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
