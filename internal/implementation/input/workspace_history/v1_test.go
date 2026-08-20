package workspace_history

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

type fixture struct {
	t          *testing.T
	history    input_itf.WorkspaceHistory
	root       string
	workflow   uuid.UUID
	projectDir string
	excludes   []string
}

func newFixture(t *testing.T, excludes ...string) *fixture {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	root := filepath.Join(t.TempDir(), "workflows")
	projectDir := t.TempDir()

	history, err := InitV1(root)
	if err != nil {
		t.Fatalf("init workspace history: %v", err)
	}

	return &fixture{
		t:          t,
		history:    history,
		root:       root,
		workflow:   uuid.New(),
		projectDir: projectDir,
		excludes:   excludes,
	}
}

func (f *fixture) gitDir() string {
	return filepath.Join(f.root, f.workflow.String()+".git")
}

// shadow talks to the store the way an operator recovering by hand would: the package
// under test is the only writer, so a test that wants to read or use a tag it does not
// expose goes straight to git.
func (f *fixture) shadow(args ...string) string {
	f.t.Helper()

	full := append([]string{"--git-dir", f.gitDir(), "--work-tree", f.projectDir}, args...)

	out, err := exec.Command("git", full...).Output()
	if err != nil {
		f.t.Fatalf("git %v: %v", args, err)
	}

	return string(out)
}

func (f *fixture) resolvedProjectDir() string {
	f.t.Helper()

	resolved, err := filepath.EvalSymlinks(f.projectDir)
	if err != nil {
		f.t.Fatalf("resolve %s: %v", f.projectDir, err)
	}

	return resolved
}

func (f *fixture) writeBytes(name string, body []byte) {
	f.t.Helper()

	if err := os.WriteFile(filepath.Join(f.projectDir, name), body, 0o644); err != nil {
		f.t.Fatalf("write %s: %v", name, err)
	}
}

func (f *fixture) requireBytes(name string, want []byte) {
	f.t.Helper()

	body, err := os.ReadFile(filepath.Join(f.projectDir, name))
	if err != nil {
		f.t.Fatalf("read %s: %v", name, err)
	}
	if string(body) != string(want) {
		f.t.Fatalf("%s = %q, want %q", name, body, want)
	}
}

// globalConfig makes the user's git config part of the test: everything this package
// hardens against arrives that way, and the machine running the suite must not decide
// whether the case is covered.
func globalConfig(t *testing.T, body string) {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "gitconfig")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write global git config: %v", err)
	}

	t.Setenv("GIT_CONFIG_GLOBAL", path)
	t.Setenv("GIT_CONFIG_SYSTEM", filepath.Join(dir, "no-system-config"))
}

func requireCritical(t *testing.T, err error, what string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s should have failed", what)
	}

	severity, ok := err.(custom_error.Severity)
	if !ok || !severity.Critical() {
		t.Fatalf("%s error = %v, want a critical error", what, err)
	}
}

