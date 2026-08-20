package coordinator_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"hexago/internal/helpers/enums"
	"hexago/internal/implementation/core/coordinator"
	"hexago/internal/implementation/core/workflow_manager"
	"hexago/internal/implementation/input/storage"
	"hexago/internal/implementation/input/workspace_history"
	"hexago/internal/implementation/output/message_queue"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

const (
	appFile         = "src/app.go"
	utilFile        = "src/util.go"
	readmeFile      = "docs/readme.md"
	featureFile     = "src/feature.go"
	featureTestFile = "src/feature_test.go"
	changelogFile   = "CHANGELOG.md"

	agentsDoc  = ".harness/context/AGENTS.md"
	gotchasDoc = ".harness/context/.agent/gotchas.md"
)

const (
	appBodyA = "package app\n\nfunc main() {}\n"
	appBodyB = "package app\n\nfunc main() {\n\tprintln(\"wired by the second step\")\n}\n"

	utilBody = "package app\n\nfunc Util() string {\n\treturn \"the first step wrote this helper\"\n}\n"

	readmeBodyA = "# docs\n\nplaceholder\n"
	readmeBodyB = "# docs\n\nthe second step explained what the first step built\n"

	featureBodyC = "package app\n\nimport \"strconv\"\n\nfunc Feature(n int) string {\n\treturn strconv.Itoa(n) + \" third step\"\n}\n"
	featureBodyD = "package app\n\nimport \"strconv\"\n\nfunc Feature(n int) string {\n\treturn strconv.Itoa(n*2) + \" fourth step\"\n}\n"

	featureTestBody = "package app\n\nimport \"testing\"\n\nfunc TestFeature(t *testing.T) {\n\tif Feature(1) == \"\" {\n\t\tt.Fatal(\"empty\")\n\t}\n}\n"

	changelogBody = "# changelog\n\n- the fourth step doubled the feature\n"

	gotchaNote = "\n- the shadow repo excludes this file on purpose\n"
	agentsNote = "\n- learned by an agent and kept across workflows\n"
)

type mutation struct {
	path     string
	body     string
	remove   bool
	appendTo bool
	change   enums.FileChangeType
}

type step struct {
	after     string
	mutations []mutation
}

var chain = []string{"a", "b", "c", "d"}

var script = map[string]step{
	"a": {mutations: []mutation{
		{path: appFile, body: appBodyA, change: enums.FileAdded},
		{path: utilFile, body: utilBody, change: enums.FileAdded},
		{path: readmeFile, body: readmeBodyA, change: enums.FileAdded},
	}},
	"b": {after: "a", mutations: []mutation{
		{path: appFile, body: appBodyB, change: enums.FileModified},
		{path: readmeFile, body: readmeBodyB, change: enums.FileModified},
	}},
	"c": {after: "b", mutations: []mutation{
		{path: featureFile, body: featureBodyC, change: enums.FileAdded},
		{path: featureTestFile, body: featureTestBody, change: enums.FileAdded},
		{path: utilFile, remove: true, change: enums.FileDeleted},
		{path: gotchasDoc, body: gotchaNote, appendTo: true},
	}},
	"d": {after: "c", mutations: []mutation{
		{path: featureFile, body: featureBodyD, change: enums.FileModified},
		{path: readmeFile, remove: true, change: enums.FileDeleted},
		{path: changelogFile, body: changelogBody, change: enums.FileAdded},
		{path: agentsDoc, body: agentsNote, appendTo: true},
	}},
}

type silentLogger struct{}

func (silentLogger) Debug(string, ...any) {}
func (silentLogger) Info(string, ...any)  {}
func (silentLogger) Warn(string, ...any)  {}
func (silentLogger) Error(string, ...any) {}

type pending struct {
	agentID uuid.UUID
	step    string
}

type harness struct {
	t         *testing.T
	dir       string
	workflow  uuid.UUID
	ids       map[string]uuid.UUID
	workflows core_itf.WorkflowManager
	history   input_itf.WorkspaceHistory
	coord     core_itf.Coordinator

	mu       sync.Mutex
	queued   []pending
	executed []string
	problems []string
}

// stubAgents stands in for the whole agent side. Prompted work is queued rather than
// done inline: a real agent reports long after the coordinator has drained the event
// its assignment emitted, and steppedWorkflows replays that ordering deterministically.
type stubAgents struct{ h *harness }

func (s *stubAgents) SupportedAgents() (map[enums.AgentHarness][]enums.ModelName, error) {
	return nil, nil
}

