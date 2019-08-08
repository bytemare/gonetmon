package main

import (
	"fmt"
	"sync"
	"time"
)

// Monitor is a goroutine that listen on the dataChan channel to pull datapackets for analysis
func Monitor(parameters *Parameters, dataChan <-chan dataMsg, reportChan chan<- reportMsg, alertChan chan<- alertMsg, syncChan <-chan struct{}, wg *sync.WaitGroup) {

	// Start a new monitoring session
	report := NewReport()
	session := session{
		report:   report,
		watchdog: NewWatchdog(parameters.AlertSpan, defaultTick, parameters.AlertThreshold, alertChan, defaultBufSize, syncChan, wg),
		alert:    false,
	}

	// Set up ticker to regularly send reports to display
	tickerReport := time.NewTicker(time.Second * time.Duration(parameters.DisplayRefresh))

monitorloop:
	for {
		select {

		case <-syncChan:
			fmt.Println("[i] Monitor received sync message")
			break monitorloop

		case tr := <-tickerReport.C:
			fmt.Println("[i] Monitor : time for building and displaying a report :", tr)

			// Build report
			report.build()

			// Send report to display
			reportChan <- buildReportMsg(report)

			// Reset report
			//report := NewReport()

		case data := <-dataChan:
			fmt.Println("[i] Monitor pulled data.")

			// Handle http data type
			if data.dataType == dataHTTP {

				// Transform data into a more convenient form
				// TODO : handle error
				packet, _ := dataToHTTP(data)
				report.addPacket(packet)

				// Update Watchdog
				session.watchdog.addHit(data.timestamp)
			}
		}

	}
}
