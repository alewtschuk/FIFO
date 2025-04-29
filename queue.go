package main

import (
	"sync"
)

// Define the queue and all data fields
type Queue struct {
	data     []interface{} // empty interface to hold any data
	capacity int           // max capacity
	head     int           // head node where dequeue will take place
	tail     int           // tail node where data will be enqued
	count    int           // number of nodes in the queue
	shutdown bool
	lock     sync.Mutex
	notFull  *sync.Cond // blocking operation conditon var
	notEmpty *sync.Cond // blocking operation conditon var
}

// Creates the new queue with the specified capacity
func NewQueue(capacity int) *Queue {
	q := &Queue{
		data:     make([]interface{}, capacity),
		capacity: capacity,
	}
	// set up lock conditions
	q.notFull = sync.NewCond(&q.lock)
	q.notEmpty = sync.NewCond(&q.lock)
	return q
}

// Enqueues an item into the queue
func (q *Queue) Enqueue(item interface{}) {
	q.lock.Lock()
	defer q.lock.Unlock()

	// If queue is is full and not shut down wait
	for q.count == q.capacity && !q.shutdown {
		q.notFull.Wait()
	}

	// If shutdown do nothing
	if q.shutdown {
		return
	}

	q.data[q.tail] = item              // enqueue data
	q.tail = (q.tail + 1) % q.capacity // circular buffer
	q.count++
	q.notEmpty.Signal() // wakes one goroutine that is not empty
}

// Dequeues an item in the queue
func (q *Queue) Dequeue() interface{} {
	q.lock.Lock()
	defer q.lock.Unlock()

	// If queue is is full and not shut down wait
	for q.count == 0 && !q.shutdown {
		q.notEmpty.Wait()
	}

	// If the queue is empty return nothing
	if q.count == 0 {
		return nil
	}

	// Gets data at head, sets equal to nil, updates wraparound head index
	item := q.data[q.head]
	q.data[q.head] = nil
	q.head = (q.head + 1) % q.capacity
	q.count--
	q.notFull.Signal() // signals not full, wakes one go routine
	return item
}

// Marks as shutdown, effectively destroys queue
func (q *Queue) Shutdown() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.shutdown = true
	// wakes all goroutines to avoid deadlocks
	q.notFull.Broadcast()
	q.notEmpty.Broadcast()
}

// Checks if the queue is empty
func (q *Queue) IsEmpty() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.count == 0
}

// Checks if the queue is shutdown
func (q *Queue) IsShutdown() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.shutdown
}
