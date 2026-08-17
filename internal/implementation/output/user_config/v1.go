package user_config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	output_itf "hexago/internal/interface/output"
)

type file struct {
	AgentDefaults map[enums.TaskLevel]*output_itf.AgentDefault `json:"agent_defaults"`
	Onboarded     bool                                         `json:"onboarded"`
	Autopilot     bool                                         `json:"autopilot"`
}

type v1 struct {
	path string
	mu   sync.RWMutex
	cfg  *file
}

var seedAgentDefaults = map[enums.TaskLevel]*output_itf.AgentDefault{
	enums.LightweightTask:   {Model: enums.Haiku, ThinkingLevel: enums.LowThinking},
	enums.DailyTask:         {Model: enums.Sonnet, ThinkingLevel: enums.MedThinking},
	enums.HeavyTask:         {Model: enums.Opus, ThinkingLevel: enums.HighThinking},
	enums.MaximumEffortTask: {Model: enums.Fable, ThinkingLevel: enums.MaxThinking},
}

func InitV1(path string) (output_itf.UserConfig, error) {
	if path == "" {
		return nil, custom_error.Critical("user config path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, custom_error.Critical("cannot create user config dir: %v", err)
	}

	cfg, err := read(path)
	if err != nil {
		return nil, err
	}

	c := &v1{path: path, cfg: cfg}

	if err := c.write(); err != nil {
		return nil, err
	}

	return c, nil
}

func read(path string) (*file, error) {
	cfg := &file{AgentDefaults: map[enums.TaskLevel]*output_itf.AgentDefault{}}

	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, custom_error.Critical("cannot read user config %s: %v", path, err)
	}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, cfg); err != nil {
			return nil, custom_error.Critical("cannot decode user config %s: %v", path, err)
		}
	}

	if cfg.AgentDefaults == nil {
		cfg.AgentDefaults = map[enums.TaskLevel]*output_itf.AgentDefault{}
	}

	for _, level := range enums.TaskLevels() {
		cfg.AgentDefaults[level] = repaired(level, cfg.AgentDefaults[level])
	}

	for level := range cfg.AgentDefaults {
		if !level.Valid() {
			delete(cfg.AgentDefaults, level)
		}
	}

	return cfg, nil
}

func repaired(level enums.TaskLevel, stored *output_itf.AgentDefault) *output_itf.AgentDefault {
	if stored != nil {
		if helpers.ValidateStruct(stored) == nil {
			return cloneAgentDefault(stored)
		}

		withoutPrices := &output_itf.AgentDefault{Model: stored.Model, ThinkingLevel: stored.ThinkingLevel}
		if helpers.ValidateStruct(withoutPrices) == nil {
			return withoutPrices
		}
	}

	seed := seedAgentDefaults[level]

	return &output_itf.AgentDefault{Model: seed.Model, ThinkingLevel: seed.ThinkingLevel}
}

func cloneAgentDefault(stored *output_itf.AgentDefault) *output_itf.AgentDefault {
	return &output_itf.AgentDefault{
		Model:         stored.Model,
		ThinkingLevel: stored.ThinkingLevel,
		Prices:        clonePrices(stored.Prices),
	}
}

func clonePrices(prices *output_itf.TokenPrices) *output_itf.TokenPrices {
	if prices == nil {
		return nil
	}

	return &output_itf.TokenPrices{
		Input:       clonePrice(prices.Input),
		CachedInput: clonePrice(prices.CachedInput),
		Output:      clonePrice(prices.Output),
	}
}

func clonePrice(price *float64) *float64 {
	if price == nil {
		return nil
	}

	copied := *price

	return &copied
}

func (c *v1) AgentDefaults() map[enums.TaskLevel]*output_itf.AgentDefault {
	c.mu.RLock()
	defer c.mu.RUnlock()

	defaults := make(map[enums.TaskLevel]*output_itf.AgentDefault, len(c.cfg.AgentDefaults))

	for level, stored := range c.cfg.AgentDefaults {
		defaults[level] = cloneAgentDefault(stored)
	}

	return defaults
}

func (c *v1) AgentDefault(level enums.TaskLevel) (*output_itf.AgentDefault, error) {
	if !level.Valid() {
		return nil, custom_error.Critical("unknown task level %q", level.String())
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	stored := c.cfg.AgentDefaults[level]
	if stored == nil {
		return nil, custom_error.Critical("task level %q has no agent default", level.String())
	}

	return cloneAgentDefault(stored), nil
}

func (c *v1) SetAgentDefault(level enums.TaskLevel, agentDefault *output_itf.AgentDefault) error {
	if !level.Valid() {
		return custom_error.Critical("unknown task level %q", level.String())
	}

	if agentDefault == nil {
		return custom_error.Critical("agent default for %q is empty", level.String())
	}

	next := &output_itf.AgentDefault{
		Model:         agentDefault.Model,
		ThinkingLevel: agentDefault.ThinkingLevel,
	}

	if err := helpers.ValidateStruct(next); err != nil {
		return custom_error.Critical("invalid agent default for %q: %v", level.String(), err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	previous := c.cfg.AgentDefaults[level]
	if previous != nil {
		next.Prices = clonePrices(previous.Prices)
	}

	c.cfg.AgentDefaults[level] = next

	if err := c.write(); err != nil {
		c.cfg.AgentDefaults[level] = previous
		return err
	}

	return nil
}

func (c *v1) SetAgentDefaultPrices(level enums.TaskLevel, prices *output_itf.TokenPrices) error {
	if !level.Valid() {
		return custom_error.Critical("unknown task level %q", level.String())
	}

	if prices != nil {
		if err := helpers.ValidateStruct(prices); err != nil {
			return custom_error.Critical("invalid prices for %q: %v", level.String(), err)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	previous := c.cfg.AgentDefaults[level]
	if previous == nil {
		return custom_error.Critical("task level %q has no agent default", level.String())
	}

	c.cfg.AgentDefaults[level] = &output_itf.AgentDefault{
		Model:         previous.Model,
		ThinkingLevel: previous.ThinkingLevel,
		Prices:        clonePrices(prices),
	}

	if err := c.write(); err != nil {
		c.cfg.AgentDefaults[level] = previous
		return err
	}

	return nil
}

func (c *v1) Onboarded() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cfg.Onboarded
}

func (c *v1) CompleteOnboarding() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cfg.Onboarded {
		return nil
	}

	c.cfg.Onboarded = true

	if err := c.write(); err != nil {
		c.cfg.Onboarded = false
		return err
	}

	return nil
}

func (c *v1) Autopilot() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cfg.Autopilot
}

func (c *v1) SetAutopilot(on bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cfg.Autopilot == on {
		return nil
	}

	previous := c.cfg.Autopilot
	c.cfg.Autopilot = on

	if err := c.write(); err != nil {
		c.cfg.Autopilot = previous
		return err
	}

	return nil
}

func (c *v1) write() error {
	raw, err := json.MarshalIndent(c.cfg, "", "  ")
	if err != nil {
		return custom_error.Critical("cannot encode user config: %v", err)
	}

	tmp := c.path + ".tmp"

	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return custom_error.Critical("cannot write user config %s: %v", filepath.Base(c.path), err)
	}

	if err := os.Rename(tmp, c.path); err != nil {
		os.Remove(tmp)
		return custom_error.Critical("cannot save user config %s: %v", filepath.Base(c.path), err)
	}

	return nil
}
