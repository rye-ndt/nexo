package workspace_history

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"

	"hexago/internal/helpers/custom_error"
	input_itf "hexago/internal/interface/input"
)

const (
	baselineTag   = "baseline"
	preRevertTag  = "pre-revert"
	workingDirKey = "nexo.workingdir"
)

// attributes neutralise the work tree's own .gitattributes and the user's global
// filters, so a restore is byte identical rather than eol-normalised or smudged
// through git lfs. `-diff` is deliberately absent: it would turn every patch binary.
const attributes = "* -text -filter -merge\n"

type keyLock struct {
	mu    sync.Mutex
	users int
}

type v1 struct {
	root   string
	git    string
	mu     sync.Mutex
	locks  map[string]*keyLock
	stores map[uuid.UUID]struct{}
}

func InitV1(root string) (input_itf.WorkspaceHistory, error) {
	git, err := exec.LookPath("git")
	if err != nil {
		git = ""
	}

	return &v1{
		root:   root,
		git:    git,
		locks:  map[string]*keyLock{},
		stores: map[uuid.UUID]struct{}{},
	}, nil
}

func (h *v1) Commit(sessionID, taskID uuid.UUID, workingDir string, excludes []string) error {
	if h.git == "" {
		return custom_error.Bypass("git is not installed, workspace history is disabled")
	}

	workingDir, err := resolveWorkingDir(sessionID, workingDir)
	if err != nil {
		return err
	}

	unlock := h.lock(dirKey(workingDir), sessionKey(sessionID))
	defer unlock()

	gitDir := h.gitDir(sessionID)
	if err := h.ensureStore(sessionID, gitDir, workingDir, excludes); err != nil {
		return err
	}

	tag := taskTag(taskID)

	// The baseline names the tree the session started from, so it is written once and
	// never moved: a second Run must not rebase "undo everything" onto the agents' work.
	if tag == baselineTag && h.hasTag(gitDir, baselineTag) {
		return nil
	}

	if err := h.dropPreviousAttempt(gitDir, workingDir, tag); err != nil {
		return err
	}

	return h.commitAll(gitDir, workingDir, tag)
}

func (h *v1) commitAll(gitDir, workingDir, tag string) error {
	if _, err := h.run(gitDir, workingDir, "add", "-A"); err != nil {
		return err
	}

	if _, err := h.run(gitDir, workingDir, "commit", "--allow-empty", "--no-verify", "-m", tag); err != nil {
		return err
	}

	_, err := h.run(gitDir, workingDir, "tag", "-f", tag)

	return err
}

func (h *v1) Diff(sessionID, taskID uuid.UUID) ([]*input_itf.FileChange, error) {
	if h.git == "" {
		return nil, custom_error.Bypass("git is not installed, workspace history is disabled")
	}

	unlock := h.lock(sessionKey(sessionID))
	defer unlock()

	gitDir := h.gitDir(sessionID)
	tag := taskTag(taskID)
	if err := h.requireTag(gitDir, tag); err != nil {
		return nil, err
	}

	nameStatus, err := h.run(gitDir, "", "show", "--format=", "-M", "-z", "--name-status", tag)
	if err != nil {
		return nil, err
	}

	numStat, err := h.run(gitDir, "", "show", "--format=", "-M", "-z", "--numstat", tag)
	if err != nil {
		return nil, err
	}

	patch, err := h.run(gitDir, "", "show", "--format=", "-M", tag)
	if err != nil {
		return nil, err
	}

	changes := parseNameStatus(nameStatus)
	applyNumStat(changes, numStat)
	applyPatches(changes, patch)

	return changes, nil
}

func (h *v1) RestoreTo(sessionID, taskID uuid.UUID, workingDir string) error {
	if h.git == "" {
		return custom_error.Bypass("git is not installed, workspace history is disabled")
	}

	workingDir, err := resolveWorkingDir(sessionID, workingDir)
	if err != nil {
		return err
	}

	unlock := h.lock(dirKey(workingDir), sessionKey(sessionID))
	defer unlock()

	gitDir := h.gitDir(sessionID)
	tag := taskTag(taskID)
	if err := h.requireTag(gitDir, tag); err != nil {
		return err
	}

	if err := h.requireSameWorkingDir(gitDir, workingDir); err != nil {
		return err
	}

	if err := h.snapshotBeforeRevert(gitDir, workingDir); err != nil {
		return err
	}

	if _, err := h.run(gitDir, workingDir, "reset", "--hard", tag); err != nil {
		return err
	}

	return nil
}