func (f *fixture) write(name, body string) {
	f.t.Helper()

	path := filepath.Join(f.projectDir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		f.t.Fatalf("mkdir for %s: %v", name, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		f.t.Fatalf("write %s: %v", name, err)
	}
}

func (f *fixture) remove(name string) {
	f.t.Helper()

	if err := os.Remove(filepath.Join(f.projectDir, name)); err != nil {
		f.t.Fatalf("remove %s: %v", name, err)
	}
}

func (f *fixture) commit(stepID uuid.UUID) {
	f.t.Helper()

	if err := f.history.Commit(f.workflow, stepID, f.projectDir, f.excludes); err != nil {
		f.t.Fatalf("commit %s: %v", stepID, err)
	}
}

func (f *fixture) restore(stepID uuid.UUID) {
	f.t.Helper()

	if err := f.history.RestoreTo(f.workflow, stepID, f.projectDir); err != nil {
		f.t.Fatalf("restore %s: %v", stepID, err)
	}
}

func (f *fixture) diff(stepID uuid.UUID) []*input_itf.FileChange {
	f.t.Helper()

	changes, err := f.history.Diff(f.workflow, stepID)
	if err != nil {
		f.t.Fatalf("diff %s: %v", stepID, err)
	}

	return changes
}

func (f *fixture) requireContent(name, want string) {
	f.t.Helper()

	f.requireBytes(name, []byte(want))
}

func (f *fixture) requireMissing(name string) {
	f.t.Helper()

	if _, err := os.Stat(filepath.Join(f.projectDir, name)); !os.IsNotExist(err) {
		f.t.Fatalf("%s still exists, want it gone", name)
	}
}

func changeFor(t *testing.T, changes []*input_itf.FileChange, path string) *input_itf.FileChange {
	t.Helper()

	for _, change := range changes {
		if change.Path == path {
			return change
		}
	}

	t.Fatalf("no change for %s in %v", path, paths(changes))
	return nil
}

func paths(changes []*input_itf.FileChange) []string {
	out := make([]string, 0, len(changes))
	for _, change := range changes {
		out = append(out, change.Path)
	}

	return out
}

func TestDiffShowsOnlyTheStepsOwnChanges(t *testing.T) {
	f := newFixture(t)

	f.write("shared.txt", "baseline\n")
	f.write("untouched.txt", "same\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("shared.txt", "from a\n")
	f.write("from-a.txt", "a\n")
	f.commit(stepA)

	stepB := uuid.New()
	f.write("shared.txt", "from b\n")
	f.write("from-b.txt", "b\n")
	f.commit(stepB)

	changes := f.diff(stepB)
	if len(changes) != 2 {
		t.Fatalf("step b changed %v, want shared.txt and from-b.txt", paths(changes))
	}

	shared := changeFor(t, changes, "shared.txt")
	if shared.ChangeType != enums.FileModified {
		t.Fatalf("shared.txt type = %s, want %s", shared.ChangeType, enums.FileModified)
	}
	if shared.Additions != 1 || shared.Deletions != 1 {
		t.Fatalf("shared.txt +%d -%d, want +1 -1", shared.Additions, shared.Deletions)
	}
	if !contains(shared.UnifiedDiff, "-from a") || !contains(shared.UnifiedDiff, "+from b") {
		t.Fatalf("shared.txt patch does not show the change: %q", shared.UnifiedDiff)
	}

	added := changeFor(t, changes, "from-b.txt")
	if added.ChangeType != enums.FileAdded {
		t.Fatalf("from-b.txt type = %s, want %s", added.ChangeType, enums.FileAdded)
	}
	if added.Additions != 1 || added.Deletions != 0 {
		t.Fatalf("from-b.txt +%d -%d, want +1 -0", added.Additions, added.Deletions)
	}

	baseline := f.diff(uuid.Nil)
	if len(baseline) != 2 {
		t.Fatalf("baseline changed %v, want the two seeded files", paths(baseline))
	}
}

func TestRestoreToEarlierStepRewindsCreatedModifiedAndDeletedFiles(t *testing.T) {
	f := newFixture(t)

	f.write("kept.txt", "baseline\n")
	f.write("deleted-by-b.txt", "still here\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("kept.txt", "from a\n")
	f.commit(stepA)

	stepB := uuid.New()
	f.write("kept.txt", "from b\n")
	f.write("created-by-b.txt", "new\n")
	f.remove("deleted-by-b.txt")
	f.commit(stepB)

	f.restore(stepA)

	f.requireContent("kept.txt", "from a\n")
	f.requireMissing("created-by-b.txt")
	f.requireContent("deleted-by-b.txt", "still here\n")
}

func TestExcludedPathsAreNeitherCommittedNorRestored(t *testing.T) {
	f := newFixture(t, "AGENTS.md", ".agent/")

	f.write("AGENTS.md", "knowledge\n")
	f.write(".agent/notes.md", "notes\n")
	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("AGENTS.md", "knowledge v2\n")
	f.write(".agent/notes.md", "notes v2\n")
	f.write(".agent/learned.md", "learned later\n")
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	changes := f.diff(stepA)
	if len(changes) != 1 || changes[0].Path != "code.txt" {
		t.Fatalf("step a changed %v, want only code.txt", paths(changes))
	}

	f.restore(uuid.Nil)

	f.requireContent("code.txt", "baseline\n")
	f.requireContent("AGENTS.md", "knowledge v2\n")
	f.requireContent(".agent/notes.md", "notes v2\n")
	f.requireContent(".agent/learned.md", "learned later\n")
}

func TestStepThatChangedNothingStillGetsARestorePoint(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	quiet := uuid.New()
	f.commit(quiet)

	if changes := f.diff(quiet); len(changes) != 0 {
		t.Fatalf("quiet step changed %v, want nothing", paths(changes))
	}

	loud := uuid.New()
	f.write("code.txt", "from loud\n")
	f.write("loud.txt", "loud\n")
	f.commit(loud)

	f.restore(quiet)

	f.requireContent("code.txt", "baseline\n")
	f.requireMissing("loud.txt")
}

func TestRetryMovesTheTagAndDiffFollowsTheNewestAttempt(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "first attempt\n")
	f.write("first-only.txt", "scratch\n")
	f.commit(stepA)

	f.write("code.txt", "second attempt\n")
	f.remove("first-only.txt")
	f.commit(stepA)

	changes := f.diff(stepA)
	if len(changes) != 1 || changes[0].Path != "code.txt" {
		t.Fatalf("retried step a changed %v, want only code.txt", paths(changes))
	}
	if !contains(changes[0].UnifiedDiff, "+second attempt") {
		t.Fatalf("retry patch does not show the newest attempt: %q", changes[0].UnifiedDiff)
	}
	if !contains(changes[0].UnifiedDiff, "-baseline") {
		t.Fatalf("retry patch is not taken against the baseline: %q", changes[0].UnifiedDiff)
	}

	stepB := uuid.New()
	f.write("code.txt", "from b\n")
	f.commit(stepB)

	f.restore(stepA)
	f.requireContent("code.txt", "second attempt\n")
}

func TestStepRerunAfterRestoreDiffsAgainstTheRestoredTree(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	stepB := uuid.New()
	f.write("code.txt", "from b\n")
	f.write("from-b.txt", "b\n")
	f.commit(stepB)

	f.restore(stepA)

	f.write("code.txt", "from b again\n")
	f.commit(stepB)

	changes := f.diff(stepB)
	if len(changes) != 1 || changes[0].Path != "code.txt" {
		t.Fatalf("re-run step b changed %v, want only code.txt", paths(changes))
	}
	if !contains(changes[0].UnifiedDiff, "-from a") || !contains(changes[0].UnifiedDiff, "+from b again") {
		t.Fatalf("re-run patch is not taken against the restored tree: %q", changes[0].UnifiedDiff)
	}

	f.restore(stepA)
	f.requireContent("code.txt", "from a\n")
	f.requireMissing("from-b.txt")
}

func TestRenameIsReportedWithItsOldPath(t *testing.T) {
	f := newFixture(t)

	f.write("old.txt", longBody())
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("new.txt", longBody())
	f.remove("old.txt")
	f.commit(stepA)

	changes := f.diff(stepA)
	if len(changes) != 1 {
		t.Fatalf("step a changed %v, want a single rename", paths(changes))
	}
	if changes[0].ChangeType != enums.FileRenamed {
		t.Fatalf("type = %s, want %s", changes[0].ChangeType, enums.FileRenamed)
	}
	if changes[0].Path != "new.txt" || changes[0].OldPath != "old.txt" {
		t.Fatalf("rename = %s -> %s, want old.txt -> new.txt", changes[0].OldPath, changes[0].Path)
	}
}

func TestBinaryChangeDoesNotBreakTheDiff(t *testing.T) {
	f := newFixture(t)

	binary := make([]byte, 256)
	for i := range binary {
		binary[i] = byte(i)
	}
	if err := os.WriteFile(filepath.Join(f.projectDir, "blob.bin"), binary, 0o644); err != nil {
		t.Fatalf("write blob.bin: %v", err)
	}
	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	for i := range binary {
		binary[i] = byte(255 - i)
	}
	if err := os.WriteFile(filepath.Join(f.projectDir, "blob.bin"), binary, 0o644); err != nil {
		t.Fatalf("rewrite blob.bin: %v", err)
	}
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	changes := f.diff(stepA)
	if len(changes) != 2 {
		t.Fatalf("step a changed %v, want blob.bin and code.txt", paths(changes))
	}

	blob := changeFor(t, changes, "blob.bin")
	if blob.Additions != 0 || blob.Deletions != 0 {
		t.Fatalf("blob.bin +%d -%d, want +0 -0", blob.Additions, blob.Deletions)
	}
	if !contains(blob.UnifiedDiff, "Binary files") {
		t.Fatalf("blob.bin patch = %q, want the binary notice", blob.UnifiedDiff)
	}

	code := changeFor(t, changes, "code.txt")
	if code.Additions != 1 || code.Deletions != 1 {
		t.Fatalf("code.txt +%d -%d, want +1 -1", code.Additions, code.Deletions)
	}
}

func TestDiffAndRestoreRejectUnknownSnapshots(t *testing.T) {
	f := newFixture(t)

	if _, err := f.history.Diff(f.workflow, uuid.New()); err == nil {
		t.Fatal("diff on a workflow with no history should fail")
	}

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	if _, err := f.history.Diff(f.workflow, uuid.New()); err == nil {
		t.Fatal("diff on an unknown step should fail")
	}
	if err := f.history.RestoreTo(f.workflow, uuid.New(), f.projectDir); err == nil {
		t.Fatal("restore to an unknown step should fail")
	}
}

// userGit runs git inside the working dir, which is where the user's own repositories
// live: the whole point of the shadow store is that none of them ever move.
func userGit(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v in %s: %v", args, dir, err)
	}

	return string(out)
}

func repoRefs(t *testing.T, dir string) string {
	t.Helper()

	return userGit(t, dir, "rev-parse", "HEAD") +
		userGit(t, dir, "for-each-ref", "--format=%(refname) %(objectname)")
}

func repoState(t *testing.T, dir string) string {
	t.Helper()

	return repoRefs(t, dir) + userGit(t, dir, "status", "--porcelain")
}

func TestUserRepositoriesInTheProjectDirAreNeverTouched(t *testing.T) {
	f := newFixture(t)

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = f.projectDir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}

	run("init", "--quiet")
	run("config", "user.name", "tester")
	run("config", "user.email", "tester@localhost")
	f.write("code.txt", "baseline\n")
	f.write(".gitignore", "ignored/\n")
	f.write("ignored/blob.txt", "not snapshotted\n")
	run("add", "-A")
	run("commit", "--quiet", "-m", "user commit")
	run("branch", "side")
	run("tag", "v1")

	nested := filepath.Join(f.projectDir, "vendor", "lib")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create nested repo dir: %v", err)
	}

	userGit(t, nested, "init", "--quiet")
	userGit(t, nested, "config", "user.name", "vendor")
	userGit(t, nested, "config", "user.email", "vendor@localhost")
	f.write("vendor/lib/lib.txt", "vendored\n")
	userGit(t, nested, "add", "-A")
	userGit(t, nested, "commit", "--quiet", "-m", "vendor commit")

	outerBefore, nestedBefore := repoState(t, f.projectDir), repoRefs(t, nested)

	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "from a\n")
	f.write("ignored/blob.txt", "changed by a\n")
	f.write("vendor/lib/lib.txt", "changed by a\n")
	f.commit(stepA)

	changes := f.diff(stepA)
	if len(changes) != 1 || changes[0].Path != "code.txt" {
		t.Fatalf("step a changed %v, want only code.txt", paths(changes))
	}

	f.restore(uuid.Nil)

	f.requireContent("code.txt", "baseline\n")
	f.requireContent("ignored/blob.txt", "changed by a\n")

	if got := repoState(t, f.projectDir); got != outerBefore {
		t.Fatalf("the user repository moved:\n%s\nwant\n%s", got, outerBefore)
	}

	// A nested repository enters the snapshot as a gitlink, so its refs must not move and
	// its contents are neither snapshotted nor rolled back. Known limitation, pinned here.
	if got := repoRefs(t, nested); got != nestedBefore {
		t.Fatalf("the nested repository moved:\n%s\nwant\n%s", got, nestedBefore)
	}
	f.requireContent("vendor/lib/lib.txt", "changed by a\n")
}

