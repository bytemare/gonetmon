package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type session struct {
	report  *Report  // Current report
	watchdog *Watchdog // Surveil traffic behaviour and raise alert is need
	alert   bool     // Current alert status
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
		fmt.Printf("Request reading : EOF\n")
		return nil, err
	}
	if err != nil {
		// TODO :handle error
		fmt.Printf("We have an error in request : %s\n", err)
		return nil, err
	}

	return req, nil
}

func readResponse(b *bufio.Reader) (*http.Response, error) {

	resp, err := http.ReadResponse(b, nil)

	if err == io.EOF {
		// TODO : this should not happen, handle anyway
		fmt.Printf("Response reading : EOF\n")
		return nil, err
	}

	if err != nil {
		// TODO :handle error
		fmt.Printf("We have an error in respinse : %s\n", err)
		return nil, err
	}

	return resp, nil
}

// Transforms the raw payload into a httpPacket struct.
// returns nil wth an error if data does not contain a valid http payload
// TODO : implement fail and error if data is not valid http payload
func dataToHTTP(data dataMsg) (*httpPacket, error) {

	b := []byte(data.payload)

	ioreader := bytes.NewReader(b)

	bufreader := bufio.NewReader(ioreader)

	//r := bufio.NewReader(bytes.NewReader(packet.ApplicationLayer().Payload()))

	packet := &httpPacket{
		messageType: "",
		device:      data.device,
		request:     nil,
		response:    nil,
	}

	request, _ := readRequest(bufreader)

	if request != nil {
		fmt.Printf("We have a request ! => '%s'\n\n", request)
		packet.messageType = httpRequest
		packet.request = request

		return packet, nil
	}

	// If the data is not a request, we might have a response Header

	response, err := readResponse(bufreader)

retry:
	for err != nil {

		fmt.Printf("Response :\n%s\n", data.payload)

		if strings.HasPrefix(data.payload, "HTTP/1.1 ") {
			fmt.Printf("match\n")
			strings.Replace(data.payload, "HTTP/1.1 ", "\n", 1)
			fmt.Printf("New response :\n%s\n", data.payload)
			response, _ := readResponse(bufio.NewReader(bytes.NewReader([]byte(data.payload))))

			if response != nil {
				break retry
			}
		}

		return nil, err
	}

	if response != nil {
		fmt.Printf("We have a response ! => '%s'\n\n", response)
		packet.messageType = httpResponse
		packet.response = response
	}

	return packet, nil
}
