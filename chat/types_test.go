package chat

import (
	"context"
	"sync"
	"testing"
	"time"
)

// ==================== Messages Tests ====================

func TestNewMessages(t *testing.T) {
	m := NewMessages()
	if m == nil {
		t.Fatal("NewMessages returned nil")
	}
	if m.Events == nil {
		t.Fatal("NewMessages returned nil Events slice")
	}
	if len(m.Events) != 0 {
		t.Fatalf("Expected empty Events, got %d events", len(m.Events))
	}
}

func TestMessages_AddEvent_Single(t *testing.T) {
	m := NewMessages()
	event := NewEventNewUserMessage("Hello")

	m.AddEvent(event)

	if len(m.Events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(m.Events))
	}
}

func TestMessages_AddEvent_Multiple(t *testing.T) {
	m := NewMessages()
	event1 := NewEventNewUserMessage("Hello")
	event2 := NewEventNewAssistantMessage("Hi there")
	event3 := NewEventNewToken("test")

	m.AddEvent(event1)
	m.AddEvent(event2)
	m.AddEvent(event3)

	if len(m.Events) != 3 {
		t.Fatalf("Expected 3 events, got %d", len(m.Events))
	}
}

func TestMessages_AddEvent_Concurrent(t *testing.T) {
	m := NewMessages()
	wg := sync.WaitGroup{}
	iterations := 100

	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.AddEvent(NewEventNewToken("token"))
		}(i)
	}

	wg.Wait()

	if len(m.Events) != iterations {
		t.Fatalf("Expected %d events, got %d", iterations, len(m.Events))
	}
}

func TestMessages_Snapshot_Empty(t *testing.T) {
	m := NewMessages()
	snapshot := m.Snapshot()

	if len(snapshot) != 0 {
		t.Fatalf("Expected empty snapshot, got %d events", len(snapshot))
	}
}

func TestMessages_Snapshot_ReturnsCopy(t *testing.T) {
	m := NewMessages()
	originalEvent := NewEventNewUserMessage("Original")
	m.AddEvent(originalEvent)

	// Get snapshot
	snapshot1 := m.Snapshot()

	// Modify original
	m.AddEvent(NewEventNewUserMessage("Modified"))

	// Get another snapshot
	snapshot2 := m.Snapshot()

	// First snapshot should not have the modified event
	if len(snapshot1) != 1 {
		t.Fatalf("Expected snapshot1 to have 1 event, got %d", len(snapshot1))
	}

	// Second snapshot should have both events
	if len(snapshot2) != 2 {
		t.Fatalf("Expected snapshot2 to have 2 events, got %d", len(snapshot2))
	}
}

func TestMessages_Snapshot_ModifyCopyDoesNotAffectOriginal(t *testing.T) {
	m := NewMessages()
	event := NewEventNewUserMessage("Test")
	m.AddEvent(event)

	// Get snapshot
	snapshot := m.Snapshot()

	// Modify snapshot directly (bypass AddEvent to test internal copy)
	// Since we can't modify the slice directly (it's a value type),
	// we verify the original is not affected by adding more events
	m.AddEvent(NewEventNewUserMessage("Another"))

	if len(snapshot) != 1 {
		t.Fatalf("Expected snapshot to still have 1 event, got %d", len(snapshot))
	}

	if len(m.Events) != 2 {
		t.Fatalf("Expected original to have 2 events, got %d", len(m.Events))
	}
}

// ==================== ApproveWaiter Tests ====================

func TestNewApproveWaiter(t *testing.T) {
	w := NewApproveWaiter()
	if w == nil {
		t.Fatal("NewApproveWaiter returned nil")
	}
	if w.verdicts == nil {
		t.Fatal("NewApproveWaiter returned nil verdicts channel")
	}
}

func TestApproveWaiter_Attach(t *testing.T) {
	w := NewApproveWaiter()
	event := NewEventNewToolCall("call-id", "tool-name", `{"arg": "value"}`)

	w.Attach(&event)

	// The approval field is private, but we can verify by calling Resolve
	// and checking that it doesn't panic
	event.Resolve(true)
}

func TestApproveWaiter_Wait_ZeroAmount(t *testing.T) {
	w := NewApproveWaiter()
	ctx := context.Background()

	verdicts := w.Wait(ctx, 0)

	// Channel should be immediately closed
	select {
	case _, ok := <-verdicts:
		if ok {
			t.Fatal("Expected channel to be closed immediately for 0 amount")
		}
	default:
		t.Fatal("Expected channel to be closed immediately for 0 amount")
	}
}

func TestApproveWaiter_Wait_SingleVerdict(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	verdicts := w.Wait(ctx, 1)

	// Resolve a verdict
	w.Resolve(Verdict{Accepted: true})

	// Should receive the verdict
	select {
	case v := <-verdicts:
		if !v.Accepted {
			t.Error("Expected Accepted to be true")
		}
	case <-ctx.Done():
		t.Fatal("Context cancelled unexpectedly")
	}
}

func TestApproveWaiter_Wait_MultipleVerdicts(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	amount := 3
	verdicts := w.Wait(ctx, amount)

	// Resolve multiple verdicts
	for i := 0; i < amount; i++ {
		w.Resolve(Verdict{Accepted: i%2 == 0})
	}

	// Should receive all verdicts
	received := 0
	for v := range verdicts {
		received++
		_ = v // Just drain the channel
	}

	if received != amount {
		t.Fatalf("Expected %d verdicts, got %d", amount, received)
	}
}

