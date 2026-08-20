package helpers

import "path/filepath"

const knowledgeDirSuffix = ".harness/context"

func KnowledgeDir(projectDir string) string {
	return filepath.Join(projectDir, filepath.FromSlash(knowledgeDirSuffix))
}
