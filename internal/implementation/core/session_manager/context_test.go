package session_manager

import (
	"strings"
	"testing"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

const structureGate = "Structured output is a hard gate"

const testStructure = "verdict: ship | fix first\nreasons:\n  - detail: what makes the call"

func addTaskWithOutput(t *testing.T, manager core_itf.SessionManager, session uuid.UUID, structure string, prompts ...string) uuid.UUID {
	t.Helper()

	taskID, err := manager.AddTask(session, &core_itf.AddTask{
		Name:            "review",
		OutputStructure: structure,
		AgentSpecs:      &core_itf.AgentRequest{Name: enums.Sonnet, SystemPrompts: prompts},
	})
	if err != nil {
		t.Fatalf("add task: %v", err)
	}

	return taskID
}

func promptsFor(t *testing.T, manager core_itf.SessionManager, session, taskID uuid.UUID) string {
	t.Helper()

	specs, err := manager.ReadyTasks(session)
	if err != nil {
		t.Fatalf("ready tasks: %v", err)
	}

	for _, spec := range specs {
		if spec.TaskID == taskID {
			return strings.Join(spec.AgentSpecs.SystemPrompts, "\n")
		}
	}

	t.Fatalf("task %v is not ready", taskID)

	return ""
}

func TestTemplatePromptsReachTheAgent(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	taskID := addTaskWithOutput(t, manager, session, "", "you review code")

	if !strings.Contains(promptsFor(t, manager, session, taskID), "you review code") {
		t.Fatal("the template prompt was dropped on the way to the agent")
	}
}

func TestStructuredOutputReachesTheAgent(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	taskID := addTaskWithOutput(t, manager, session, testStructure, "you review code")

	prompt := promptsFor(t, manager, session, taskID)

	if !strings.Contains(prompt, structureGate) {
		t.Fatal("the structured output protocol is missing")
	}

	if !strings.Contains(prompt, testStructure) {
		t.Fatal("the structure itself is missing from the prompt")
	}

	if !strings.Contains(prompt, "Report the task as failed") {
		t.Fatal("the prompt does not tell the agent what to do when it cannot comply")
	}
}

func TestFreeOutputAddsNoStructureProtocol(t *testing.T) {
	manager, _ := newManager(t)
	session := newSession(t, manager)

	taskID := addTaskWithOutput(t, manager, session, "   \n  ")

	if strings.Contains(promptsFor(t, manager, session, taskID), structureGate) {
		t.Fatal("a free-output task was given the structured output protocol")
	}
}
