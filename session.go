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

// NewMetaPacket returns a new struct initialised with values from the packetMsg
func NewMetaPacket(data *packetMsg) *MetaPacket {
	return &MetaPacket{
		messageType: "",
		device:      data.device,
		deviceIP:    data.deviceIP,
		remoteIP:    data.remoteIP,
		request:     nil,
		response:    nil,
		packet:      data.rawPacket,
	}
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