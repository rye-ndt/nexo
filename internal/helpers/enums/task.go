package enums

import "slices"

type TaskStatus string

const (
	TaskNotTaken       TaskStatus = "not_taken"
	TaskProcessing     TaskStatus = "processing"
	TaskAwaitingAccept TaskStatus = "awaiting_accept"
	TaskCompleted      TaskStatus = "completed"
	TaskCancelled      TaskStatus = "cancelled"
	TaskFailed         TaskStatus = "failed"
)

var takeable = []TaskStatus{
	TaskNotTaken,
	TaskCancelled,
	TaskFailed,
}

var retryable = []TaskStatus{
	TaskFailed,
	TaskCancelled,
}

var removable = []TaskStatus{
	TaskCompleted,
	TaskCancelled,
}

var cancellable = []TaskStatus{
	TaskNotTaken,
	TaskProcessing,
	TaskAwaitingAccept,
}

func (s TaskStatus) Takeable() bool {
	return slices.Contains(takeable, s)
}

func (s TaskStatus) Retryable() bool {
	return slices.Contains(retryable, s)
}

func (s TaskStatus) Removable() bool {
	return slices.Contains(removable, s)
}

func (s TaskStatus) Cancellable() bool {
	return slices.Contains(cancellable, s)
}

type SessionHalt string

const (
	HaltNone      SessionHalt = ""
	HaltPaused    SessionHalt = "paused"
	HaltCancelled SessionHalt = "cancelled"
)

func (h SessionHalt) Park(s TaskStatus) (TaskStatus, bool) {
	switch {
	case h == HaltCancelled && s.Cancellable():
		return TaskCancelled, true
	case h == HaltPaused && s == TaskProcessing:
		return TaskNotTaken, true
	default:
		return s, false
	}
}

type SessionEvent string

const (
	SessionCreated           SessionEvent = "session_created"
	SessionDrained           SessionEvent = "session_drained"
	SessionTaskCreated       SessionEvent = "task_created"
	SessionTaskStatusChanged SessionEvent = "task_status_changed"
	SessionTaskReported      SessionEvent = "task_reported"
	SessionTaskDropped       SessionEvent = "task_dropped"
)

var sessionEvents = []SessionEvent{
	SessionCreated,
	SessionDrained,
	SessionTaskCreated,
	SessionTaskStatusChanged,
	SessionTaskReported,
	SessionTaskDropped,
}

func SessionEvents() []SessionEvent {
	return slices.Clone(sessionEvents)
}

func (se SessionEvent) Event() string {
	return string(se)
}

type SessionStatus string

const (
	SessionInit       SessionStatus = "init"
	SessionProcessing SessionStatus = "processing"
	SessionPaused     SessionStatus = "paused"
	SessionCompleted  SessionStatus = "completed"
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

type TaskLevel string

const (
	MaximumEffortTask TaskLevel = "maximum_effort_task"
	HeavyTask         TaskLevel = "heavy_task"
	DailyTask         TaskLevel = "daily_task"
	LightweightTask   TaskLevel = "lightweight_task"
)

var taskLevels = []TaskLevel{
	LightweightTask,
	DailyTask,
	HeavyTask,
	MaximumEffortTask,
}

func TaskLevels() []TaskLevel {
	return slices.Clone(taskLevels)
}

func (l TaskLevel) Valid() bool {
	return slices.Contains(taskLevels, l)
}

func (l TaskLevel) String() string {
	return string(l)
}