func TestInitV1SucceedsBeforeAnythingExists(t *testing.T) {
	root := filepath.Join(t.TempDir(), "not", "created", "yet")

	history, err := InitV1(root)
	if err != nil {
		t.Fatalf("init workspace history: %v", err)
	}
	if history == nil {
		t.Fatal("init workspace history returned no history")
	}

	if _, err := os.Stat(root); !os.IsNotExist(err) {
		t.Fatalf("init created %s before the first commit, stat err = %v", root, err)
	}

}

func TestMissingGitDegradesToBypass(t *testing.T) {
	t.Setenv("PATH", "")

	history, err := InitV1(t.TempDir())
	if err != nil {
		t.Fatalf("init workspace history without git: %v", err)
	}

	workflow := uuid.New()
	if err := history.Commit(workflow, uuid.Nil, t.TempDir(), nil); err == nil {
		t.Fatal("commit without git should fail")
	} else if severity, ok := err.(custom_error.Severity); !ok || severity.Critical() {
		t.Fatalf("commit error = %v, want a bypass", err)
	}

	if _, err := history.Diff(workflow, uuid.Nil); err == nil {
		t.Fatal("diff without git should fail")
	} else if severity, ok := err.(custom_error.Severity); !ok || severity.Critical() {
		t.Fatalf("diff error = %v, want a bypass", err)
	}

	if err := history.RestoreTo(workflow, uuid.Nil, t.TempDir()); err == nil {
		t.Fatal("restore without git should fail")
	} else if severity, ok := err.(custom_error.Severity); !ok || severity.Critical() {
		t.Fatalf("restore error = %v, want a bypass", err)
	}
}

