// Queue implements a FIFO queue with maximum element indicator
// TODO : interfaces serve here as generic type. maybe find something more elegant
// TODO : add logic to ensure that the list never exceeds maximum length (e.g. if max is 10, but we somehow have 12 elements, pop until we reach max)
// TODO : maybe use an interface for HitQueue and Probes ?
package http_sniffer

import (
	"container/list"
)

type Queue struct {
	queue *list.List // Linked list containing hits
	max   int        // Maximum number of authorised elements
}

func (q Queue) push(i interface{}) {
	if q.queue.Len() >= q.max {
		q.pop()
	}
	q.queue.PushBack(i)
}

func (q Queue) pop() {
	if q.queue.Len() > 0 {
		q.queue.Remove(q.queue.Front())
	}
}

func (q Queue) len() int{
	return q.queue.Len()
}

func NewQueue(max int) *Queue {
	return &Queue{queue: list.New(), max:max}
}
