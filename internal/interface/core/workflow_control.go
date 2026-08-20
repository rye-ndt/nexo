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
	StartWorkflow(workflowID uuid.UUID) error
	PauseWorkflow(workflowID uuid.UUID) error
	CancelWorkflow(workflowID uuid.UUID) error
	WorkflowState(workflowID uuid.UUID) (*WorkflowStatus, error)
	ListWorkflows(limit int) ([]*WorkflowStatus, error)
	ListRoles() ([]*Role, error)
	AnswerReview(stepID uuid.UUID, accepted bool) error
}
