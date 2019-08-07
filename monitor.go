package main

import (
	"fmt"
	"time"
)

// Monitor is a goroutine that listen on the dataChan channel to pull datapackets for analysis
func Monitor(parameters *Parameters, dataChan <-chan dataMsg, reportChan chan<- reportMsg, alertChan chan<- alertMsg, syncChan <-chan struct{}){

	// Start a new monitoring session
	report 	:= NewReport()
	session := session{
		report:       report,
		watcher: WatchDog{
			//set:       cache.New(parameters.AlertSpan*time.Second, parameters.AlertSpan*time.Second),
			threshold: parameters.AlertThreshold,
			alert:     false,
			toggled:   false,
		},
		alert: false,
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
				//_ = session.watcher.set.Add(ivalidkey, 1)

				// Check if alert is triggered and report if needed
				if session.watcher.alert || session.watcher.toggled {
					alertChan <- buildAlertMsg(&session.watcher) //"Red Alert" or "Orange Recovery"
				}

			}
		}



	}
}