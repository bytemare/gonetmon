package main

import (
	"bufio"
	"bytes"
	"github.com/google/gopacket"
	"io"
	"net/http"
)

func packetToHTTP(packet gopacket.Packet) {

	r := bufio.NewReader(bytes.NewReader(packet.ApplicationLayer().Payload()))

	req, err := http.ReadRequest(r)
	if err == io.EOF {
		// TODO : this should not happen, handle anyway
	}
	if err != nil {
		// TODO :handle error
	}

	resp, err := http.ReadResponse(r, req)
	if err != nil {
		return stream, err
	}
}

