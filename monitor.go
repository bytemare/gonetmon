package main

import (
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
	"time"
)

// Monitor is a goroutine that listen on the dataChan channel to pull data packets for analysis
func Monitor(parameters *Parameters, packetChan <-chan packetMsg, reportChan chan<- reportMsg, alertChan chan<- alertMsg, syncChan <-chan struct{}, wg *sync.WaitGroup) {

	// Start a new monitoring session
	session := session{
		report:   NewReport(),
		watchdog: NewWatchdog(parameters.AlertSpan, defaultTick, parameters.AlertThreshold, alertChan, defaultBufSize, syncChan, wg),
		alert:    false,
	}

	// Set up ticker to regularly send reports to display
	tickerReport := time.NewTicker(time.Second * parameters.DisplayRefresh)

monitorLoop:
	for {
		select {

		case <-syncChan:
			log.Info("[i] Monitor received sync message")
			break monitorLoop

		case tr := <-tickerReport.C:
			log.Info("[i] Monitor : time for building and displaying a report :", tr)

			// Build report
			session.report.build()

			// Send report to display
			reportChan <- buildReportMsg(session.report)

			// Reset report
			session.report = NewReport()

		case data := <-packetChan:
			log.Info("[i] Monitor pulled data.")

			// Handle http data type
			if data.dataType == dataHTTP {
				// Transform data into a more convenient form
				packet, err := DataToHTTP(&data)
				if err != nil {
					log.WithFields(log.Fields{
						"interface": data.device,
						"capture timestamp": data.rawPacket.Metadata().Timestamp,
						"payload": strings.Replace(string(data.rawPacket.ApplicationLayer().Payload()), "\n", "{newline}", -1), // Flatten to a single line to avoid breaking log file
					}).Error("Could not interpret package as http.")
					continue
				}

				// Add packet to analysis
				session.report.AddPacket(packet)

				// Update Watchdog
				session.watchdog.AddHit(packet.packet.Metadata().Timestamp)
			}
		}

	}

	wg.Done()
}
