package main

import "os"

func main() {
	if Sniff(nil) != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