// snapshotBeforeRevert makes the reset that follows it reversible: whatever was written
// since the last task snapshot is otherwise destroyed with no way back.
func (h *v1) snapshotBeforeRevert(gitDir, workingDir string) error {
	return h.commitAll(gitDir, workingDir, preRevertTag)
}

// lock serialises on every key it is given, in the order given, and drops a key once
// nobody holds it. Sessions can share a working dir, so the tree — not the session — is
// what has to be held while git writes; callers always take the dir key first.
func (h *v1) lock(keys ...string) func() {
	locks := make([]*keyLock, 0, len(keys))

	h.mu.Lock()
	for _, key := range keys {
		lock, found := h.locks[key]
		if !found {
			lock = &keyLock{}
			h.locks[key] = lock
		}

		lock.users++
		locks = append(locks, lock)
	}
	h.mu.Unlock()

	for _, lock := range locks {
		lock.mu.Lock()
	}

	return func() {
		h.mu.Lock()
		defer h.mu.Unlock()

		for i := len(locks) - 1; i >= 0; i-- {
			locks[i].mu.Unlock()
			locks[i].users--

			if locks[i].users == 0 {
				delete(h.locks, keys[i])
			}
		}
	}
}

func (h *v1) gitDir(sessionID uuid.UUID) string {
	return filepath.Join(h.root, sessionID.String()+".git")
}

// ensureStore only ever creates a store that never existed. A session whose store was
// created and is now missing or unreadable is a hard error: re-initialising it would
// turn the next task into a root commit and every later restore into a no-op.
func (h *v1) ensureStore(sessionID uuid.UUID, gitDir, workingDir string, excludes []string) error {
	_, err := os.Stat(filepath.Join(gitDir, "HEAD"))

	switch {
	case err == nil:
		h.rememberStore(sessionID)
	case h.knownStore(sessionID):
		return custom_error.Critical("workspace history of session %s vanished from %s: %v", sessionID, gitDir, err)
	default:
		if _, dirErr := os.Stat(gitDir); dirErr == nil {
			return custom_error.Critical("workspace history at %s is unreadable: %v", gitDir, err)
		}

		if err := h.initStore(gitDir, workingDir); err != nil {
			return err
		}

		h.rememberStore(sessionID)
	}

	if err := h.requireSameWorkingDir(gitDir, workingDir); err != nil {
		return err
	}

	if err := writeInfoFile(gitDir, "attributes", attributes); err != nil {
		return err
	}

	return writeInfoFile(gitDir, "exclude", excludeBody(excludes))
}

func (h *v1) initStore(gitDir, workingDir string) error {
	if err := os.MkdirAll(h.root, 0o755); err != nil {
		return custom_error.Critical("create workspace history root %s: %v", h.root, err)
	}

	// An empty --template stops the user's init.templateDir hooks from being copied in,
	// and hooksPath keeps a global one out — including from init's own ref writes, which
	// happen before the local config below exists. --no-verify only covers pre-commit.
	hooksPath := filepath.Join(gitDir, "disabled-hooks")

	cmd := exec.Command(h.git, "-c", "core.hooksPath="+hooksPath, "init", "--bare", "--quiet", "--template=", gitDir)
	cmd.Dir = h.root
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return custom_error.Critical("git init %s: %v: %s", gitDir, err, strings.TrimSpace(stderr.String()))
	}

	settings := [][2]string{
		{"core.autocrlf", "false"},
		{"core.hooksPath", hooksPath},
		{"color.ui", "false"},
		{"commit.gpgsign", "false"},
		{"user.name", "nexo"},
		{"user.email", "nexo@localhost"},
		{workingDirKey, workingDir},
	}
	for _, setting := range settings {
		if _, err := h.run(gitDir, "", "config", setting[0], setting[1]); err != nil {
			return err
		}
	}

	return nil
}

