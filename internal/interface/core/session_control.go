package core_itf

import "github.com/google/uuid"

type ControlTaskSpec struct {
	ClientID             string
	Name                 string
	Prompt               string
	TemplateID           uuid.UUID
	Params               map[string]string
	TaskLevel            string
	SystemPrompts        []string
	OutputStructure      string
	DependsOn            []string
	AutoRetry            bool
	ManualAcceptRequired bool
}

type ControlSessionSpec struct {
	WorkingDirPath string
	ContextDirPath string
	Autostart      *bool
	Tasks          []*ControlTaskSpec
}

type ControlSessionRef struct {
	SessionID uuid.UUID
	TaskIDs   map[string]uuid.UUID
	Started   bool
}

type SessionControl interface {
	CreateSession(spec *ControlSessionSpec) (*ControlSessionRef, error)
	StartSession(sessionID uuid.UUID) error
	PauseSession(sessionID uuid.UUID) error
	CancelSession(sessionID uuid.UUID) error
	SessionState(sessionID uuid.UUID) (*SessionStatus, error)
	ListSessions(limit int) ([]*SessionStatus, error)
	ListTemplates() ([]*Template, error)
	AnswerAcceptance(taskID uuid.UUID, accepted bool) error
}
