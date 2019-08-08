package main

import "fmt"

func outputReport(r *reportMsg, output string) {
	// TODO
	fmt.Printf("[i] Display received a report to '%s' : '%s' !\n", output, reportMsg{})
}

func Display(parameters *Parameters, reportChan <-chan reportMsg, alertChan <-chan alertMsg, syncChan <-chan struct{}) {

displayLoop:
	for {

		select {

		case <-syncChan:
			fmt.Println("[i] Display received sync message")
			break displayLoop

		case alert := <-alertChan:

			fmt.Printf("[ALERT] %s\n", alert.body)

		case report := <-reportChan:

			// Interpret report and adapt to desired output
			outputReport(&report, parameters.DisplayType)
		}
	}
}
