package kmutex

import (
	"sync"
	"testing"
	"time"
)

// Number of unique resources to access
const number = 100

func makeIds(count int) []int {
	ids := make([]int, count)
	for i := 0; i < count; i++ {
		ids[i] = i
	}
	return ids
}

func TestKmutex(t *testing.T) {
	km := New[int]()
	ids := makeIds(number)
	resources := make([]int, number)
	wg := sync.WaitGroup{}

	lc := make(chan int)
	uc := make(chan int)
	// Start 10n goroutines accessing n resources 10 times each
	for i := 0; i < 10*number; i++ {
		wg.Add(1)
		go func(k int) {
			for j := 0; j < 10; j++ {
				lc <- k
				km.Lock(ids[k])
				// read and write resource to check for race
				resources[k] = resources[k] + 1
				km.Unlock(ids[k])
				uc <- k
			}
			wg.Done()
		}(i % len(ids))
	}

	to := time.After(time.Second)
	counts := make(map[int]int)
	var lCount, ulCount int
loop:
	for {
		select {
		case k := <-lc:
			counts[k] = counts[k] + 1
			lCount++
		case k := <-uc:
			counts[k] = counts[k] - 1
			ulCount++
		case <-to:
			t.Fatal("timed out waiting for results")
			break loop
		}
		expectCount := 100 * number
		if lCount == expectCount && ulCount == expectCount {
			// Have all results
			break
		}
	}
	for k, c := range counts {
		if c != 0 {
			t.Errorf("Key %d count != 0: %d\n", k, c)
		}
	}

	wg.Wait()
}

func BenchmarkKmutex1000(b *testing.B) {
	km := New[int]()
	ids := makeIds(number)
	resources := make([]int, number)
	wg := sync.WaitGroup{}

	// Start 1000 goroutines accessing 100 resources N times each
	b.ResetTimer()
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(k int) {
			for j := 0; j < b.N; j++ {
				km.Lock(ids[k])
				// read and write resource to check for race
				resources[k] = resources[k] + 1
				km.Unlock(ids[k])
			}
			wg.Done()
		}(i % len(ids))
	}
	wg.Wait()
}
