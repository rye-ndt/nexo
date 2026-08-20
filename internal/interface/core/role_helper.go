package core_itf

import "github.com/google/uuid"

type DraftGraphStep struct {
	ID        string   `json:"id"`
	Title     string   `json:"title"`
	RoleID    string   `json:"role_id"`
	DependsOn []string `json:"depends_on"`
}

type DraftRequest struct {
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	WorkflowName string            `json:"workflow_name"`
	ProjectDir   string            `json:"project_dir"`
	Steps        []*DraftGraphStep `json:"steps"`
}

type RoleHelper interface {
	Blocked() string
	Draft(req *DraftRequest) (*Role, error)
	Drafting(agentID uuid.UUID) bool
	Deliver(agentID uuid.UUID, role *Role) error
}
