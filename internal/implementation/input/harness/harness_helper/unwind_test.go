package harness_helper

import (
	"runtime"
	"testing"

	"hexago/internal/helpers/enums"
)

func TestUnwindRunsStepsNewestFirst(t *testing.T) {
	order := []string{}

	u := &Unwind{}
	u.Push(func() { order = append(order, "first") })
	u.Push(func() { order = append(order, "second") })
	u.Push(func() { order = append(order, "third") })

	u.Run()

	want := []string{"third", "second", "first"}

	if len(order) != len(want) {
		t.Fatalf("order = %v, want %v", order, want)
	}

	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("order = %v, want %v", order, want)
		}
	}
}

func TestUnwindDoneDisarmsEveryStep(t *testing.T) {
	ran := false

	u := &Unwind{}
	u.Push(func() { ran = true })
	u.Done()

	u.Run()

	if ran {
		t.Fatal("a step ran after Done")
	}
}

func TestUnwindRunIsIdempotent(t *testing.T) {
	count := 0

	u := &Unwind{}
	u.Push(func() { count++ })

	u.Run()
	u.Run()

	if count != 1 {
		t.Fatalf("count = %d, want 1", count)
	}
}

func TestPlatformNamesTheMachine(t *testing.T) {
	names := map[string]string{
		enums.Mac.String():     "darwin",
		enums.Linux.String():   "linux",
		enums.Windows.String(): "win32",
	}

	got, err := Platform(names)
	if err != nil {
		t.Fatalf("platform: %v", err)
	}

	arch := map[string]string{"arm64": "arm64", "amd64": "x64"}[runtime.GOARCH]
	want := names[runtime.GOOS] + "-" + arch

	if got != want {
		t.Fatalf("platform = %q, want %q", got, want)
	}
}

func TestPlatformRejectsAnUnknownOS(t *testing.T) {
	if _, err := Platform(map[string]string{"plan9": "plan9"}); err == nil {
		t.Fatal("want an error for an OS the table does not name")
	}
}
