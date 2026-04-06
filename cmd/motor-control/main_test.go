package main

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCommandGateSerializesConcurrentCalls(t *testing.T) {
	gate := newCommandGate(0)

	var current int32
	var maxConcurrent int32
	var wg sync.WaitGroup

	run := func() error {
		n := atomic.AddInt32(&current, 1)
		defer atomic.AddInt32(&current, -1)

		for {
			m := atomic.LoadInt32(&maxConcurrent)
			if n <= m || atomic.CompareAndSwapInt32(&maxConcurrent, m, n) {
				break
			}
		}

		time.Sleep(5 * time.Millisecond)
		return nil
	}

	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := gate.Run(run); err != nil {
				t.Errorf("run failed: %v", err)
			}
		}()
	}
	wg.Wait()

	if got := atomic.LoadInt32(&maxConcurrent); got != 1 {
		t.Fatalf("expected max concurrency to be 1, got %d", got)
	}
}

func TestCommandGateCooldown(t *testing.T) {
	const cooldown = 15 * time.Millisecond
	gate := newCommandGate(cooldown)

	var runTimes []time.Time
	for i := 0; i < 2; i++ {
		if err := gate.Run(func() error {
			runTimes = append(runTimes, time.Now())
			return nil
		}); err != nil {
			t.Fatalf("run failed: %v", err)
		}
	}

	if len(runTimes) != 2 {
		t.Fatalf("expected two run times, got %d", len(runTimes))
	}

	if delta := runTimes[1].Sub(runTimes[0]); delta < cooldown {
		t.Fatalf("expected cooldown >= %s, got %s", cooldown, delta)
	}
}

func TestCommandGateRunReturnsError(t *testing.T) {
	gate := newCommandGate(0)
	expectedErr := errors.New("boom")

	err := gate.Run(func() error {
		return expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error %v, got %v", expectedErr, err)
	}
}
