package output_itf

import "hexago/internal/helpers/enums"

// TokenPrices is what the vendor charges for one model, in US dollars per million
// tokens. Every field is optional and nil means "not told": zero is a real price a
// free model can have, so it cannot stand in for a blank field. CachedInput falls
// back to Input when it is nil.
type TokenPrices struct {
	Input       *float64 `json:"input_per_mtok" validate:"omitempty,gte=0"`
	CachedInput *float64 `json:"cached_input_per_mtok" validate:"omitempty,gte=0"`
	Output      *float64 `json:"output_per_mtok" validate:"omitempty,gte=0"`
}

type ModelReady func(model enums.ModelName) bool

type AgentDefault struct {
	Model         enums.ModelName     `json:"model" validate:"required,model_name"`
	ThinkingLevel enums.ThinkingLevel `json:"thinking_level" validate:"required,thinking_level"`
}

// Prices hang off the model, not the step level: two levels that name the same model
// are billed the same, and the effort a level asks for changes how many tokens are
// spent, never what one costs. SetModelPrices writes the three prices of one model
// and a nil argument clears them; the step level's own writer only ever moves the
// model and the effort.
type UserConfig interface {
	AgentDefaults() map[enums.Effort]*AgentDefault
	AgentDefault(level enums.Effort) (*AgentDefault, error)
	SetAgentDefault(level enums.Effort, agentDefault *AgentDefault) error
	ModelPrice(model enums.ModelName) *TokenPrices
	SetModelPrices(model enums.ModelName, prices *TokenPrices) error
	Onboarded() bool
	CompleteOnboarding() error
	Language() enums.Language
	SetLanguage(language enums.Language) error
	Autopilot() bool
	SetAutopilot(on bool) error
	MaxRunningAgents() int
	SetMaxRunningAgents(limit int) error
}
