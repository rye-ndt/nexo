package input_itf

import (
	"hexago/internal/helpers/enums"

	"github.com/google/uuid"
)

type FileChange struct {
	Path        string
	OldPath     string
	ChangeType  enums.FileChangeType
	Additions   int
	Deletions   int
	UnifiedDiff string
}

// WorkspaceHistory snapshots the workflow working tree so a step can be reverted to.
// Implementations are lazy: the first Commit for a workflow creates its store.
type WorkspaceHistory interface {
	Commit(workflowID, stepID uuid.UUID, projectDir string, excludes []string) error
	Diff(workflowID, stepID uuid.UUID) ([]*FileChange, error)
	RestoreTo(workflowID, stepID uuid.UUID, projectDir string) error
}
