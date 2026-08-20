package coordinator

import (
	"reflect"
	"testing"
)

func TestContextExcludesCoversSkeletonTopLevelEntries(t *testing.T) {
	got := contextExcludes("/work")

	want := []string{"/.harness/context/.agent", "/.harness/context/AGENTS.md"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("excludes = %v, want %v", got, want)
	}
}
