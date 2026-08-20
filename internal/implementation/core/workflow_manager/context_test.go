package workflow_manager

import (
	"strings"
	"testing"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"

	"github.com/google/uuid"
)

const structureGate = "Structured output is a hard gate"

const testStructure = "verdict: ship | fix first\nreasons:\n  - detail: what makes the call"

func addStepWithOutput(t *testing.T, manager core_itf.WorkflowManager, workflow uuid.UUID, structure string, prompts ...string) uuid.UUID {
	t.Helper()

	stepID, err := manager.AddStep(workflow, &core_itf.AddStep{
		Name:            "review",
		OutputStructure: structure,
		AgentSpecs:      &core_itf.AgentRequest{Name: enums.Sonnet, Instructions: prompts},
	})
	if err != nil {
		t.Fatalf("add step: %v", err)
	}

	return stepID
}

func promptsFor(t *testing.T, manager core_itf.WorkflowManager, workflow, stepID uuid.UUID) string {
	t.Helper()

	specs, err := manager.ReadySteps(workflow)
	if err != nil {
		t.Fatalf("ready steps: %v", err)
	}

	for _, spec := range specs {
		if spec.StepID == stepID {
			return strings.Join(spec.AgentSpecs.Instructions, "\n")
		}
	}

	t.Fatalf("step %v is not ready", stepID)

	return ""
}

func TestStructuredOutputReachesTheAgent(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	stepID := addStepWithOutput(t, manager, workflow, testStructure, "you review code")

	prompt := promptsFor(t, manager, workflow, stepID)

	if !strings.Contains(prompt, "you review code") {
		t.Fatal("the role prompt was dropped on the way to the agent")
	}

	if !strings.Contains(prompt, structureGate) {
		t.Fatal("the structured output protocol is missing")
	}

	if !strings.Contains(prompt, testStructure) {
		t.Fatal("the structure itself is missing from the prompt")
	}

	if !strings.Contains(prompt, "Report the step as failed") {
		t.Fatal("the prompt does not tell the agent what to do when it cannot comply")
	}
}

func TestFreeOutputAddsNoStructureProtocol(t *testing.T) {
	manager, _ := newManager(t)
	workflow := newWorkflow(t, manager)

	stepID := addStepWithOutput(t, manager, workflow, "   \n  ")

	if strings.Contains(promptsFor(t, manager, workflow, stepID), structureGate) {
		t.Fatal("a free-output step was given the structured output protocol")
	}
}
