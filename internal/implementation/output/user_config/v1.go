package user_config

import (
	"encoding/json"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"
)

type file struct {
	AgentDefaults    map[enums.Effort]*output_itf.AgentDefault   `json:"agent_defaults"`
	ModelPrices      map[enums.ModelName]*output_itf.TokenPrices `json:"model_prices"`
	Onboarded        bool                                        `json:"onboarded"`
	Autopilot        bool                                        `json:"autopilot"`
	Language         enums.Language                              `json:"language"`
	MaxRunningAgents int                                         `json:"max_running_agents"`
}

type v1 struct {
	path     string
	defaults []*input_itf.HarnessDefaults
	ready    output_itf.ModelReady
	mu       sync.RWMutex
	cfg      *file
}

func InitV1(
	path string,
	defaults []*input_itf.HarnessDefaults,
	ready output_itf.ModelReady,
	maxRunningAgents int,
) (output_itf.UserConfig, error) {
	if path == "" {
		return nil, custom_error.Critical("user config path is empty")
	}

	if len(defaults) == 0 {
		return nil, custom_error.Critical("no coding tool declares a model per effort level")
	}

	if ready == nil {
		return nil, custom_error.Critical("user config cannot tell which models are runnable")
	}

	if maxRunningAgents < 1 {
		return nil, custom_error.Critical("default max running agents must be at least 1")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, custom_error.Critical("cannot create user config dir: %v", err)
	}

	cfg, err := read(path, maxRunningAgents)
	if err != nil {
		return nil, err
	}

	c := &v1{path: path, defaults: defaults, ready: ready, cfg: cfg}

	if err := c.write(); err != nil {
		return nil, err
	}

	return c, nil
}

func read(path string, maxRunningAgents int) (*file, error) {
	cfg := &file{
		AgentDefaults: map[enums.Effort]*output_itf.AgentDefault{},
		ModelPrices:   map[enums.ModelName]*output_itf.TokenPrices{},
	}

	raw, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, custom_error.Critical("cannot read user config %s: %v", path, err)
	}

	if len(raw) > 0 {
		if err := json.Unmarshal(raw, cfg); err != nil {
			return nil, custom_error.Critical("cannot decode user config %s: %v", path, err)
		}
	}

	if !cfg.Language.Valid() {
		cfg.Language = enums.English
	}

	if cfg.MaxRunningAgents < 1 {
		cfg.MaxRunningAgents = maxRunningAgents
	}

	cfg.AgentDefaults = picked(renamedEfforts(cfg.AgentDefaults))

	cfg.ModelPrices = usablePrices(cfg.ModelPrices)

	return cfg, nil
}

func usablePrices(stored map[enums.ModelName]*output_itf.TokenPrices) map[enums.ModelName]*output_itf.TokenPrices {
	prices := make(map[enums.ModelName]*output_itf.TokenPrices, len(stored))

	for model, price := range stored {
		if model.Valid() && price != nil && helpers.ValidateStruct(price) == nil {
			prices[model] = price
		}
	}

	return prices
}

var retiredEfforts = map[enums.Effort]enums.Effort{
	"lightweight_task":    enums.EffortQuick,
	"daily_task":          enums.EffortStandard,
	"heavy_task":          enums.EffortDeep,
	"maximum_effort_task": enums.EffortExhaustive,
}

func renamedEfforts(stored map[enums.Effort]*output_itf.AgentDefault) map[enums.Effort]*output_itf.AgentDefault {
	carried := make(map[enums.Effort]*output_itf.AgentDefault, len(stored))

	for level, value := range stored {
		if _, retired := retiredEfforts[level]; !retired {
			carried[level] = value
		}
	}

	for level, value := range stored {
		if current, retired := retiredEfforts[level]; retired && carried[current] == nil {
			carried[current] = value
		}
	}

	return carried
}

func picked(stored map[enums.Effort]*output_itf.AgentDefault) map[enums.Effort]*output_itf.AgentDefault {
	kept := make(map[enums.Effort]*output_itf.AgentDefault, len(stored))

	for _, level := range enums.Efforts() {
		if choice := stored[level]; choice != nil && helpers.ValidateStruct(choice) == nil {
			kept[level] = cloneAgentDefault(choice)
		}
	}

	return kept
}

