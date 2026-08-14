package core_itf

import "github.com/google/uuid"

type DraftGraphNode struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	TemplateID string   `json:"template_id"`
	DependsOn  []string `json:"depends_on"`
}

type DraftRequest struct {
	Name        string            `json:"name"`
	Role        string            `json:"role"`
	SessionName string            `json:"session_name"`
	WorkingDir  string            `json:"working_dir"`
	ContextDir  string            `json:"context_dir"`
	Nodes       []*DraftGraphNode `json:"nodes"`
}

type TemplateHelper interface {
	Blocked() string
	Draft(req *DraftRequest) (*Template, error)
	Drafting(agentID uuid.UUID) bool
	Deliver(agentID uuid.UUID, template *Template) error
}
