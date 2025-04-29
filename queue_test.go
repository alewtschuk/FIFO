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

func TestDequeueBlocksUntilEnqueue(t *testing.T) {
	q := NewQueue(1)
	done := make(chan int)

	go func() {
		result := q.Dequeue()
		if result == nil {
			t.Error("Expected item, got nil")
		}
		done <- 1
	}()

	time.Sleep(50 * time.Millisecond) // ensure goroutine blocks
	q.Enqueue(42)
	<-done
	q.Shutdown()
}

func TestEnqueueBlocksUntilDequeue(t *testing.T) {
	q := NewQueue(1)
	q.Enqueue(1)

	done := make(chan int)
	go func() {
		q.Enqueue(2)
		done <- 1
	}()

	time.Sleep(50 * time.Millisecond) // ensure goroutine blocks
	q.Dequeue()                       // free up space
	<-done
	q.Shutdown()
}

func TestShutdownUnblocksConsumers(t *testing.T) {
	q := NewQueue(1)
	done := make(chan int)

	go func() {
		item := q.Dequeue()
		if item != nil {
			t.Error("Expected nil after shutdown")
		}
		done <- 1
	}()

	time.Sleep(50 * time.Millisecond)
	q.Shutdown()
	<-done
	q.Shutdown()
}

func TestShutdownWhileBlockedBothEnds(t *testing.T) {
	q := NewQueue(1)
	q.Enqueue(1) // fills queue

	started := make(chan struct{})
	done := make(chan struct{})

	go func() {
		started <- struct{}{}
		q.Enqueue(2) // will block
		done <- struct{}{}
	}()

	go func() {
		started <- struct{}{}
		q.Dequeue() // will unblock Enqueue
		q.Dequeue() // will block (empty)
		done <- struct{}{}
	}()

	<-started
	<-started
	time.Sleep(50 * time.Millisecond) // give both goroutines time to block

	q.Shutdown()

	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Blocked producer/consumer not unblocked by shutdown")
	}
	select {
	case <-done:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("Only one blocked goroutine was released")
	}
	q.Shutdown()
}

func TestStressConcurrentProducersConsumers(t *testing.T) {
	const (
		numProducers = 10
		numConsumers = 10
		itemsPerProd = 1000
	)

	totalItems := numProducers * itemsPerProd
	q := NewQueue(64)

	var prodWg, consWg sync.WaitGroup
	prodWg.Add(numProducers)
	consWg.Add(numConsumers)

	var consumed int64
	var mu sync.Mutex

	// Consumers
	for i := 0; i < numConsumers; i++ {
		go func() {
			for {
				item := q.Dequeue()
				if item == nil {
					break
				}
				mu.Lock()
				consumed++
				mu.Unlock()
			}
			consWg.Done()
		}()
	}

	// Producers
	for i := 0; i < numProducers; i++ {
		go func() {
			for j := 0; j < itemsPerProd; j++ {
				q.Enqueue(j)
			}
			prodWg.Done()
		}()
	}

	prodWg.Wait()
	q.Shutdown()
	consWg.Wait()

	if int(consumed) != totalItems {
		t.Fatalf("Expected %d items consumed, got %d", totalItems, consumed)
	}
	q.Shutdown()
}
