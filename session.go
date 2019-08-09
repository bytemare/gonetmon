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
	watchdog *Watchdog // Surveil traffic behaviour and raise alert is need
	alert    bool      // Current alert status
}

type httpPacket struct {
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
		// TODO : this should not happen, handle anyway
		log.Error("Request reading : EOF\n")
		return nil, err
	}
	if err != nil {
		// TODO :handle error
		log.Error("We have an error in request : %s\n", err)
		return nil, err
	}

	return req, nil
}

func readResponse(b *bufio.Reader) (*http.Response, error) {

	resp, err := http.ReadResponse(b, nil)

	if err == io.EOF {
		// TODO : this should not happen, handle anyway
		log.Error("We have an error in an HTTP response packet : %s\n", err)
		return nil, err
	}

	if err != nil {
		// TODO :handle error
		log.Error("We have an error in an HTTP response packet : %s\n", err)
		return nil, err
	}

	return resp, nil
}

// DataToHTTP transforms the raw payload into a httpPacket struct.
// Returns nil wth an error if data does not contain a valid http payload
// TODO : implement fail and error if data is not valid http payload
func DataToHTTP(data dataMsg) (*httpPacket, error) {

	// In order to use the /net/http functions to interpret http packets,
	// we have to present *bufio.Reader containing the payload
	b := []byte(data.payload)
	bufReader := bufio.NewReader(bytes.NewReader(b))

	packet := &httpPacket{
		messageType: "",
		device:      data.device,
		request:     nil,
		response:    nil,
	}

	// If it is a Response, it starts with 'HTTP/'
	if strings.HasPrefix(data.payload, "HTTP/") {

		response, err := readResponse(bufReader)

		if response != nil {
			//fmt.Printf("We have a response ! => '%s'\n\n", response)
			packet.messageType = httpResponse
			packet.response = response

			return packet, nil
		}

		return nil, err

	}

	// If not, it may be a Request
	request, err := readRequest(bufReader)

	if request != nil {
		//fmt.Printf("We have a request ! => '%s'\n\n", request)
		packet.messageType = httpRequest
		packet.request = request
		return packet, nil
	}

	return nil, err
}