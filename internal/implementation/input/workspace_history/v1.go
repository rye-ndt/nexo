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
	projectDirKey = "nexo.workingdir"
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

func (h *v1) Commit(workflowID, stepID uuid.UUID, projectDir string, excludes []string) error {
	if h.git == "" {
		return custom_error.Bypass("git is not installed, workspace history is disabled")
	}

	projectDir, err := resolveProjectDir(workflowID, projectDir)
	if err != nil {
		return err
	}

	unlock := h.lock(dirKey(projectDir), workflowKey(workflowID))
	defer unlock()

	gitDir := h.gitDir(workflowID)
	if err := h.ensureStore(workflowID, gitDir, projectDir, excludes); err != nil {
		return err
	}

	tag := stepTag(stepID)

	// The baseline names the tree the workflow started from, so it is written once and
	// never moved: a second Run must not rebase "undo everything" onto the agents' work.
	if tag == baselineTag && h.hasTag(gitDir, baselineTag) {
		return nil
	}

	if err := h.dropPreviousAttempt(gitDir, projectDir, tag); err != nil {
		return err
	}

	return h.commitAll(gitDir, projectDir, tag)
}

func (h *v1) commitAll(gitDir, projectDir, tag string) error {
	if _, err := h.run(gitDir, projectDir, "add", "-A"); err != nil {
		return err
	}

	if _, err := h.run(gitDir, projectDir, "commit", "--allow-empty", "--no-verify", "-m", tag); err != nil {
		return err
	}

	_, err := h.run(gitDir, projectDir, "tag", "-f", tag)

	return err
}

func (h *v1) Diff(workflowID, stepID uuid.UUID) ([]*input_itf.FileChange, error) {
	if h.git == "" {
		return nil, custom_error.Bypass("git is not installed, workspace history is disabled")
	}

	unlock := h.lock(workflowKey(workflowID))
	defer unlock()

	gitDir := h.gitDir(workflowID)
	tag := stepTag(stepID)
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

func (h *v1) RestoreTo(workflowID, stepID uuid.UUID, projectDir string) error {
	if h.git == "" {
		return custom_error.Bypass("git is not installed, workspace history is disabled")
	}

	projectDir, err := resolveProjectDir(workflowID, projectDir)
	if err != nil {
		return err
	}

	unlock := h.lock(dirKey(projectDir), workflowKey(workflowID))
	defer unlock()

	gitDir := h.gitDir(workflowID)
	tag := stepTag(stepID)
	if err := h.requireTag(gitDir, tag); err != nil {
		return err
	}

	if err := h.requireSameProjectDir(gitDir, projectDir); err != nil {
		return err
	}

	if err := h.snapshotBeforeRevert(gitDir, projectDir); err != nil {
		return err
	}

	if _, err := h.run(gitDir, projectDir, "reset", "--hard", tag); err != nil {
		return err
	}

	return nil
}

// snapshotBeforeRevert makes the reset that follows it reversible: whatever was written
// since the last step snapshot is otherwise destroyed with no way back.
func (h *v1) snapshotBeforeRevert(gitDir, projectDir string) error {
	return h.commitAll(gitDir, projectDir, preRevertTag)
}

// lock serialises on every key it is given, in the order given, and drops a key once
// nobody holds it. Workflows can share a project folder, so the tree — not the workflow — is
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

func (h *v1) gitDir(workflowID uuid.UUID) string {
	return filepath.Join(h.root, workflowID.String()+".git")
}

// ensureStore only ever creates a store that never existed. A workflow whose store was
// created and is now missing or unreadable is a hard error: re-initialising it would
// turn the next step into a root commit and every later restore into a no-op.
func (h *v1) ensureStore(workflowID uuid.UUID, gitDir, projectDir string, excludes []string) error {
	_, err := os.Stat(filepath.Join(gitDir, "HEAD"))

	switch {
	case err == nil:
		h.rememberStore(workflowID)
	case h.knownStore(workflowID):
		return custom_error.Critical("workspace history of workflow %s vanished from %s: %v", workflowID, gitDir, err)
	default:
		if _, dirErr := os.Stat(gitDir); dirErr == nil {
			return custom_error.Critical("workspace history at %s is unreadable: %v", gitDir, err)
		}

		if err := h.initStore(gitDir, projectDir); err != nil {
			return err
		}

		h.rememberStore(workflowID)
	}

	if err := h.requireSameProjectDir(gitDir, projectDir); err != nil {
		return err
	}

	if err := writeInfoFile(gitDir, "attributes", attributes); err != nil {
		return err
	}

	return writeInfoFile(gitDir, "exclude", excludeBody(excludes))
}

func (h *v1) initStore(gitDir, projectDir string) error {
	if err := os.MkdirAll(h.root, 0o755); err != nil {
		return custom_error.Critical("create workspace history root %s: %v", h.root, err)
	}

	// An empty --role stops the user's init.templateDir hooks from being copied in,
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
		{projectDirKey, projectDir},
	}
	for _, setting := range settings {
		if _, err := h.run(gitDir, "", "config", setting[0], setting[1]); err != nil {
			return err
		}
	}

	return nil
}

