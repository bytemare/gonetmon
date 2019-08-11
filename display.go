package main

import (
	log "github.com/sirupsen/logrus"
)

func outputReport(r *Report, output string) {
	// TODO
	log.Info("[i] Display received a report to '%s' : '%s' !\n", output, r)
}

// Display loops on receiving channels to print alerts and reports
func Display(parameters *Parameters, reportChan <-chan *Report, alertChan <-chan alertMsg, syn *Sync) {
	defer syn.wg.Done()

displayLoop:
	for {
		select {

		case <-syn.syncChan:
			break displayLoop

		case alert := <-alertChan:

			if alert.recovery {
				log.Info(alert.body)
			} else {
				log.Warn(alert.body)
			}

		case report := <-reportChan:
			// Interpret report and adapt to desired output
			outputReport(report, parameters.DisplayType)
		}
	}

	log.Info("Display terminating.")
}
