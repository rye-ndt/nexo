package output_itf

import (
	"context"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
)

type AgentInfo struct {
	ID     string                 `json:"id"`
	Status *input_itf.AgentStatus `json:"status"`
}

type ApprovalInfo struct {
	ID          string                     `json:"id"`
	AgentID     string                     `json:"agent_id"`
	TaskID      string                     `json:"task_id"`
	Kind        string                     `json:"kind"`
	Question    string                     `json:"question"`
	Detail      string                     `json:"detail"`
	Options     []*core_itf.ApprovalOption `json:"options"`
	MultiSelect bool                       `json:"multi_select"`
	RequestedAt string                     `json:"requested_at"`
}

type TemplateParamInfo struct {
	Description string   `json:"description"`
	Required    bool     `json:"required"`
	Type        string   `json:"type"`
	Default     string   `json:"default"`
	Options     []string `json:"options"`
}

type TemplateInfo struct {
	ID                   string                        `json:"id"`
	Name                 string                        `json:"name"`
	Role                 string                        `json:"role"`
	TaskLevel            string                        `json:"task_level"`
	Retryable            bool                          `json:"retryable"`
	ManualAcceptRequired bool                          `json:"manual_accept_required"`
	Params               map[string]*TemplateParamInfo `json:"params"`
	SystemPrompts        map[string]string             `json:"system_prompts"`
	OutputStructure      string                        `json:"output_structure"`
}

type RunTaskSpec struct {
	ClientID             string   `json:"client_id"`
	Name                 string   `json:"name"`
	Prompt               string   `json:"prompt"`
	TaskLevel            string   `json:"task_level"`
	SystemPrompts        []string `json:"system_prompts"`
	OutputStructure      string   `json:"output_structure"`
	DependsOn            []string `json:"depends_on"`
	AutoRetry            bool     `json:"auto_retry"`
	ManualAcceptRequired bool     `json:"manual_accept_required"`
}

type RunSessionSpec struct {
	WorkingDirPath string         `json:"working_dir_path"`
	ContextDirPath string         `json:"context_dir_path"`
	Tasks          []*RunTaskSpec `json:"tasks"`
}

type RunSessionResult struct {
	SessionID string            `json:"session_id"`
	TaskIDs   map[string]string `json:"task_ids"`
}

type HandoverDocInfo struct {
	Task              string            `json:"task"`
	TLDR              string            `json:"tldr"`
	Outcome           string            `json:"outcome"`
	Blockers          map[string]string `json:"blockers"`
	ApprovedDecisions map[string]string `json:"approved_decisions"`
	RejectedDecisions map[string]string `json:"rejected_decisions"`
	CurrentBehaviors  map[string]string `json:"current_behaviors"`
	ChangedBehaviors  map[string]string `json:"changed_behaviors"`
	MustAvoid         map[string]string `json:"must_avoid"`
	Nuances           map[string]string `json:"nuances"`
	KnownGaps         map[string]string `json:"known_gaps"`
}

type FileChangeInfo struct {
	Path        string `json:"path"`
	OldPath     string `json:"old_path"`
	ChangeType  string `json:"change_type"`
	Additions   int    `json:"additions"`
	Deletions   int    `json:"deletions"`
	UnifiedDiff string `json:"unified_diff"`
}

type TaskActivityInfo struct {
	Seq  int    `json:"seq"`
	At   string `json:"at"`
	Text string `json:"text"`
}

// Spent is every token this node has cost across all of its attempts; CostUSD is what
// that comes to at the prices of the model behind it, and is only meaningful when
// Priced is true.
type SessionTaskInfo struct {
	TaskID       string                  `json:"task_id"`
	AgentID      string                  `json:"agent_id"`
	TaskLevel    string                  `json:"task_level"`
	Status       string                  `json:"status"`
	HandoverDocs []*HandoverDocInfo      `json:"handover_docs"`
	ContextUsage *input_itf.ContextUsage `json:"context_usage"`
	Spent        *input_itf.ContextUsage `json:"spent"`
	CostUSD      float64                 `json:"cost_usd"`
	Priced       bool                    `json:"priced"`
	Activity     []*TaskActivityInfo     `json:"activity"`
}

// Priced is false as soon as one node that spent something ran on a model with no
// prices filled in, because a total missing one node reads as the whole bill.
type SessionStatusInfo struct {
	SessionID    string             `json:"session_id"`
	Status       string             `json:"status"`
	Tasks        []*SessionTaskInfo `json:"tasks"`
	TokensBilled int                `json:"tokens_billed"`
	TokensInput  int                `json:"tokens_input"`
	TokensCached int                `json:"tokens_cached"`
	CostUSD      float64            `json:"cost_usd"`
	Priced       bool               `json:"priced"`
	StartedAt    string             `json:"started_at"`
	CompletedAt  string             `json:"completed_at"`
}

type AgentDefaultInfo struct {
	TaskLevel     string `json:"task_level"`
	Model         string `json:"model"`
	ModelLabel    string `json:"model_label"`
	ThinkingLevel string `json:"thinking_level"`
}

// One row per model this app can run, priced or not. The prices are the text the user
// typed, empty when the field is blank, so a price that never round-trips through a
// number keeps the difference between "zero" and "not told".
type ModelPriceInfo struct {
	Model            string `json:"model"`
	ModelLabel       string `json:"model_label"`
	InputPrice       string `json:"input_price"`
	CachedInputPrice string `json:"cached_input_price"`
	OutputPrice      string `json:"output_price"`
}

type ModelOptionInfo struct {
	Model   string `json:"model"`
	Label   string `json:"label"`
	Harness string `json:"harness"`
}

type SessionDraftInfo struct {
	ID        string `json:"id"`
	Doc       string `json:"doc"`
	UpdatedAt string `json:"updated_at"`
}

type MCPServerInfo struct {
	Name         string `json:"name"`
	URL          string `json:"url"`
	Authorized   bool   `json:"authorized"`
	AuthorizedAt string `json:"authorized_at"`
	Account      string `json:"account"`
	Kind         string `json:"kind"`
}

type AgentDefaultOptionsInfo struct {
	TaskLevels     []string           `json:"task_levels"`
	Models         []*ModelOptionInfo `json:"models"`
	ThinkingLevels []string           `json:"thinking_levels"`
}

// FEAPI is the lifecycle the app builder drives. The methods JavaScript actually
// calls are not listed here on purpose: the builder hands the value to Wails as
// `any`, so Wails reflects the concrete type and binds every exported method on
// it. Re-declaring that surface here bound nothing — it only meant each new
// screen's method had to be written twice and could drift.
type FEAPI interface {
	Startup(ctx context.Context)
	Shutdown(ctx context.Context)
}
