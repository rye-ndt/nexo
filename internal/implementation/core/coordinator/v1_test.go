package coordinator

import (
	"reflect"
	"testing"

	core_itf "hexago/internal/interface/core"
)

func TestContextExcludesCoversSkeletonTopLevelEntries(t *testing.T) {
	got := contextExcludes(&core_itf.SessionStatus{
		WorkingDirPath: "/work",
		ContextDirPath: "/work",
	})

	want := []string{"/.agent", "/AGENTS.md"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("excludes = %v, want %v", got, want)
	}
}

func TestContextExcludesAreRelativeToTheWorkingDir(t *testing.T) {
	got := contextExcludes(&core_itf.SessionStatus{
		WorkingDirPath: "/work",
		ContextDirPath: "/work/docs/context",
	})

	want := []string{"/docs/context/.agent", "/docs/context/AGENTS.md"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("excludes = %v, want %v", got, want)
	}
}

func TestContextExcludesAreEmptyWhenTheContextDirIsOutsideTheWorkingDir(t *testing.T) {
	got := contextExcludes(&core_itf.SessionStatus{
		WorkingDirPath: "/work",
		ContextDirPath: "/elsewhere",
	})

	if len(got) != 0 {
		t.Fatalf("excludes = %v, want none", got)
	}
}
