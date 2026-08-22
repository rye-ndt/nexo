package core_itf

import "github.com/google/uuid"

type ControlStepSpec struct {
	ClientID        string
	Name            string
	Prompt          string
	RoleID          uuid.UUID
	Inputs          map[string]string
	Effort          string
	Instructions    []string
	OutputStructure string
	DependsOn       []string
	AutoRetry       bool
	PauseForReview  bool
}

type ControlWorkflowSpec struct {
	ProjectDirPath string
	Autostart      *bool
	Steps          []*ControlStepSpec
}

type ControlWorkflowRef struct {
	WorkflowID uuid.UUID
	StepIDs    map[string]uuid.UUID
	Started    bool
}

type WorkflowControl interface {
	CreateWorkflow(spec *ControlWorkflowSpec) (*ControlWorkflowRef, error)
	ListWorkflows(limit int) ([]*WorkflowStatus, error)
}
