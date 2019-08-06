package main

import "time"

type dataMsg struct {
	dataType	string
	timestamp	time.Time
	device		string
	body		string
}

const (
	// dataTypes
	dataHTTP	= "http"
)