func (h *v1) requireSameWorkingDir(gitDir, workingDir string) error {
	stored, err := h.run(gitDir, "", "config", "--get", workingDirKey)
	if err != nil {
		_, err = h.run(gitDir, "", "config", workingDirKey, workingDir)

		return err
	}

	if strings.TrimSpace(stored) != workingDir {
		return custom_error.Critical(
			"workspace history at %s snapshotted %s, refusing to use it for %s",
			gitDir, strings.TrimSpace(stored), workingDir,
		)
	}

	return nil
}

func (h *v1) rememberStore(sessionID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stores[sessionID] = struct{}{}
}

func (h *v1) knownStore(sessionID uuid.UUID) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, found := h.stores[sessionID]

	return found
}

// dropPreviousAttempt keeps one commit per task: when a retry commits straight on
// top of that same task's previous attempt, the previous attempt is unwound first so
// the new snapshot is taken against the state the task started from.
func (h *v1) dropPreviousAttempt(gitDir, workingDir, tag string) error {
	head, found := h.commitish(gitDir, workingDir, "HEAD")
	if !found {
		return nil
	}

	previous, found := h.commitish(gitDir, workingDir, tag)
	if !found || head != previous {
		return nil
	}

	if _, found := h.commitish(gitDir, workingDir, "HEAD~1"); !found {
		return nil
	}

	_, err := h.run(gitDir, workingDir, "reset", "--soft", "HEAD~1")

	return err
}

func (h *v1) commitish(gitDir, workingDir, rev string) (string, bool) {
	out, err := h.run(gitDir, workingDir, "rev-parse", "--quiet", "--verify", rev+"^{commit}")
	if err != nil {
		return "", false
	}

	return strings.TrimSpace(out), true
}

func (h *v1) hasTag(gitDir, tag string) bool {
	_, found := h.commitish(gitDir, "", tag)

	return found
}

func (h *v1) requireTag(gitDir, tag string) error {
	if _, err := os.Stat(filepath.Join(gitDir, "HEAD")); err != nil {
		return custom_error.Critical("no workspace history at %s", gitDir)
	}

	if !h.hasTag(gitDir, tag) {
		return custom_error.Critical("no workspace snapshot named %s", tag)
	}

	return nil
}

func (h *v1) run(gitDir, workingDir string, args ...string) (string, error) {
	full := make([]string, 0, len(args)+4)
	full = append(full, "--git-dir", gitDir)
	if workingDir != "" {
		full = append(full, "--work-tree", workingDir)
	}
	full = append(full, args...)

	cmd := exec.Command(h.git, full...)
	if workingDir != "" {
		cmd.Dir = workingDir
	} else {
		cmd.Dir = gitDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stdout.String(), custom_error.Critical("git %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}

	return stdout.String(), nil
}

// resolveWorkingDir refuses anything git would resolve against its own cwd: a relative
// path becomes --work-tree <dir>/<dir> and snapshots the wrong tree.
func resolveWorkingDir(sessionID uuid.UUID, workingDir string) (string, error) {
	if workingDir == "" {
		return "", custom_error.Critical("working dir is empty for session %s", sessionID)
	}

	if !filepath.IsAbs(workingDir) {
		return "", custom_error.Critical("working dir %s of session %s is not an absolute path", workingDir, sessionID)
	}

	resolved, err := filepath.EvalSymlinks(workingDir)
	if err != nil {
		return "", custom_error.Critical("working dir %s of session %s is unreachable: %v", workingDir, sessionID, err)
	}

	return resolved, nil
}

func dirKey(workingDir string) string {
	return "dir:" + workingDir
}

func sessionKey(sessionID uuid.UUID) string {
	return "session:" + sessionID.String()
}

func excludeBody(excludes []string) string {
	lines := make([]string, 0, len(excludes))
	for _, exclude := range excludes {
		exclude = strings.TrimSpace(exclude)
		if exclude == "" {
			continue
		}
		lines = append(lines, exclude)
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\n") + "\n"
}

func writeInfoFile(gitDir, name, body string) error {
	info := filepath.Join(gitDir, "info")
	if err := os.MkdirAll(info, 0o755); err != nil {
		return custom_error.Critical("create %s: %v", info, err)
	}

	path := filepath.Join(info, name)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return custom_error.Critical("write %s: %v", path, err)
	}

	return nil
}

func taskTag(taskID uuid.UUID) string {
	if taskID == uuid.Nil {
		return baselineTag
	}

	return "task-" + taskID.String()
}
