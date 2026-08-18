package core_itf

import "github.com/google/uuid"

type Coordinator interface {
	Run(session uuid.UUID) error
	Cancel(session uuid.UUID) error
	Pause(session uuid.UUID) error
	RevertTo(session, taskID uuid.UUID) error
	Stop()
}
