package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
)



type WatchDog struct {
	// TODO : implement new aggregator and alarm system that is time based
}


type session struct {
	report		*Report		// Current report
	watcher		WatchDog	// Surveil traffic behaviour and raise alert is need
	alert		bool		// Alert flag
}

type httpPacket struct {
	messageType		string	// Either request or response
	device			string	// Interface on which the packet was recorded

	//TODO somehow integrate interfaces in the stats

	// Request information
	request			*http.Request

	// Response information
	response		*http.Response
}


const (
	httpResponse	= "response"
	httpRequest		= "request"
)

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
		device:		 data.device,
		request:     nil,
		response:    nil,
	}

	req, err := http.ReadRequest(bufreader)
	if err == io.EOF {
		// TODO : this should not happen, handle anyway
		fmt.Printf("Request reading : EOF\n")
		return nil, err
	}
	if err != nil {
		// TODO :handle error
		fmt.Printf("We have an error in request : %s\n", err)
	}

	if req != nil {
		fmt.Printf("We have a request ! => '%s'\n\n", req)
		packet.messageType = httpRequest
		packet.request = req
	}

	resp, err := http.ReadResponse(bufreader, req)
	if err != nil {
		// TODO :handle error
		fmt.Printf("We have an error in respinse : %s\n", err)
		return nil, err
	}

	if resp != nil {
		fmt.Printf("We have a response ! => '%s'\n\n", resp)
		packet.messageType = httpResponse
		packet.response = resp
	}

	return packet, nil
}