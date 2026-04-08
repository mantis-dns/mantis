package resolver

import (
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedAllows(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)
	if !cb.Allow() {
		t.Error("closed circuit should allow")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(3, time.Second)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitClosed {
		t.Error("should still be closed after 2 failures")
	}

	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Error("should be open after 3 failures")
	}
	if cb.Allow() {
		t.Error("open circuit should reject")
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	cb := NewCircuitBreaker(2, 20*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Fatal("should be open")
	}

	time.Sleep(30 * time.Millisecond)
	if !cb.Allow() {
		t.Error("should allow after cooldown (half-open)")
	}
	if cb.State() != CircuitHalfOpen {
		t.Error("should be half-open")
	}
}

func TestCircuitBreaker_SuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker(2, 20*time.Millisecond)

	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(30 * time.Millisecond)
	cb.Allow() // transitions to half-open

	cb.RecordSuccess()
	if cb.State() != CircuitClosed {
		t.Error("success should close circuit")
	}
	if !cb.Allow() {
		t.Error("closed circuit should allow")
	}
}

func TestCircuitBreaker_FailureReopens(t *testing.T) {
	cb := NewCircuitBreaker(1, 20*time.Millisecond)

	cb.RecordFailure()
	time.Sleep(30 * time.Millisecond)
	cb.Allow() // half-open

	cb.RecordFailure()
	if cb.State() != CircuitOpen {
		t.Error("failure in half-open should reopen")
	}
}
