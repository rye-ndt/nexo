package core_itf

import "github.com/google/uuid"

type TemplateHelper interface {
	// Blocked says why a template cannot be drafted right now, and is empty when
	// one can be.
	Blocked() string
	Draft(name string, role string) (*Template, error)
	Drafting(agentID uuid.UUID) bool
	Deliver(agentID uuid.UUID, template *Template) error
}