func (s *stubAgents) Admin(enums.AgentHarness) (input_itf.AgentAdmin, error) { return nil, nil }

func (s *stubAgents) RequestInstance(*core_itf.AgentRequest) (*core_itf.Agent, error) {
	return &core_itf.Agent{ID: uuid.New(), HealthStatus: enums.Healthy}, nil
}

func (s *stubAgents) Send(agentID uuid.UUID, message string) error {
	s.h.queue(agentID, stepNameOf(message))

	return nil
}

func (s *stubAgents) Listen(uuid.UUID) (<-chan string, error) { return nil, nil }

func (s *stubAgents) ContextUsage(uuid.UUID) (*input_itf.ContextUsage, error) { return nil, nil }

func (s *stubAgents) Activity(uuid.UUID) ([]input_itf.Activity, error) { return nil, nil }

func (s *stubAgents) Kill(uuid.UUID) error { return nil }

func (s *stubAgents) HeartBeat(uuid.UUID) error { return nil }

// steppedWorkflows is the real workflow manager plus a clock for the stubbed agents: a
// queued agent runs after the coordinator has asked for ready steps and found none, so
// every report lands on an idle loop, with no timers and no sleeps involved.
type steppedWorkflows struct {
	core_itf.WorkflowManager
	h *harness
}

func (s *steppedWorkflows) ReadySteps(workflow uuid.UUID) ([]*core_itf.StepSpec, error) {
	specs, err := s.WorkflowManager.ReadySteps(workflow)

	s.h.flush()

	return specs, err
}

func stepNameOf(prompt string) string {
	line, _, _ := strings.Cut(prompt, "\n")

	return strings.TrimSpace(strings.TrimPrefix(line, "# Step:"))
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	h := &harness{t: t, dir: t.TempDir(), ids: map[string]uuid.UUID{}}

	store, err := storage.New(filepath.Join(t.TempDir(), "harness.db"))
	if err != nil {
		t.Fatalf("init storage: %v", err)
	}

	cfg := &input_itf.WorkflowConfig{
		HeartbeatTimeout:       time.Hour,
		HeartbeatScanInterval:  time.Hour,
		AgentHeartbeatInterval: time.Hour,
	}

	workflows, err := workflow_manager.InitV1(cfg, store.StepStore(), message_queue.InitV1())
	if err != nil {
		t.Fatalf("init workflow manager: %v", err)
	}

	t.Cleanup(workflows.Stop)

	history, err := workspace_history.InitV1(filepath.Join(t.TempDir(), "workflows"))
	if err != nil {
		t.Fatalf("init workspace history: %v", err)
	}

	coord, err := coordinator.InitV1(cfg, &steppedWorkflows{WorkflowManager: workflows, h: h}, &stubAgents{h: h}, history, silentLogger{})
	if err != nil {
		t.Fatalf("init coordinator: %v", err)
	}

	t.Cleanup(coord.Stop)

	h.workflows = workflows
	h.history = history
	h.coord = coord

	workflow, err := workflows.NewWorkflow(&core_itf.InitWorkflow{ProjectDirPath: h.dir})
	if err != nil {
		t.Fatalf("new workflow: %v", err)
	}

	h.workflow = workflow

	for _, name := range chain {
		deps := []uuid.UUID{}
		if after := script[name].after; after != "" {
			deps = append(deps, h.ids[after])
		}

		stepID, err := workflows.AddStep(workflow, &core_itf.AddStep{
			Name:       name,
			DependsOn:  deps,
			AgentSpecs: &core_itf.AgentRequest{Name: enums.Sonnet},
		})
		if err != nil {
			t.Fatalf("add step %s: %v", name, err)
		}

		h.ids[name] = stepID
	}

	return h
}

func (h *harness) queue(agentID uuid.UUID, step string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.queued = append(h.queued, pending{agentID: agentID, step: step})
}

func (h *harness) flush() {
	h.mu.Lock()
	work := h.queued
	h.queued = nil
	h.mu.Unlock()

	for _, item := range work {
		h.perform(item)
	}
}

func (h *harness) perform(item pending) {
	s, found := script[item.step]
	if !found {
		h.fail("an agent was prompted for unknown step %q", item.step)

		return
	}

	h.requireSnapshotOf(s.after, item.step)

	for _, m := range s.mutations {
		h.apply(m)
	}

	h.mu.Lock()
	h.executed = append(h.executed, item.step)
	h.mu.Unlock()

	if err := h.workflows.Report(item.agentID, enums.StepCompleted, []*core_itf.Handoff{{
		TLDR:    "step " + item.step + " finished",
		Outcome: "step " + item.step + " changed the files it owns",
	}}); err != nil {
		h.fail("report of %s: %v", item.step, err)
	}
}

