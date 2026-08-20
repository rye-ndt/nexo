package input_itf

import (
	"time"

	"hexago/internal/helpers/enums"

	"github.com/google/uuid"
)

type InstallProgress struct {
	Stage      enums.InstallationStage `json:"stage"`
	Downloaded int64                   `json:"downloaded"`
	Total      int64                   `json:"total"`
}

type AgentStatus struct {
	Name          string `json:"name"`
	Installed     bool   `json:"installed"`
	InstanceCount int    `json:"instance_count"`
	LoggedIn      bool   `json:"logged_in"`
	Version       string `json:"version"`
}

// Used is what the agent's window holds right now and falls back when it compacts.
// The other three only ever grow: they are what the vendor charges for, summed over
// every turn. Billed is the output the agent wrote. Input is the prompt it was
// charged full price to read, cache writes included. Cached is the prompt it read
// back out of the cache, which every vendor bills at a fraction of the input rate —
// on a long run it dwarfs Input, so folding the two together would overstate the
// bill several times over.
type ContextUsage struct {
	Total  int `json:"total"`
	Used   int `json:"used"`
	Billed int `json:"billed"`
	Input  int `json:"input"`
	Cached int `json:"cached"`
}

type Activity struct {
	Seq  int       `json:"seq"`
	At   time.Time `json:"at"`
	Text string    `json:"text"`
}

type AgentAdmin interface {
	Auth() (string, error)
	SubmitAuthCode(code string) error
	Logout() error
	Status() (*AgentStatus, error)
	SupportedModels() []enums.ModelName
	Install(onProgress func(InstallProgress)) error
	Uninstall() error
	Shutdown()
}

type AgentHarness interface {
	AgentAdmin
	Support(name enums.ModelName) bool
	Spawn(
		name enums.ModelName,
		thinkingLevel enums.ThinkingLevel,
		instructions []string,
		workdir string,
	) (uuid.UUID, error)
	Send(id string, message string) error
	Listen(id string) (<-chan string, error)
	Alive(id string) (time.Time, error)
	Usage(id string) (*ContextUsage, error)
	Activity(id string) ([]Activity, error)
	Kill(id string) error
}
