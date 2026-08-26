package enums

import "slices"

type StepStatus string

const (
	StepNotTaken       StepStatus = "not_taken"
	StepProcessing     StepStatus = "processing"
	StepAwaitingReview StepStatus = "awaiting_review"
	StepCompleted      StepStatus = "completed"
	StepCancelled      StepStatus = "cancelled"
	StepFailed         StepStatus = "failed"
)

var takeable = []StepStatus{
	StepNotTaken,
	StepCancelled,
	StepFailed,
}

var retryable = []StepStatus{
	StepFailed,
	StepCancelled,
}

var removable = []StepStatus{
	StepCompleted,
	StepCancelled,
}

var reached = []StepStatus{
	StepCompleted,
	StepFailed,
}

var cancellable = []StepStatus{
	StepNotTaken,
	StepProcessing,
	StepAwaitingReview,
}

func (s StepStatus) Takeable() bool {
	return slices.Contains(takeable, s)
}

func (s StepStatus) Retryable() bool {
	return slices.Contains(retryable, s)
}

func (s StepStatus) Removable() bool {
	return slices.Contains(removable, s)
}

func (s StepStatus) Reached() bool {
	return slices.Contains(reached, s)
}

func (s StepStatus) Cancellable() bool {
	return slices.Contains(cancellable, s)
}

type WorkflowHalt string

const (
	HaltNone      WorkflowHalt = ""
	HaltPaused    WorkflowHalt = "paused"
	HaltCancelled WorkflowHalt = "cancelled"
)

func (h WorkflowHalt) Park(s StepStatus) (StepStatus, bool) {
	switch {
	case h == HaltCancelled && s.Cancellable():
		return StepCancelled, true
	case h == HaltPaused && s == StepProcessing:
		return StepNotTaken, true
	default:
		return s, false
	}
}

type WorkflowEvent string

const (
	WorkflowDrained           WorkflowEvent = "workflow_drained"
	WorkflowStepCreated       WorkflowEvent = "step_created"
	WorkflowStepStatusChanged WorkflowEvent = "step_status_changed"
	WorkflowStepResulted      WorkflowEvent = "step_reported"
	WorkflowStepDropped       WorkflowEvent = "step_dropped"
)

type WorkflowStatus string

const (
	WorkflowInit       WorkflowStatus = "init"
	WorkflowProcessing WorkflowStatus = "processing"
	WorkflowPaused     WorkflowStatus = "paused"
	WorkflowCompleted  WorkflowStatus = "completed"
)

type FileChangeType string

const (
	FileAdded    FileChangeType = "added"
	FileModified FileChangeType = "modified"
	FileDeleted  FileChangeType = "deleted"
	FileRenamed  FileChangeType = "renamed"
)

func (f FileChangeType) String() string {
	return string(f)
}

type Effort string

const (
	EffortExhaustive Effort = "exhaustive"
	EffortDeep       Effort = "deep"
	EffortStandard   Effort = "standard"
	EffortQuick      Effort = "quick"
)

var efforts = []Effort{
	EffortQuick,
	EffortStandard,
	EffortDeep,
	EffortExhaustive,
}

func Efforts() []Effort {
	return slices.Clone(efforts)
}

func (l Effort) Valid() bool {
	return slices.Contains(efforts, l)
}

func (l Effort) String() string {
	return string(l)
}
