package enums

import "slices"

type ModelName string

const (
	Fable          ModelName = "fable"
	Opus           ModelName = "opus"
	Sonnet         ModelName = "sonnet"
	Haiku          ModelName = "haiku"
	Deepseek4Flash ModelName = "opencode/deepseek-v4-flash-free"
	GPT56Sol       ModelName = "gpt-5.6-sol"
	GPT56Terra     ModelName = "gpt-5.6-terra"
	GPT56Luna      ModelName = "gpt-5.6-luna"
	ModelUnknown   ModelName = "unknown"
)

var modelNames = []ModelName{
	Fable,
	Opus,
	Sonnet,
	Haiku,
	Deepseek4Flash,
	GPT56Sol,
	GPT56Terra,
	GPT56Luna,
}

func ModelNames() []ModelName {
	return slices.Clone(modelNames)
}

func (m ModelName) String() string {
	return string(m)
}

func (m ModelName) Valid() bool {
	return slices.Contains(modelNames, m)
}

func (m ModelName) DisplayName() string {
	switch m {
	case Fable:
		return "Claude Fable"
	case Opus:
		return "Claude Opus"
	case Sonnet:
		return "Claude Sonnet"
	case Haiku:
		return "Claude Haiku"
	case Deepseek4Flash:
		return "Deepseek V4 Flash"
	case GPT56Sol:
		return "GPT-5.6 Sol"
	case GPT56Terra:
		return "GPT-5.6 Terra"
	case GPT56Luna:
		return "GPT-5.6 Luna"
	default:
		return "Unknown Model"
	}
}

func (m ModelName) HarnessTool() AgentHarness {
	switch m {
	case Deepseek4Flash:
		return OpenCode
	case GPT56Sol, GPT56Terra, GPT56Luna:
		return Codex
	default:
		return ClaudeCode
	}
}

type AgentHarness string

const (
	ClaudeCode AgentHarness = "claude_code"
	OpenCode   AgentHarness = "open_code"
	Codex      AgentHarness = "codex"
)

type ThinkingLevel string

const (
	LowThinking   ThinkingLevel = "low"
	MedThinking   ThinkingLevel = "medium"
	HighThinking  ThinkingLevel = "high"
	XHighThinking ThinkingLevel = "xhigh"
	MaxThinking   ThinkingLevel = "max"
)

var thinkingLevels = []ThinkingLevel{
	LowThinking,
	MedThinking,
	HighThinking,
	XHighThinking,
	MaxThinking,
}

func ThinkingLevels() []ThinkingLevel {
	return slices.Clone(thinkingLevels)
}

func (t ThinkingLevel) Valid() bool {
	return slices.Contains(thinkingLevels, t)
}

func (t ThinkingLevel) String() string {
	return string(t)
}

func (a AgentHarness) String() string {
	return string(a)
}

func (a AgentHarness) DisplayName() string {
	switch a {
	case ClaudeCode:
		return "Claude Code"
	case OpenCode:
		return "OpenCode"
	case Codex:
		return "Codex"
	default:
		return string(a)
	}
}

type InstallationStage string

const (
	InstallStageResolve  InstallationStage = "resolve"
	InstallStageDownload InstallationStage = "download"
	InstallStageExtract  InstallationStage = "extract"
	InstallStageDone     InstallationStage = "done"
)

func (s InstallationStage) String() string {
	return string(s)
}

type AgentInstanceStatus string

const (
	Healthy       AgentInstanceStatus = "healthy"
	AwaitingHuman AgentInstanceStatus = "awaiting_human"
	NotResponding AgentInstanceStatus = "not_responding"
	Terminated    AgentInstanceStatus = "terminated"
)