func TestBaselineIsWrittenOncePerWorkflow(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "PRECIOUS ORIGINAL\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	stepB := uuid.New()
	f.write("code.txt", "from b\n")
	f.commit(stepB)

	f.restore(stepB)

	// Every Run re-commits the baseline; the second one must not move it on top of the
	// work the agents already did.
	f.commit(uuid.Nil)

	f.restore(uuid.Nil)
	f.requireContent("code.txt", "PRECIOUS ORIGINAL\n")

	changes := f.diff(stepA)
	if len(changes) != 1 || !contains(changes[0].UnifiedDiff, "-PRECIOUS ORIGINAL") {
		t.Fatalf("step a no longer diffs against the original tree: %q", changes[0].UnifiedDiff)
	}
}

func TestRestoreKeepsUnsnapshottedWorkOnThePreRevertTag(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	f.write("code.txt", "hand written and never snapshotted\n")
	f.write("half-done.txt", "the killed step got this far\n")

	f.restore(uuid.Nil)

	f.requireContent("code.txt", "baseline\n")

	if body := f.shadow("show", preRevertTag+":code.txt"); body != "hand written and never snapshotted\n" {
		t.Fatalf("pre-revert snapshot of code.txt = %q, want the unsnapshotted edit", body)
	}

	f.shadow("reset", "--hard", preRevertTag)

	f.requireContent("code.txt", "hand written and never snapshotted\n")
	f.requireContent("half-done.txt", "the killed step got this far\n")
}

