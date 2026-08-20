package core_itf

import "github.com/google/uuid"

type Coordinator interface {
	Run(workflow uuid.UUID) error
	Cancel(workflow uuid.UUID) error
	Pause(workflow uuid.UUID) error
	RevertTo(workflow, stepID uuid.UUID) error
	Stop()
}
