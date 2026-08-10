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

// WorkspaceHistory snapshots the session working tree so a task can be reverted to.
// Implementations are lazy: the first Commit for a session creates its store.
type WorkspaceHistory interface {
	Commit(sessionID, taskID uuid.UUID, workingDir string, excludes []string) error
	Diff(sessionID, taskID uuid.UUID) ([]*FileChange, error)
	RestoreTo(sessionID, taskID uuid.UUID, workingDir string) error
}
