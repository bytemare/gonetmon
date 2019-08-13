// Display is in charge of rendering a report in to the format of the final output
// For now, only console output is supported
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
	reportAlert   = "Alert Watchdog :\t %s / %d hits over past %s"
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
func buildAlertBarOutput(r *Report, p *Parameters) string {
	var output string
	hits := strconv.Itoa(r.WatchdogHits)
	if r.WatchdogHits >= p.AlertThreshold {
		hits = red + hits + stop
	}
	output += fmt.Sprintf(reportAlert, hits, p.AlertThreshold, p.AlertSpan)
	return output
}

// buildTrafficOutput builds and returns a string containing the bit rate and total amount of bits per network device
func buildTrafficOutput(r *Report, p *Parameters) string {
	var output string
	for dev, bits := range r.traffic {
		speed := float64(bits) / p.DisplayRefresh.Seconds()
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
func displayToConsole(r *Report, alerts *[]string, p *Parameters) {
	var output string

	output += fmt.Sprintf(topLine+"\n", int(p.DisplayRefresh.Seconds()), p.AlertThreshold, int(p.AlertSpan.Seconds()), time.Now().Format("2006-01-02 15:04:05"))
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
func outputReport(r *Report, alerts *[]string, parameters *Parameters) {

	switch parameters.DisplayType {
	case consoleOutput:
		displayToConsole(r, alerts, parameters)

		// TODO
		/*case fileOutput :
		 */

	}

}

// Display loops on receiving channels to print alerts and reports
func Display(parameters *Parameters, reportChan <-chan *Report, alertChan <-chan alertMsg, syn *Sync) {
	defer syn.wg.Done()

	var alerts []string

	// Display empty monitoring console
	if parameters.DisplayType == consoleOutput {
		displayToConsole(&Report{
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
