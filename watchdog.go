package gonetmon

import (
	"container/list"
	"fmt"
	"time"
)

// hitCache is a fifo LRU time-based cache of hits to monitor
type hitCache struct {

	// Channels to send operations on
	push    chan time.Time // Receive new entries through this channel
	bufSize uint           // size of buffered channel

	// Doubly linked list to hold values
	list list.List
}

// watchdog struct holds the cache and information necessary to watch for traffic spike
type watchdog struct {

	// Cache to store timely identified hits and time window to keep them
	cache     hitCache
	timeFrame time.Duration // Monitoring time frame / expiration time
	tick      time.Duration // Interval at which to update cache

	// Threshold above which an alert will be raised
	threshold int

	// Channel to send alerts to
	alertChan chan<- alertMsg

	// Current state of alert
	alert bool

	// Synchronisation
	syn *synchronisation
}

// Hits returns the current number of elements in the cache
func (w *watchdog) Hits() int {
	return w.cache.list.Len()
}

// buildAlertMsg builds an alert message appropriately to the current situation of recovery
func buildAlertMsg(w *watchdog, recovery bool, t time.Time) alertMsg {
	var message string

	if recovery {
		message = fmt.Sprintf(defRecoveryFormat, t.Format(defTimeLayout))
	} else {
		message = fmt.Sprintf(defAlertFormat, w.Hits(), t.Format(defTimeLayout))
	}

	return alertMsg{
		recovery:  recovery,
		body:      message,
		timestamp: time.Time{},
	}
}

// AddHit adds an element to the cache by sending a push request to the goroutine
func (w *watchdog) AddHit(t time.Time) {
	w.cache.push <- t
}

// Verify checks the cache, raising or lowering the alert and sending a message if necessary
func (w *watchdog) verify() {

	// If the cache is empty, no need to go further
	if w.cache.list.Len() <= 0 {
		// If we were previously in alert, deescalate and send recovery message
		if w.alert {
			w.alert = false
			w.alertChan <- buildAlertMsg(w, true, time.Now())
		}
		return
	}

	// Threshold reached
	if w.cache.list.Len() >= w.threshold {
		// New Alert
		if !w.alert {
			w.alert = true
			w.alertChan <- buildAlertMsg(w, false, time.Now())
		}
	} else {
		// Recovery
		if w.alert {
			w.alert = false
			w.alertChan <- buildAlertMsg(w, true, time.Now())
		}
	}
}

// Evict pops all values from the cache that have passed the authorised window
func (w *watchdog) evict(now time.Time) {
	for {

		if w.cache.list.Len() <= 0 {
			break
		}

		e := w.cache.list.Front()

		// If the element is older than allowed window
		if now.Sub(e.Value.(time.Time)) > w.timeFrame {
			w.cache.list.Remove(e)
		} else {
			// Since we store timed values incrementally, following values are all still valid
			break
		}
	}
}

// WatchdogRoutine is an alert monitor that records a timestamp of each packet inside the current time frame.
// The watchdog raises an alert if the number of packets meet a given threshold, and informs if alert has recovered.
// It continuously verifies the cache and will inform about alert status
func WatchdogRoutine(dog *watchdog, syn *synchronisation) {
	defer syn.wg.Done()
	ticker := time.NewTicker(dog.tick)
watchdogLoop:
	for {
		select {

		// Synchronisation/Exit trigger
		case <-syn.syncChan:
			ticker.Stop()
			log.Info("watchdog terminating.")
			break watchdogLoop

		// Continuously evict old elements
		case t := <-ticker.C:
			dog.evict(t)
			dog.verify()

		// Push request
		case p := <-dog.cache.push:
			dog.cache.list.PushBack(p)
			dog.verify()
		}
	}
}

// NewWatchdog returns a watchdog struct and launches a goroutine that will observe its cache to detect alert triggering
func NewWatchdog(c chan<- alertMsg, syn *synchronisation) *watchdog {

	dog := &watchdog{
		cache: hitCache{
			push:    make(chan time.Time, config.alert.watchdogBufSize),
			bufSize: config.alert.watchdogBufSize,
			list:    list.List{},
		},
		timeFrame: config.alert.span,
		tick:      config.alert.watchdogTick,
		threshold: config.alert.threshold,
		alertChan: c,
		alert:     false,
		syn:       syn,
	}

	// Routine that continuously verifies the cache and will inform about alert status
	syn.addRoutine()
	go WatchdogRoutine(dog, syn)

	return dog
}