// requireSnapshotOf is the property the whole feature rests on: by the time an agent is
// free to write, its predecessor must already be committed, and that commit must hold
// exactly what the predecessor changed and nothing this agent is about to do.
func (h *harness) requireSnapshotOf(predecessor, step string) {
	snapshot := uuid.Nil
	if predecessor != "" {
		snapshot = h.ids[predecessor]
	}

	changes, err := h.history.Diff(h.workflow, snapshot)
	if err != nil {
		h.fail("no workspace snapshot of %q when %q started writing: %v", predecessor, step, err)

		return
	}

	if predecessor == "" {
		return
	}

	got, want := changeSet(changes), trackedChanges(script[predecessor])
	if !reflect.DeepEqual(got, want) {
		h.fail("snapshot of %q read when %q started writing = %v, want %v", predecessor, step, got, want)
	}
}

func (h *harness) apply(m mutation) {
	path := filepath.Join(h.dir, filepath.FromSlash(m.path))

	if m.remove {
		if err := os.Remove(path); err != nil {
			h.fail("delete %s: %v", m.path, err)
		}

		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		h.fail("create dir for %s: %v", m.path, err)

		return
	}

	body := []byte(m.body)

	if m.appendTo {
		existing, err := os.ReadFile(path)
		if err != nil {
			h.fail("read %s: %v", m.path, err)

			return
		}

		body = append(existing, body...)
	}

	if err := os.WriteFile(path, body, 0o644); err != nil {
		h.fail("write %s: %v", m.path, err)
	}
}

func (h *harness) fail(format string, args ...any) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.problems = append(h.problems, fmt.Sprintf(format, args...))
}

func (h *harness) reportProblems() {
	h.t.Helper()

	h.mu.Lock()
	defer h.mu.Unlock()

	for _, problem := range h.problems {
		h.t.Error(problem)
	}

	h.problems = nil
}

func (h *harness) ran() []string {
	h.mu.Lock()
	defer h.mu.Unlock()

	return append([]string{}, h.executed...)
}

func (h *harness) forget() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.executed = nil
}

func (h *harness) runToCompletion(expected ...string) {
	h.t.Helper()

	if err := h.coord.Run(h.workflow); err != nil {
		h.t.Fatalf("run workflow: %v", err)
	}

	h.waitFor("steps "+strings.Join(expected, ",")+" to run and be snapshotted", func() bool {
		if !reflect.DeepEqual(h.ran(), expected) {
			return false
		}

		_, err := h.history.Diff(h.workflow, h.ids[expected[len(expected)-1]])

		return err == nil
	})

	h.reportProblems()
}

func (h *harness) waitFor(what string, done func() bool) {
	h.t.Helper()

	deadline := time.Now().Add(15 * time.Second)

	for time.Now().Before(deadline) {
		if done() {
			return
		}

		time.Sleep(time.Millisecond)
	}

	h.t.Fatalf("timed out waiting for %s, steps run so far: %v", what, h.ran())
}

func (h *harness) statuses() map[string]*core_itf.StepResult {
	h.t.Helper()

	status, err := h.workflows.Status(h.workflow)
	if err != nil {
		h.t.Fatalf("workflow status: %v", err)
	}

	byName := map[string]*core_itf.StepResult{}
	for name, id := range h.ids {
		byName[name] = status.Steps[id]
	}

	return byName
}

func changeSet(changes []*input_itf.FileChange) map[string]enums.FileChangeType {
	set := map[string]enums.FileChangeType{}

	for _, change := range changes {
		set[change.Path] = change.ChangeType
	}

	return set
}

func trackedChanges(s step) map[string]enums.FileChangeType {
	set := map[string]enums.FileChangeType{}

	for _, m := range s.mutations {
		if m.change == "" {
			continue
		}

		set[m.path] = m.change
	}

	return set
}

// treeAfter replays the script to the tree the working dir should hold once the given
// steps have run. Knowledge base files are left out: they never enter the shadow repo.
func treeAfter(steps ...string) map[string]string {
	tree := map[string]string{}

	for _, name := range steps {
		for _, m := range script[name].mutations {
			switch {
			case m.appendTo:
				continue
			case m.remove:
				delete(tree, m.path)
			default:
				tree[m.path] = m.body
			}
		}
	}

	return tree
}

