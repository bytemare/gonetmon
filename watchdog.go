package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

const (
	defaultTick    = 500 * time.Millisecond
	defaultBufSize = 1000
)

type hitCache struct {

	// Channels to send operations on
	push    chan time.Time
	bufSize uint // size of channels

	// Doubly linked list
	list list.List

	// Number of elements in current list
	size uint
}

// Watchdog struct holds fifo LRU time-based cache and information necessary to watch for traffic spike
type Watchdog struct {

	// Cache to store timely identified hits and time window to keep them
	cache  hitCache
	window time.Duration
	tick   time.Duration

	// Threshold above which an alert will be raised
	threshold uint

	// Channel to send alerts to
	alertChan chan<- alertMsg

	// Current state of alert
	alert bool

	// Synchronisation
	stop <-chan struct{}
	wg   *sync.WaitGroup
}

// Hits returns the current number of elements in the cache
func (w *Watchdog) Hits() int {
	return int(w.cache.size)
}

func buildAlertMsg(w *Watchdog, recovery bool, t time.Time) alertMsg {

	var message string

	if recovery {
		message = fmt.Sprintf(defRecoveryFormat, t.String())
	} else {
		message = fmt.Sprintf(defAlertFormat, w.Hits(), t.String())
	}

	return alertMsg{
		recovery:  recovery,
		body:      message,
		timestamp: time.Time{},
	}
}

// addHit adds an element to the cache by sending a push request to the goroutine
func (w *Watchdog) AddHit(t time.Time) {
	w.cache.push <- t
}

// Verify checks the cache, raising or lowering the alert and sending a message if necessary
func (w *Watchdog) verify() {

	// If the cache is empty, no need to go further
	if w.cache.list.Len() <= 0 {
		// If we were previously in alert, deescalate and send recovery message
		if w.alert {
			w.alert = false
			w.alertChan <- buildAlertMsg(w, true, time.Now())
			return
		}
	}

	// New Alert
	if w.cache.size >= w.threshold && !w.alert {
		w.alert = true
		w.alertChan <- buildAlertMsg(w, false, time.Now())
	} else {
		// Recovery
		if w.alert {
			w.alert = false
			w.alertChan <- buildAlertMsg(w, true, time.Now())
		}
	}

	return
}

// Evict pops all values from the cache that have passed the authorised window
func (w *Watchdog) evict(now time.Time) {
	defer w.verify()

	for {

		if w.cache.list.Len() <= 0 {
			return
		}

		e := w.cache.list.Front()

		// If the element is older than allowed window
		if now.Sub(e.Value.(time.Time)) > w.window {
			w.cache.list.Remove(e)
			w.cache.size--
		} else {
			// Since we store timed values incrementally, following values are all still valid
			break
		}
	}
}

// NewWatchdog returns a watchdog struct and launches a goroutine that will observe its cache to detect alert triggering
func NewWatchdog(window time.Duration, tick time.Duration, threshold uint, c chan<- alertMsg, bufSize uint, stop <-chan struct{}, wg *sync.WaitGroup) *Watchdog {

	wg.Add(1)

	dog := Watchdog{
		cache: hitCache{
			push:    make(chan time.Time, bufSize),
			bufSize: bufSize,
			list:    list.List{},
			size:    0,
		},
		window:    window,
		tick:      tick,
		threshold: threshold,
		alertChan: c,
		alert:     false,
		stop:      stop,
		wg:        wg,
	}

	if tick == 0 {
		dog.tick = defaultTick
	}

	ticker := time.NewTicker(dog.tick)

	// Routine that continuously verifies the cache and will inform about alert status
	go func() {
		for {
			select {

			// Synchronisation/Exit trigger
			case <- stop:
				wg.Done()

			// Continuously evict old elements
			case t := <-ticker.C:
				dog.evict(t)

			// Push request
			case p := <-dog.cache.push:
				dog.cache.list.PushBack(p)
				dog.cache.size++

				dog.verify()
			}
		}
	}()

	return &dog
}
