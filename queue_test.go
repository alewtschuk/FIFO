package main

import (
	"sync"
	"testing"
	"time"
)

func TestCreateShutdown(t *testing.T) {
	q := NewQueue(10)
	if q == nil {
		t.Fatal("Queue creation failed")
	}
	q.Shutdown()
}

func TestQueueDequeue(t *testing.T) {
	q := NewQueue(10)
	data := 1
	q.Enqueue(&data)
	result := q.Dequeue()
	if result != &data {
		t.Fatal("Expected to dequeue the same pointer that was enqueued")
	}
	q.Shutdown()
}

func TestQueueDequeueMultiple(t *testing.T) {
	q := NewQueue(10)
	data1, data2, data3 := 1, 2, 3
	q.Enqueue(&data1)
	q.Enqueue(&data2)
	q.Enqueue(&data3)

	if q.Dequeue() != &data1 {
		t.Fatal("Expected data1")
	}
	if q.Dequeue() != &data2 {
		t.Fatal("Expected data2")
	}
	if q.Dequeue() != &data3 {
		t.Fatal("Expected data3")
	}
	q.Shutdown()
}

func TestQueueDequeueShutdown(t *testing.T) {
	q := NewQueue(10)
	data1, data2, data3 := 1, 2, 3
	q.Enqueue(&data1)
	q.Enqueue(&data2)
	q.Enqueue(&data3)

	if q.Dequeue() != &data1 {
		t.Fatal("Expected data1")
	}
	if q.Dequeue() != &data2 {
		t.Fatal("Expected data2")
	}

	q.Shutdown()

	if q.Dequeue() != &data3 {
		t.Fatal("Expected data3")
	}
	if !q.IsShutdown() {
		t.Fatal("Expected queue to be shutdown")
	}
	if !q.IsEmpty() {
		t.Fatal("Expected queue to be empty")
	}
	q.Shutdown()
}

func TestBlockingBehavior(t *testing.T) {
	q := NewQueue(1)
	done := make(chan struct{})

	go func() {
		q.Enqueue(1)
		q.Enqueue(2)
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	q.Dequeue() // allow second enqueue
	<-done
	q.Shutdown()
}

func TestConcurrentEnqueueDequeue(t *testing.T) {
	q := NewQueue(100)
	var wg sync.WaitGroup
	items := 1000

	wg.Add(2)

	go func() {
		for i := 0; i < items; i++ {
			q.Enqueue(i)
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < items; i++ {
			q.Dequeue()
		}
		wg.Done()
	}()

	wg.Wait()
	if !q.IsEmpty() {
		t.Fatal("Expected queue to be empty after concurrent operations")
	}
	q.Shutdown()
}

func TestShutdownUnblocksWaiters(t *testing.T) {
	q := NewQueue(1)
	q.Enqueue(1)

	done := make(chan struct{})
	go func() {
		q.Enqueue(2) // will block
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	q.Shutdown()
	<-done
	q.Shutdown()
}
