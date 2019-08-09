package main

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

func outputReport(r *reportMsg, output string) {
	// TODO
	log.Info("[i] Display received a report to '%s' : '%s' !\n", output, r)
}

// Display loops on receiving channels to print alerts and reports
func Display(parameters *Parameters, reportChan <-chan reportMsg, alertChan <-chan alertMsg, syncChan <-chan struct{}, wg *sync.WaitGroup) {

displayLoop:
	for {
		select {

		case <-syncChan:
			break displayLoop

		case alert := <-alertChan:

			if alert.recovery {
				log.Info(alert.body)
			} else {
				log.Warn(alert.body)
			}

		case report := <-reportChan:
			// Interpret report and adapt to desired output
			outputReport(&report, parameters.DisplayType)
		}
	}

	log.Info("Display terminating.")
	wg.Done()
}
