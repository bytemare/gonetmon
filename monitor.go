package gonetmon

import (
	"github.com/sirupsen/logrus"
	"golang.org/x/tools/go/ssa/interp/testdata/src/strings"
	"time"
)

//Monitor is the link between packet capture, alerting, and display, that accumulates data, analyses it and builds report to display
func Monitor(parameters *Parameters, packetChan <-chan packetMsg, reportChan chan<- *Report, alertChan chan<- alertMsg, syn *Sync) {
	defer syn.wg.Done()

	// Start a new monitoring session
	session := NewSession(parameters, alertChan, syn)

	// Set up ticker to regularly send reports to display
	tickerReport := time.NewTicker(parameters.DisplayRefresh)

monitorLoop:
	for {
		select {

		case <-syn.syncChan:
			log.Info("Monitor received sync message.")
			break monitorLoop

		case tr := <-tickerReport.C:
			log.Info("Preparing report.")

			// Build report and send to display
			reportChan <- session.BuildReport(session.watchdog.Hits(), tr)

			// Renew session analysis
			session.analysis = NewAnalysis()

		case data := <-packetChan:

			// Handle http data type
			if data.dataType == parameters.PacketFilter.Type {
				// Transform data into a more convenient form
				packet, err := DataToHTTP(&data)
				if err != nil {
					log.WithFields(logrus.Fields{
						"interface":         data.device,
						"capture timestamp": data.rawPacket.Metadata().Timestamp,
						"payload":           strings.Replace(string(data.rawPacket.ApplicationLayer().Payload()), "\n", "{newline}", -1), // Flatten to a single line to avoid breaking log file
					}).Error("Could not interpret package as http.")
					continue
				}

				// Add packet to analysis
				session.analysis.AddPacket(packet)

				// Update Watchdog
				session.watchdog.AddHit(packet.packet.Metadata().Timestamp)
			}
		}

	}

	tickerReport.Stop()
	log.Info("Monitor terminating")
}
