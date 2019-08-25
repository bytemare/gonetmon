package gonetmon

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	clearConsole  = "\x1Bc"
	topLine       = green + "[gonetmon]" + blue + " Refresh : %d seconds - Alert %d hits / %d seconds. - updated : %s" + stop
	noReport      = "\t\t\t--- No report available : no traffic detected ---"
	reportAlert   = "Alert watchdog :\t %s / %d hits over past %s"
	reportTraffic = "HTTP traffic per interface :  %s"
	reportTop     = "Top host : %s\t - %d hits\t"
	reportResp    = "%s" // OK(%d), Redirect(%d), Server Error(%d), Client Error(%d)"
	reportSection = "\t> %s\t-\t %d hits\t"
	reportReqs    = "%s" //" POST, GET, PUT, PATCH, and DELETE"

	// ANSI Colours
	red   = "\033[31;1;1m"
	green = "\033[32m"
	blue  = "\033[34m"
	stop  = "\033[0m"
)

// buildAlertBarOutput builds the line with the current number of hits over past time frame of alert watching
func buildAlertBarOutput(r *report, p *configuration) string {
	var output string
	hits := strconv.Itoa(r.watchdogHits)
	if r.watchdogHits >= p.alert.threshold {
		hits = red + hits + stop
	}
	output += fmt.Sprintf(reportAlert, hits, p.alert.threshold, p.alert.span)
	return output
}

// buildTrafficOutput builds and returns a string containing the bit rate and total amount of bits per network device
func buildTrafficOutput(r *report, p *configuration) string {
	var output string
	for dev, bits := range r.traffic {
		speed := float64(bits) / p.displayRefresh.Seconds()
		output += fmt.Sprintf("%s : %.2f bits/s (%d bits)   ", dev, speed, bits)
	}
	return output
}

// buildRequestOutput returns a string representation of elements in given map
func buildRequestOutput(methods map[string]uint) string {
	var output string
	for method, nb := range methods {
		output += fmt.Sprintf("%s(%d) ", method, nb)
	}
	return output
}

// buildResponseOutput returns a string representation of elements in given map
func buildResponseOutput(status map[int]uint) string {
	var output string
	for code, nb := range status {
		output += fmt.Sprintf("%d(%d) ", code, nb)
	}
	return output
}

// min returns the minimum between the two values
/*
func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
*/

// displayToConsole builds the final report with passed alerts, clears the terminal and prints the result
func displayToConsole(r *report, alerts *[]string, p *configuration) {
	var output string

	output += fmt.Sprintf(topLine+"\n", int(p.displayRefresh.Seconds()), p.alert.threshold, int(p.alert.span.Seconds()), time.Now().Format("2006-01-02 15:04:05"))
	output += buildAlertBarOutput(r, p) + "\n"
	if r.topHost == nil {
		output += noReport + "\n"
	} else {
		output += fmt.Sprintf(reportTraffic+"\n", buildTrafficOutput(r, p))
		output += fmt.Sprintf(reportTop, r.topHost.host, r.topHost.hits)
		output += fmt.Sprintf(reportResp+"\n", buildResponseOutput(r.topHost.nbStatus))
		//for _, section := range r.sections[:min(p.PacketFilter.NbSections, len(r.sections))] {
		for _, section := range r.sections {
			output += fmt.Sprintf(reportSection, section.section, section.nbHits)
			output += fmt.Sprintf(reportReqs+"\n", buildRequestOutput(section.nbMethods))
		}
	}
	output += strings.Join(*alerts, "")

	fmt.Print(clearConsole)
	fmt.Print(output)
}

// outputReport is a selector between outputs : for now, only console is supported
func outputReport(r *report, alerts *[]string, parameters *configuration) {

	switch parameters.displayType {
	case consoleOutput:
		displayToConsole(r, alerts, parameters)

		// TODO
		/*case fileOutput :
		 */

	}

}

// Display is in charge of rendering a report in to the format of the final output
// For now, only console output is supported
func Display(parameters *configuration, reportChan <-chan *report, alertChan <-chan alertMsg, syn *synchronisation) {
	defer syn.wg.Done()

	var alerts []string

	// Display empty monitoring console
	if parameters.displayType == consoleOutput {
		displayToConsole(&report{
			topHost:   nil,
			sections:  nil,
			timestamp: time.Now(),
		}, &alerts, parameters)
	}

displayLoop:
	for {
		select {

		case <-syn.syncChan:
			break displayLoop

		case alert := <-alertChan:

			if !alert.recovery {
				alert.body = red + alert.body + stop // Red text
			}
			alerts = append(alerts, alert.body+"\n")

			fmt.Println(alert.body)

		case report := <-reportChan:
			// Interpret report and adapt to desired output
			outputReport(report, &alerts, parameters)
		}
	}

	log.Info("Display terminating.")
}
