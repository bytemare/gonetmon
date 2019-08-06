// Monitor is a goroutine that supervises the behaviour of data collected by the collector, and informs the operator about specified indicators.
package main

import (
	"fmt"
	"time"
)


/*

Problem :

A probe is an examination of collected data containing traffic to be analysed.
Probes are regularly operated to extrapolate information about traffic.

An alert is raised when a total number of hits of the last t seconds is reached.

Every n seconds, a Report is sent to output to be displayed.

A session is an object that holds the current state of the monitoring session :
- hit queue
- current alert status
- current report accumulator

=================================================







=================================================

We need an accumulator that records the traffic of last t seconds.
=> queue of size threshold / probe-period for a lru cache
=> at every probe, push new value, pop oldest
-> traffic is the sum of all values

-> set as a condition that seconds are the base here, and that threshold must be a multiple of probe-period


=================================================


When reporting every n seconds, the thresold can be reached in a first part within that span,
but fall down to 0 for the remaining time, thus not triggering an alert.

Alert triggering should therefore be designed to detect threshold surpassing within a single collection probe,
since we probe more than once every n seconds.

A report is the result of analysis about probes

*/






func Monitor(parameters *Parameters, displayChan chan struct{}, logsChan chan logs, syncChan chan int) {

	// Start a new monitoring session
	s := NewSession(parameters)

	// Set up Tickers for redundant tasks of analysing and reporting
	tickerProbe := time.NewTicker(time.Second * time.Duration(parameters.ProbePeriod))
	tickerReport := time.NewTicker(time.Second * time.Duration(parameters.DisplayRefresh))

	// Infinite loop
	for {
		// Todo : this method loops continuously even if there's nothing to do between analyses ticks, but is reactive when syncing is needed.
		// Maybe implement waking up when message is received ?

		select {

		// Using the channel in a select with at least one valid or default clause will make it non-blocking
		case sync := <-syncChan:
			// TODO : treat sync
			fmt.Println("[i] Monitor received sync message : %d", sync)
			break

		// If Ticker reached Report time
		case r := <-tickerReport.C:
			{
				fmt.Println("[i] tickerReport")
				fmt.Println("[i] Monitor will report accumulated analyses from last %d seconds ago.", r)
				displayChan <- s.SendReport(&r)
			}

		// If Ticker reached Probing time
		case pt := <-tickerProbe.C:
			{
				fmt.Println("[i] tickerProbe")
				// Get data from Collector
				probe, err := NewProbe(s)
				if err != nil {
					// TODO handle error : could not open file, so we have to retry or stop
				}

				// Decompose raw sniffed lines into fields
				hitset, err := probe.decompose()
				if err != nil {
					// TODO handle error
				}

				// Add new hits to current hitset
				s.addHits(hitset)


				// Check for alert condition
				if s.alertWatcher.alert {
					displayChan <- "Red Alert"
				} else {
					if s.alertWatcher.toggled {
						displayChan <- "Orange Recovery"
					}
				}


			}


		}
	}
}