func readTree(t *testing.T, dir string) map[string]string {
	t.Helper()

	tree := map[string]string{}

	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		rel = filepath.ToSlash(rel)

		if rel == ".harness" {
			return fs.SkipDir
		}

		if entry.IsDir() {
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		tree[rel] = string(body)

		return nil
	})
	if err != nil {
		t.Fatalf("read working dir: %v", err)
	}

	return tree
}

func assertTree(t *testing.T, dir string, want map[string]string) {
	t.Helper()

	got := readTree(t, dir)

	for _, path := range sortedKeys(want) {
		body, found := got[path]

		switch {
		case !found:
			t.Errorf("%s is missing from the working dir", path)
		case body != want[path]:
			t.Errorf("%s = %q, want %q", path, body, want[path])
		}
	}

	for _, path := range sortedKeys(got) {
		if _, expected := want[path]; !expected {
			t.Errorf("%s is still in the working dir and should not be", path)
		}
	}
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	return keys
}

func TestRevertToRestoresTheWorkspaceAndRewindsTheGraph(t *testing.T) {
	h := newHarness(t)

	h.runToCompletion(chain...)

	assertTree(t, h.dir, treeAfter(chain...))

	if err := h.coord.RevertTo(h.workflow, h.ids["b"]); err != nil {
		t.Fatalf("revert to b: %v", err)
	}

	h.reportProblems()

	assertTree(t, h.dir, treeAfter("a", "b"))

	if body := mustRead(t, h.dir, utilFile); body != utilBody {
		t.Errorf("%s deleted by c was not restored, got %q", utilFile, body)
	}

	if body := mustRead(t, h.dir, readmeFile); body != readmeBodyB {
		t.Errorf("%s deleted by d was not restored to b's content, got %q", readmeFile, body)
	}

	if body := mustRead(t, h.dir, appFile); body != appBodyB {
		t.Errorf("%s = %q, want b's content %q", appFile, body, appBodyB)
	}

	for _, path := range []string{featureFile, featureTestFile, changelogFile} {
		if _, err := os.Stat(filepath.Join(h.dir, filepath.FromSlash(path))); !os.IsNotExist(err) {
			t.Errorf("%s was created after b and should be gone, stat err = %v", path, err)
		}
	}

	if body := mustRead(t, h.dir, agentsDoc); !strings.Contains(body, agentsNote) {
		t.Errorf("%s lost the note the fourth step wrote, got %q", agentsDoc, body)
	}

	if body := mustRead(t, h.dir, gotchasDoc); !strings.Contains(body, gotchaNote) {
		t.Errorf("%s lost the note the third step wrote, got %q", gotchasDoc, body)
	}

	if _, err := os.Stat(filepath.Join(h.dir, ".harness", "context", ".agent", "flows", "build-test-run.md")); err != nil {
		t.Errorf("the knowledge base was rolled back: %v", err)
	}

	reports := h.statuses()

	for _, name := range []string{"a", "b"} {
		if reports[name].Status != enums.StepCompleted {
			t.Errorf("step %s status = %s, want %s", name, reports[name].Status, enums.StepCompleted)
		}

		if len(reports[name].Handoffs) == 0 {
			t.Errorf("step %s lost its handoffs", name)
			continue
		}

		if got := reports[name].Handoffs[0].Step; got != name {
			t.Errorf("step %s handoff names %q, want the step itself", name, got)
		}
	}

	for _, name := range []string{"c", "d"} {
		if reports[name].Status != enums.StepNotTaken {
			t.Errorf("step %s status = %s, want %s", name, reports[name].Status, enums.StepNotTaken)
		}

		if len(reports[name].Handoffs) != 0 {
			t.Errorf("step %s kept %d handoffs for work that no longer exists", name, len(reports[name].Handoffs))
		}
	}
}

func TestWorkflowRunsAgainFromTheRevertPoint(t *testing.T) {
	h := newHarness(t)

	h.runToCompletion(chain...)

	if err := h.coord.RevertTo(h.workflow, h.ids["b"]); err != nil {
		t.Fatalf("revert to b: %v", err)
	}

	specs, err := h.workflows.ReadySteps(h.workflow)
	if err != nil {
		t.Fatalf("ready steps: %v", err)
	}

	if len(specs) != 1 || specs[0].StepID != h.ids["c"] {
		t.Fatalf("after a revert the pool must offer c again, got %d ready steps", len(specs))
	}

	h.forget()
	h.runToCompletion("c", "d")

	assertTree(t, h.dir, treeAfter(chain...))
}

func mustRead(t *testing.T, dir, path string) string {
	t.Helper()

	body, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(path)))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	return string(body)
}
