// HitQueue implements a FIFO queue containing hits for traffic analysis
package main

import "container/list"

type HitQueue struct {
	q         *Queue
	queue     *list.List
	sum       int  // Current sum of all elements
	threshold int  // threshold for alert
	alert     bool // alert indicator flag
	toggled	  bool // flag indicating if there was a change of state in alert
}

func (h HitQueue) push(hits int) {
	if h.queue.Len() >= h.q.max {
		h.pop()
	}
	// TODO : yes, nasty, should use the queue push, but if we somehow exceed the max + 1 limit, we would loose track of sum if Queue's pop is used
	h.queue.PushBack(hits)
	h.sum += hits

	h.updateAlert()
}

func (h HitQueue) pop() {
	if h.queue.Len() > 0 {
		h.sum -= h.frontValue()
		h.q.pop()
	}
	h.updateAlert()
}

func (h HitQueue) updateAlert() {
	if h.sum >= h.threshold {
		if h.alert == false {
			h.toggled = true
		} else {
			h.toggled = false
		}
		h.alert = true
	} else {
		if h.alert == false {
			h.toggled = true
		} else {
			h.toggled = true
		}
		h.alert = false
	}
}

func (h HitQueue) frontValue() int {
	return h.queue.Front().Value.(int)
}

func NewHitQueue(max int, threshold int) *HitQueue {
	q := NewQueue(max)
	return &HitQueue{
		q:         q,
		queue:     q.queue,
		sum:       0,
		threshold: threshold,
		alert:         false,
	}
}