func TestRestoreIsByteIdenticalUnderWorkTreeAttributesAndGlobalFilters(t *testing.T) {
	globalConfig(t, "[filter \"mangle\"]\n\tsmudge = sed s/ORIGINAL/MANGLED/\n\tclean = cat\n")

	f := newFixture(t)

	original := []byte("ORIGINAL\r\nBYTES\r\n")

	f.write(".gitattributes", "* text=auto\n*.txt filter=mangle\n")
	f.writeBytes("code.txt", original)
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.writeBytes("code.txt", []byte("REWRITTEN BY THE AGENT\n"))
	f.commit(stepA)

	f.restore(uuid.Nil)

	f.requireBytes("code.txt", original)
}

func TestGlobalColorConfigDoesNotSwallowPatches(t *testing.T) {
	globalConfig(t, "[color]\n\tui = always\n")

	f := newFixture(t)

	f.write("first.txt", "first baseline\n")
	f.write("second.txt", "second baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("first.txt", "first from a\n")
	f.write("second.txt", "second from a\n")
	f.commit(stepA)

	changes := f.diff(stepA)
	if len(changes) != 2 {
		t.Fatalf("step a changed %v, want both files", paths(changes))
	}

	for _, change := range changes {
		if !contains(change.UnifiedDiff, "+++ b/"+change.Path) {
			t.Fatalf("%s carries the wrong patch: %q", change.Path, change.UnifiedDiff)
		}
		if contains(change.UnifiedDiff, "\x1b[") {
			t.Fatalf("%s patch holds terminal escapes: %q", change.Path, change.UnifiedDiff)
		}
	}
}

func TestGlobalHooksNeverRunAgainstTheShadowStore(t *testing.T) {
	f := newFixture(t)

	role := filepath.Join(t.TempDir(), "role")
	hooks := filepath.Join(role, "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatalf("create role hooks: %v", err)
	}

	marker := filepath.Join(t.TempDir(), "hook-ran")
	for _, name := range []string{"post-commit", "pre-commit", "reference-transaction"} {
		script := "#!/bin/sh\necho ran > " + marker + "\n"
		if err := os.WriteFile(filepath.Join(hooks, name), []byte(script), 0o755); err != nil {
			t.Fatalf("write %s hook: %v", name, err)
		}
	}

	globalConfig(t, "[init]\n\troleDir = "+role+"\n[core]\n\thooksPath = "+hooks+"\n")

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatalf("a hook from the user's git config ran against the shadow store, stat err = %v", err)
	}

	if _, err := os.Stat(filepath.Join(f.gitDir(), "hooks", "post-commit")); !os.IsNotExist(err) {
		t.Fatalf("the user's init role was copied into the shadow store, stat err = %v", err)
	}
}

