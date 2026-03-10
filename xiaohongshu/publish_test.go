package xiaohongshu

import (
	"testing"
	"time"
)

type fakeClock struct {
	current time.Time
	sleeps  []time.Duration
}

func newFakeClock() *fakeClock {
	return &fakeClock{current: time.Unix(0, 0)}
}

func (c *fakeClock) now() time.Time {
	return c.current
}

func (c *fakeClock) sleep(d time.Duration) {
	c.sleeps = append(c.sleeps, d)
	c.current = c.current.Add(d)
}

func TestWaitForConditionReturnsAsSoonAsConditionIsMet(t *testing.T) {
	clock := newFakeClock()
	attempts := 0

	ok := waitForCondition(500*time.Millisecond, 100*time.Millisecond, clock.now, clock.sleep, func() bool {
		attempts++
		return attempts == 3
	})

	if !ok {
		t.Fatal("expected condition to be met before timeout")
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
	if len(clock.sleeps) != 2 {
		t.Fatalf("expected 2 sleeps before success, got %d", len(clock.sleeps))
	}
	for i, got := range clock.sleeps {
		if got != 100*time.Millisecond {
			t.Fatalf("sleep %d: expected 100ms, got %s", i, got)
		}
	}
}

func TestWaitForConditionStopsAtTimeoutBoundary(t *testing.T) {
	clock := newFakeClock()
	attempts := 0

	ok := waitForCondition(250*time.Millisecond, 100*time.Millisecond, clock.now, clock.sleep, func() bool {
		attempts++
		return false
	})

	if ok {
		t.Fatal("expected timeout when condition never becomes true")
	}
	if attempts != 4 {
		t.Fatalf("expected 4 attempts, got %d", attempts)
	}
	wantSleeps := []time.Duration{100 * time.Millisecond, 100 * time.Millisecond, 50 * time.Millisecond}
	if len(clock.sleeps) != len(wantSleeps) {
		t.Fatalf("expected %d sleeps, got %d", len(wantSleeps), len(clock.sleeps))
	}
	for i, want := range wantSleeps {
		if clock.sleeps[i] != want {
			t.Fatalf("sleep %d: expected %s, got %s", i, want, clock.sleeps[i])
		}
	}
}
