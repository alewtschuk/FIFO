package main

import (
	"flag"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	MaxC     = 8
	MaxP     = 8
	MaxSleep = 1000000 // in nanoseconds
)

var (
	delay       bool
	pcQueue     *Queue
	numProduced struct {
		num  int
		lock sync.Mutex
	}
	numConsumed struct {
		num  int
		lock sync.Mutex
	}
)

// Produces and enques data
func producer(numItems int, wg *sync.WaitGroup) {
	defer wg.Done()
	// Random num gen with seed
	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)

	// For each index in the numItems
	for i := range numItems {
		if delay {
			sleepNs := rng.Intn(MaxSleep)
			time.Sleep(time.Duration(sleepNs) * time.Nanosecond) // sleep the delay
		}
		item := i
		pcQueue.Enqueue(item) // enqueue item

		// Increment produced count in thread safe manner
		numProduced.lock.Lock()
		numProduced.num++
		numProduced.lock.Unlock()
	}
}

// Consumes and dequeues data
func consumer(wg *sync.WaitGroup) {
	defer wg.Done()
	seed := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(seed)

	for {
		if delay {
			sleepNs := rng.Intn(MaxSleep)
			time.Sleep(time.Duration(sleepNs) * time.Nanosecond) // sleep if delay
		}

		// Removes item from queue
		val := pcQueue.Dequeue()
		if val != nil {
			// Increment if not nil
			numConsumed.lock.Lock()
			numConsumed.num++
			numConsumed.lock.Unlock()
		} else {
			// Nothing to consume and is shutdown
			if !pcQueue.IsShutdown() {
				fmt.Println("ERROR: Null item when queue was not shutdown")
			}
			break
		}
	}
}

func main() {
	// Vars
	var numProducers int
	var numConsumers int
	var numItems int
	var queueSize int

	// Command line flags
	flag.IntVar(&numProducers, "p", 1, "number of producer threads")
	flag.IntVar(&numConsumers, "c", 1, "number of consumer threads")
	flag.IntVar(&numItems, "i", 10, "number of items per producer")
	flag.IntVar(&queueSize, "s", 5, "queue size")
	flag.BoolVar(&delay, "d", false, "introduce random delay")
	flag.Parse()

	// Caps number of producers and consumers
	if numProducers > MaxP {
		numProducers = MaxP
	}
	if numConsumers > MaxC {
		numConsumers = MaxC
	}

	perThread := numItems / numProducers

	fmt.Printf("Simulating %d producers, %d consumers, %d items per producer, queue size %d\n",
		numProducers, numConsumers, perThread, queueSize)

	start := time.Now()

	pcQueue = NewQueue(queueSize)

	var wgProducers sync.WaitGroup
	var wgConsumers sync.WaitGroup

	wgProducers.Add(numProducers)
	wgConsumers.Add(numConsumers)

	for range numProducers {
		go producer(perThread, &wgProducers)
	}

	for range numConsumers {
		go consumer(&wgConsumers)
	}

	// Waits for producers and consumers to sleep
	wgProducers.Wait()
	pcQueue.Shutdown()
	wgConsumers.Wait()

	if numProduced.num != numConsumed.num {
		fmt.Println("ERROR! Produced != Consumed")
		panic("Mismatch")
	}

	fmt.Printf("Queue is empty: %v\n", pcQueue.IsEmpty())
	fmt.Printf("Total produced: %d\n", numProduced.num)
	fmt.Printf("Total consumed: %d\n", numConsumed.num)

	end := time.Now()
	fmt.Printf("%f ms %d items\n", end.Sub(start).Seconds()*1000, numProduced.num)
}