func (h *v1) requireSameProjectDir(gitDir, projectDir string) error {
	stored, err := h.run(gitDir, "", "config", "--get", projectDirKey)
	if err != nil {
		_, err = h.run(gitDir, "", "config", projectDirKey, projectDir)

		return err
	}

	if strings.TrimSpace(stored) != projectDir {
		return custom_error.Critical(
			"workspace history at %s snapshotted %s, refusing to use it for %s",
			gitDir, strings.TrimSpace(stored), projectDir,
		)
	}

	return nil
}

func (h *v1) rememberStore(workflowID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.stores[workflowID] = struct{}{}
}

func (h *v1) knownStore(workflowID uuid.UUID) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, found := h.stores[workflowID]

	return found
}

// dropPreviousAttempt keeps one commit per step: when a retry commits straight on
// top of that same step's previous attempt, the previous attempt is unwound first so
// the new snapshot is taken against the state the step started from.
func (h *v1) dropPreviousAttempt(gitDir, projectDir, tag string) error {
	head, found := h.commitish(gitDir, projectDir, "HEAD")
	if !found {
		return nil
	}

	previous, found := h.commitish(gitDir, projectDir, tag)
	if !found || head != previous {
		return nil
	}

	if _, found := h.commitish(gitDir, projectDir, "HEAD~1"); !found {
		return nil
	}

	_, err := h.run(gitDir, projectDir, "reset", "--soft", "HEAD~1")

	return err
}

func (h *v1) commitish(gitDir, projectDir, rev string) (string, bool) {
	out, err := h.run(gitDir, projectDir, "rev-parse", "--quiet", "--verify", rev+"^{commit}")
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

func (h *v1) run(gitDir, projectDir string, args ...string) (string, error) {
	full := make([]string, 0, len(args)+4)
	full = append(full, "--git-dir", gitDir)
	if projectDir != "" {
		full = append(full, "--work-tree", projectDir)
	}
	full = append(full, args...)

	cmd := exec.Command(h.git, full...)
	if projectDir != "" {
		cmd.Dir = projectDir
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

// resolveProjectDir refuses anything git would resolve against its own cwd: a relative
// path becomes --work-tree <dir>/<dir> and snapshots the wrong tree.
func resolveProjectDir(workflowID uuid.UUID, projectDir string) (string, error) {
	if projectDir == "" {
		return "", custom_error.Critical("project folder is empty for workflow %s", workflowID)
	}

	if !filepath.IsAbs(projectDir) {
		return "", custom_error.Critical("project folder %s of workflow %s is not an absolute path", projectDir, workflowID)
	}

	resolved, err := filepath.EvalSymlinks(projectDir)
	if err != nil {
		return "", custom_error.Critical("project folder %s of workflow %s is unreachable: %v", projectDir, workflowID, err)
	}

	return resolved, nil
}

func dirKey(projectDir string) string {
	return "dir:" + projectDir
}

func workflowKey(workflowID uuid.UUID) string {
	return "workflow:" + workflowID.String()
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

func stepTag(stepID uuid.UUID) string {
	if stepID == uuid.Nil {
		return baselineTag
	}

	return "step-" + stepID.String()
}
