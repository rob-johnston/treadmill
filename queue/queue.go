package queue

import (
	"github.com/rob-johnston/plana/job"
	"sync"
)

// thread safe slice for our job queue
type PendingQueue struct {
	sync.Mutex
	items []*job.Job
}

func NewPendingQueue() *PendingQueue {
	pq := &PendingQueue{
		items: []*job.Job{},
	}
	return pq
}

func (pq *PendingQueue) Append (j *job.Job) {
	pq.Lock()
	defer pq.Unlock()
	pq.items = append(pq.items, j)
}

func (pq * PendingQueue) Length() int {
	return len(pq.items)
}

func (pq *PendingQueue) Iterate() <- chan *job.Job {
	c := make(chan *job.Job)

	go func(){
		pq.Lock()
		defer pq.Unlock()

		var v *job.Job
		for len(pq.items) > 0 {
			v, pq.items = pq.items[0], pq.items[1:]
			c <- v
		}

		for i, v := range pq.items {
			v.QueueIndex = i
			c <- v
		}

		close(c)
	}()

	return c
}