func TestVanishedStoreFailsInsteadOfSilentlyStartingOver(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	stepA := uuid.New()
	f.write("code.txt", "from a\n")
	f.commit(stepA)

	if err := os.RemoveAll(f.gitDir()); err != nil {
		t.Fatalf("remove the store: %v", err)
	}

	f.write("code.txt", "from b\n")
	requireCritical(t, f.history.Commit(f.workflow, uuid.New(), f.projectDir, f.excludes), "commit after the store was deleted")
}

func TestUnreadableStoreFailsInsteadOfSilentlyStartingOver(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	if err := os.Remove(filepath.Join(f.gitDir(), "HEAD")); err != nil {
		t.Fatalf("break the store: %v", err)
	}

	// A restarted app has no memory of the workflow, so the damage has to be visible on
	// disk rather than only in this process.
	restarted, err := InitV1(f.root)
	if err != nil {
		t.Fatalf("init workspace history: %v", err)
	}

	requireCritical(t, restarted.Commit(f.workflow, uuid.New(), f.projectDir, f.excludes), "commit against a broken store")
}

func TestWorkflowsSharingAProjectDirDoNotWriteAtTheSameTime(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	history := f.history.(*v1)
	release := history.lock(dirKey(f.resolvedProjectDir()))

	done := make(chan error, 1)
	go func() {
		done <- f.history.Commit(uuid.New(), uuid.Nil, f.projectDir, nil)
	}()

	select {
	case err := <-done:
		release()
		t.Fatalf("another workflow committed while this working dir was held: %v", err)
	case <-time.After(250 * time.Millisecond):
	}

	release()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("commit of the second workflow: %v", err)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the second workflow never got the working dir")
	}

	if _, err := f.history.Diff(f.workflow, uuid.Nil); err != nil {
		t.Fatalf("diff after both workflows: %v", err)
	}

	history.mu.Lock()
	held := len(history.locks)
	history.mu.Unlock()

	if held != 0 {
		t.Fatalf("%d locks were kept after every call returned, want none", held)
	}
}

func TestRelativeProjectDirIsRejected(t *testing.T) {
	f := newFixture(t)

	f.write("nested/nested/inner.txt", "the wrong tree\n")
	f.write("nested/outer.txt", "the tree that was meant\n")

	t.Chdir(f.projectDir)

	requireCritical(t, f.history.Commit(f.workflow, uuid.Nil, "nested", nil), "commit with a relative working dir")
	requireCritical(t, f.history.RestoreTo(f.workflow, uuid.Nil, "nested"), "restore with a relative working dir")
}

func TestStoreRefusesAProjectDirItDidNotSnapshot(t *testing.T) {
	f := newFixture(t)

	f.write("code.txt", "baseline\n")
	f.commit(uuid.Nil)

	other := t.TempDir()
	if err := os.WriteFile(filepath.Join(other, "someone-elses.txt"), []byte("untouched\n"), 0o644); err != nil {
		t.Fatalf("write into the other dir: %v", err)
	}

	requireCritical(t, f.history.Commit(f.workflow, uuid.New(), other, nil), "commit against a different working dir")
	requireCritical(t, f.history.RestoreTo(f.workflow, uuid.Nil, other), "restore into a different working dir")

	if _, err := os.Stat(filepath.Join(other, "someone-elses.txt")); err != nil {
		t.Fatalf("the unrelated directory was rewritten: %v", err)
	}
}

func contains(body, want string) bool {
	return strings.Contains(body, want)
}

func longBody() string {
	body := ""
	for i := 0; i < 40; i++ {
		body += "a stable line of text\n"
	}

	return body
}