func cloneAgentDefault(stored *output_itf.AgentDefault) *output_itf.AgentDefault {
	return &output_itf.AgentDefault{
		Model:         stored.Model,
		ThinkingLevel: stored.ThinkingLevel,
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

func (c *v1) resolve(level enums.Effort) *output_itf.AgentDefault {
	if choice := c.cfg.AgentDefaults[level]; choice != nil && c.ready(choice.Model) {
		return cloneAgentDefault(choice)
	}

	for _, harness := range c.defaults {
		if offered := harness.Models[level]; offered != nil && c.ready(offered.Model) {
			return &output_itf.AgentDefault{Model: offered.Model, ThinkingLevel: offered.ThinkingLevel}
		}
	}

	return nil
}

func (c *v1) toolNames() string {
	names := make([]string, 0, len(c.defaults))

	for _, harness := range c.defaults {
		names = append(names, harness.Harness.DisplayName())
	}

	return strings.Join(names, ", ")
}

func (c *v1) AgentDefaults() map[enums.Effort]*output_itf.AgentDefault {
	c.mu.RLock()
	defer c.mu.RUnlock()

	defaults := make(map[enums.Effort]*output_itf.AgentDefault, len(enums.Efforts()))

	for _, level := range enums.Efforts() {
		if resolved := c.resolve(level); resolved != nil {
			defaults[level] = resolved
		}
	}

	return defaults
}

func (c *v1) AgentDefault(level enums.Effort) (*output_itf.AgentDefault, error) {
	if !level.Valid() {
		return nil, custom_error.Critical("unknown step level %q", level.String())
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	resolved := c.resolve(level)
	if resolved == nil {
		return nil, custom_error.Critical(
			"no model runs %q steps yet: log in to %s",
			level.String(),
			c.toolNames(),
		)
	}

	return resolved, nil
}

func (c *v1) SetAgentDefault(level enums.Effort, agentDefault *output_itf.AgentDefault) error {
	if !level.Valid() {
		return custom_error.Critical("unknown step level %q", level.String())
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
	c.cfg.AgentDefaults[level] = next

	if err := c.write(); err != nil {
		c.cfg.AgentDefaults[level] = previous
		return err
	}

	return nil
}

func (c *v1) ModelPrice(model enums.ModelName) *output_itf.TokenPrices {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return clonePrices(c.cfg.ModelPrices[model])
}

func (c *v1) SetModelPrices(model enums.ModelName, prices *output_itf.TokenPrices) error {
	if !model.Valid() {
		return custom_error.Critical("unknown model %q", model.String())
	}

	if prices != nil {
		if err := helpers.ValidateStruct(prices); err != nil {
			return custom_error.Critical("invalid prices for %q: %v", model.String(), err)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	previous := maps.Clone(c.cfg.ModelPrices)
	delete(c.cfg.ModelPrices, model)

	if prices != nil {
		c.cfg.ModelPrices[model] = clonePrices(prices)
	}

	if err := c.write(); err != nil {
		c.cfg.ModelPrices = previous
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

func (c *v1) MaxRunningAgents() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.cfg.MaxRunningAgents
}

func (c *v1) SetMaxRunningAgents(limit int) error {
	if limit < 1 {
		return custom_error.Critical("at least one agent has to be allowed to run")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cfg.MaxRunningAgents == limit {
		return nil
	}

	previous := c.cfg.MaxRunningAgents
	c.cfg.MaxRunningAgents = limit

	if err := c.write(); err != nil {
		c.cfg.MaxRunningAgents = previous
		return err
	}

	return nil
}

func (c *v1) Language() enums.Language {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if !c.cfg.Language.Valid() {
		return enums.English
	}

	return c.cfg.Language
}

func (c *v1) SetLanguage(language enums.Language) error {
	if !language.Valid() {
		return custom_error.Critical("%s is not a language this app speaks", language)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cfg.Language == language {
		return nil
	}

	previous := c.cfg.Language
	c.cfg.Language = language

	if err := c.write(); err != nil {
		c.cfg.Language = previous
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
