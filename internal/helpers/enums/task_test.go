package enums

import "testing"

// The scheduler asks these three questions of every task, so the sets they answer
// from are the one place a status is allowed to mean something.
func TestTaskStatusSets(t *testing.T) {
	cases := []struct {
		status      TaskStatus
		takeable    bool
		retryable   bool
		removable   bool
		cancellable bool
	}{
		{TaskNotTaken, true, false, false, true},
		{TaskProcessing, false, false, false, true},
		{TaskAwaitingAccept, false, false, false, true},
		{TaskCompleted, false, false, true, false},
		{TaskCancelled, true, true, true, false},
		{TaskFailed, true, true, false, false},
	}

	for _, c := range cases {
		t.Run(string(c.status), func(t *testing.T) {
			if got := c.status.Takeable(); got != c.takeable {
				t.Errorf("Takeable() = %v, want %v", got, c.takeable)
			}

			if got := c.status.Retryable(); got != c.retryable {
				t.Errorf("Retryable() = %v, want %v", got, c.retryable)
			}

			if got := c.status.Removable(); got != c.removable {
				t.Errorf("Removable() = %v, want %v", got, c.removable)
			}

			if got := c.status.Cancellable(); got != c.cancellable {
				t.Errorf("Cancellable() = %v, want %v", got, c.cancellable)
			}
		})
	}
}
