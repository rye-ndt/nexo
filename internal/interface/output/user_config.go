package output_itf

import "hexago/internal/helpers/enums"

// TokenPrices is what the vendor charges for this task level, in US dollars per
// million tokens. Every field is optional and nil means "not told": zero is a real
// price a free model can have, so it cannot stand in for a blank field. CachedInput
// falls back to Input when it is nil.
type TokenPrices struct {
	Input       *float64 `json:"input_per_mtok" validate:"omitempty,gte=0"`
	CachedInput *float64 `json:"cached_input_per_mtok" validate:"omitempty,gte=0"`
	Output      *float64 `json:"output_per_mtok" validate:"omitempty,gte=0"`
}

type AgentDefault struct {
	Model         enums.ModelName     `json:"model" validate:"required,model_name"`
	ThinkingLevel enums.ThinkingLevel `json:"thinking_level" validate:"required,thinking_level"`
	Prices        *TokenPrices        `json:"prices" validate:"omitempty"`
}

// The two writers are deliberately narrow: the settings screen changes a model and
// a price on separate keystrokes, so each one keeps what the other owns. SetAgentDefault
// writes the model and the effort and leaves the stored prices alone, whatever the
// argument carries in Prices; SetAgentDefaultPrices writes the prices and leaves the
// model alone, and a nil argument clears them.
type UserConfig interface {
	AgentDefaults() map[enums.TaskLevel]*AgentDefault
	AgentDefault(level enums.TaskLevel) (*AgentDefault, error)
	SetAgentDefault(level enums.TaskLevel, agentDefault *AgentDefault) error
	SetAgentDefaultPrices(level enums.TaskLevel, prices *TokenPrices) error
	Onboarded() bool
	CompleteOnboarding() error
	Autopilot() bool
	SetAutopilot(on bool) error
}
