package main

import "time"
import . "github.com/bytemare/gonetmon"

func main() {

	duration := 180 * time.Second

	SnifferTest(duration)
}