func TestApproveWaiter_Wait_ContextCancellation(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())

	verdicts := w.Wait(ctx, 5)

	// Cancel context
	cancel()

	// Channel should be closed
	select {
	case _, ok := <-verdicts:
		if ok {
			t.Error("Expected channel to be closed after context cancellation")
		}
	case <-verdicts:
		// Channel closed
	}
}

func TestApproveWaiter_Resolve_AcceptTrue(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	verdicts := w.Wait(ctx, 1)

	w.Resolve(Verdict{Accepted: true})

	select {
	case v := <-verdicts:
		if !v.Accepted {
			t.Error("Expected Accepted to be true")
		}
	case <-ctx.Done():
		t.Fatal("Context cancelled unexpectedly")
	}
}

func TestApproveWaiter_Resolve_AcceptFalse(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	verdicts := w.Wait(ctx, 1)

	w.Resolve(Verdict{Accepted: false})

	select {
	case v := <-verdicts:
		if v.Accepted {
			t.Error("Expected Accepted to be false")
		}
	case <-ctx.Done():
		t.Fatal("Context cancelled unexpectedly")
	}
}

func TestEventNewToolCall_Resolve_NoApproval(t *testing.T) {
	event := NewEventNewToolCall("call-id", "tool-name", `{}`)

	// Should not panic when called without approval attached
	event.Resolve(true)
}

func TestEventNewToolCall_Resolve_DoubleCall(t *testing.T) {
	w := NewApproveWaiter()
	event := NewEventNewToolCall("call-id", "tool-name", `{}`)

	w.Attach(&event)

	// First resolve
	event.Resolve(true)

	// Second resolve should be ignored (answered is now true)
	// We verify this by checking that only one verdict is received
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	verdicts := w.Wait(ctx, 1)
	<-verdicts

	// Try to resolve again - should not cause issues
	event.Resolve(false)

	// Wait a bit to ensure no second verdict comes through
	select {
	case v, ok := <-verdicts:
		if ok {
			t.Errorf("Expected no second verdict, but got: %+v", v)
		}
	case <-ctx.Done():
		// Context done is fine
	case <-verdicts:
		// Channel closed
	}
}

func TestApproveWaiter_Resolve_Concurrent(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	amount := 10
	verdicts := w.Wait(ctx, amount)

	// Resolve from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < amount; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			w.Resolve(Verdict{Accepted: n%2 == 0})
		}(i)
	}

	// Wait for all goroutines to finish resolving
	wg.Wait()

	// Wait for the Wait goroutine to finish and channel to be closed
	// (it closes after receiving all verdicts)
	received := 0
	for v, ok := <-verdicts; ok; v, ok = <-verdicts {
		received++
		_ = v
	}

	if received != amount {
		t.Fatalf("Expected %d verdicts, got %d", amount, received)
	}
}

// TestApproveWaiter_Wait_CtxDoneDuringCollection tests return when ctx.Done() is triggered
// inside the collection loop (not immediately after Wait)
// This covers: case <-ctx.Done(): return inside the for loop
func TestApproveWaiter_Wait_CtxDoneDuringCollection(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())

	// Channel for communication between sender and reader goroutines
	// sender -> ready to send next
	// reader -> received verdict, can continue
	ready := make(chan struct{})

	amount := 10
	verdicts := w.Wait(ctx, amount)

	var (
		received int
		wg       sync.WaitGroup
	)

	// Goroutine "Reader" - reads verdicts, decides whether to continue or cancel
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case _, ok := <-verdicts:
				if !ok {
					// Channel closed - return from function
					return
				}
				received++

				// If we received 3 verdicts, cancel context (ctx.Done() inside loop)
				if received >= 3 {
					cancel()
					return
				}

				// Otherwise, tell sender it can continue
				select {
				case ready <- struct{}{}:
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Goroutine "Sender" - sends approvals, waits for ready signal
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 5; i++ {
			w.Resolve(Verdict{Accepted: true})

			select {
			case <-ready:
				// Can send next
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for both goroutines
	wg.Wait()

	// Should have received exactly 3 verdicts before context cancellation
	if received != 3 {
		t.Fatalf("Expected exactly 3 verdicts, got %d", received)
	}
}

// TestApproveWaiter_Wait_InnerCtxDone tests ctx.Done() triggered inside the inner select
// This covers: after receiving verdict from a.verdicts, ctx.Done() in inner select causes return
func TestApproveWaiter_Wait_InnerCtxDone(t *testing.T) {
	w := NewApproveWaiter()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resolved := make(chan struct{})
	verdicts := w.Wait(ctx, 1)

	// start blocking resolveSync
	go func() {
		w.resolveSync(Verdict{Accepted: true})
		close(resolved)
	}()

	// wait for resolveSync to complete (worker has read verdict and is now in inner select)
	select {
	case <-resolved:
		// ok
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for resolveSync to complete; worker didn't reach inner select")
	}

	// cancel context -> races with out <- v in the worker's inner select
	cancel()

	// wait for verdicts channel to close (meaning worker returned).
	// Both ok=true (verdict delivered) and ok=false (channel closed without verdict)
	// are valid outcomes — we only assert that the worker does not hang.
	select {
	case <-verdicts:
		// either a verdict arrived or the channel was closed — worker finished either way
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timeout waiting for worker to finish after cancel")
	}
}
