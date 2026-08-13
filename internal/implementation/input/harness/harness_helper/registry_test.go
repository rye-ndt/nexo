package harness_helper

import (
	"strings"
	"sync"
	"testing"
)

type fakeProc struct {
	name string
}

func TestGetUnknownIDNamesTheID(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 2)

	_, err := r.Get("missing")
	if err == nil {
		t.Fatal("want an error for an unknown id")
	}

	if !strings.Contains(err.Error(), "missing") {
		t.Fatalf("want the id in %q", err.Error())
	}
}

func TestAdmitRefusesPastMax(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 1)

	if err := r.Admit("a", &fakeProc{name: "a"}); err != nil {
		t.Fatalf("first admit: %v", err)
	}

	err := r.Admit("b", &fakeProc{name: "b"})
	if err == nil {
		t.Fatal("want the second admit refused")
	}

	if !strings.Contains(err.Error(), "limit of 1") {
		t.Fatalf("want the limit in %q", err.Error())
	}

	if _, err := r.Get("b"); err == nil {
		t.Fatal("a refused proc must not be stored")
	}

	if r.Count() != 1 {
		t.Fatalf("count = %d, want 1", r.Count())
	}
}

func TestReserveRefusesPastMax(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 1)

	if err := r.Reserve(); err != nil {
		t.Fatalf("reserve on an empty registry: %v", err)
	}

	if err := r.Admit("a", &fakeProc{name: "a"}); err != nil {
		t.Fatalf("admit: %v", err)
	}

	if err := r.Reserve(); err == nil {
		t.Fatal("want reserve refused once full")
	}
}

func TestAdmitRefusedAfterUninstallAndAllowedAgainAfterInstall(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 2)

	r.MarkUninstalled()

	err := r.Admit("a", &fakeProc{name: "a"})
	if err == nil {
		t.Fatal("want admit refused while uninstalled")
	}

	if !strings.Contains(err.Error(), "uninstalled") {
		t.Fatalf("want the reason in %q", err.Error())
	}

	if r.Count() != 0 {
		t.Fatalf("count = %d, want 0", r.Count())
	}

	r.MarkInstalled()

	if err := r.Admit("a", &fakeProc{name: "a"}); err != nil {
		t.Fatalf("admit after reinstall: %v", err)
	}
}

func TestTakeRemovesTheProc(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 2)
	proc := &fakeProc{name: "a"}

	if err := r.Admit("a", proc); err != nil {
		t.Fatalf("admit: %v", err)
	}

	taken, err := r.Take("a")
	if err != nil {
		t.Fatalf("take: %v", err)
	}

	if taken != proc {
		t.Fatal("take returned a different proc")
	}

	if _, err := r.Take("a"); err == nil {
		t.Fatal("want the second take to fail")
	}

	if r.Count() != 0 {
		t.Fatalf("count = %d, want 0", r.Count())
	}
}

func TestForgetUnknownIDIsANoOp(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 2)

	r.Forget("missing")

	if r.Count() != 0 {
		t.Fatalf("count = %d, want 0", r.Count())
	}
}

func TestDrainEmptiesAndReturnsEverything(t *testing.T) {
	r := NewRegistry[*fakeProc]("test harness", 4)

	for _, id := range []string{"a", "b", "c"} {
		if err := r.Admit(id, &fakeProc{name: id}); err != nil {
			t.Fatalf("admit %s: %v", id, err)
		}
	}

	drained := r.Drain()
	if len(drained) != 3 {
		t.Fatalf("drained %d, want 3", len(drained))
	}

	if r.Count() != 0 {
		t.Fatalf("count = %d, want 0", r.Count())
	}

	if got := r.Drain(); len(got) != 0 {
		t.Fatalf("second drain returned %d, want 0", len(got))
	}
}

// The registry is reached from the spawn path, the kill path and the output pumps at
// once, so every method has to hold together under concurrent use.
func TestRegistryNeverExceedsMaxUnderConcurrency(t *testing.T) {
	const max = 8

	r := NewRegistry[*fakeProc]("test harness", max)

	wg := sync.WaitGroup{}

	for i := range 64 {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			id := string(rune('a' + i%26))

			if err := r.Admit(id, &fakeProc{name: id}); err != nil {
				return
			}

			r.Get(id)
			r.Take(id)
		}(i)
	}

	wg.Wait()

	if r.Count() > max {
		t.Fatalf("count = %d, want at most %d", r.Count(), max)
	}
}
