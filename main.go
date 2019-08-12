// gonetmon is a network monitoring tool.
// It captures packets on the wire from devices based on given criteria, and displays statistics about traffic.
package main

import "os"

func main() {
	if Sniff(nil) != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
