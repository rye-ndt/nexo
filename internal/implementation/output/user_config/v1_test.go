package user_config

import (
	"os"
	"path/filepath"
	"testing"

	"hexago/internal/helpers/enums"
	output_itf "hexago/internal/interface/output"
)

func price(v float64) *float64 {
	return &v
}

func openConfig(t *testing.T, path string) output_itf.UserConfig {
	t.Helper()

	cfg, err := InitV1(path)
	if err != nil {
		t.Fatalf("open user config: %v", err)
	}

	return cfg
}

func agentDefault(t *testing.T, cfg output_itf.UserConfig, level enums.TaskLevel) *output_itf.AgentDefault {
	t.Helper()

	stored, err := cfg.AgentDefault(level)
	if err != nil {
		t.Fatalf("read agent default for %s: %v", level, err)
	}

	return stored
}

func TestPricesRoundTripThroughTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.HeavyTask, &output_itf.TokenPrices{
		Input:       price(15),
		CachedInput: price(1.5),
		Output:      price(75),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	reopened := agentDefault(t, openConfig(t, path), enums.HeavyTask)

	if reopened.Prices == nil {
		t.Fatal("prices are gone after reopening the config")
	}

	if *reopened.Prices.Input != 15 || *reopened.Prices.CachedInput != 1.5 || *reopened.Prices.Output != 75 {
		t.Fatalf("prices came back as %+v", reopened.Prices)
	}
}

func TestPricesAreReadBackByAgentDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.DailyTask, &output_itf.TokenPrices{
		Input:  price(3),
		Output: price(15),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	stored := cfg.AgentDefaults()[enums.DailyTask]
	if stored == nil || stored.Prices == nil {
		t.Fatal("agent defaults dropped the prices")
	}

	if *stored.Prices.Input != 3 || *stored.Prices.Output != 15 {
		t.Fatalf("prices came back as %+v", stored.Prices)
	}

	if stored.Prices.CachedInput != nil {
		t.Fatalf("blank cached price came back as %v", *stored.Prices.CachedInput)
	}
}

func TestZeroPriceIsNotABlankPrice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.LightweightTask, &output_itf.TokenPrices{
		Input:  price(0),
		Output: price(0),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	stored := agentDefault(t, openConfig(t, path), enums.LightweightTask)

	if stored.Prices == nil {
		t.Fatal("a free model's prices were stored as blank")
	}

	if stored.Prices.Input == nil || *stored.Prices.Input != 0 {
		t.Fatalf("input price came back as %v, want a set 0", stored.Prices.Input)
	}

	if stored.Prices.Output == nil || *stored.Prices.Output != 0 {
		t.Fatalf("output price came back as %v, want a set 0", stored.Prices.Output)
	}

	if stored.Prices.CachedInput != nil {
		t.Fatalf("cached price came back as %v, want blank", *stored.Prices.CachedInput)
	}
}

func TestSetAgentDefaultKeepsStoredPrices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.DailyTask, &output_itf.TokenPrices{
		Input:  price(3),
		Output: price(15),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	if err := cfg.SetAgentDefault(enums.DailyTask, &output_itf.AgentDefault{
		Model:         enums.Opus,
		ThinkingLevel: enums.HighThinking,
	}); err != nil {
		t.Fatalf("set agent default: %v", err)
	}

	stored := agentDefault(t, openConfig(t, path), enums.DailyTask)

	if stored.Model != enums.Opus || stored.ThinkingLevel != enums.HighThinking {
		t.Fatalf("model change did not stick: %+v", stored)
	}

	if stored.Prices == nil || *stored.Prices.Input != 3 || *stored.Prices.Output != 15 {
		t.Fatalf("changing the model wiped the prices: %+v", stored.Prices)
	}
}

func TestSetAgentDefaultPricesKeepsStoredModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefault(enums.HeavyTask, &output_itf.AgentDefault{
		Model:         enums.Haiku,
		ThinkingLevel: enums.LowThinking,
	}); err != nil {
		t.Fatalf("set agent default: %v", err)
	}

	if err := cfg.SetAgentDefaultPrices(enums.HeavyTask, &output_itf.TokenPrices{
		Input:  price(1),
		Output: price(5),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	stored := agentDefault(t, openConfig(t, path), enums.HeavyTask)

	if stored.Model != enums.Haiku || stored.ThinkingLevel != enums.LowThinking {
		t.Fatalf("setting prices changed the model: %+v", stored)
	}
}

func TestNilPricesClearStoredPrices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.MaximumEffortTask, &output_itf.TokenPrices{
		Input:  price(15),
		Output: price(75),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	if err := cfg.SetAgentDefaultPrices(enums.MaximumEffortTask, nil); err != nil {
		t.Fatalf("clear prices: %v", err)
	}

	if stored := agentDefault(t, cfg, enums.MaximumEffortTask); stored.Prices != nil {
		t.Fatalf("prices are still %+v after clearing them", stored.Prices)
	}

	if stored := agentDefault(t, openConfig(t, path), enums.MaximumEffortTask); stored.Prices != nil {
		t.Fatalf("cleared prices came back from the file as %+v", stored.Prices)
	}
}

func TestConfigWrittenBeforePricesExistedStillLoads(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	old := `{
  "agent_defaults": {
    "daily_task": {
      "model": "opus",
      "thinking_level": "high"
    },
    "heavy_task": {
      "model": "fable",
      "thinking_level": "max"
    }
  },
  "onboarded": true,
  "autopilot": true
}`

	if err := os.WriteFile(path, []byte(old), 0o600); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	cfg := openConfig(t, path)

	daily := agentDefault(t, cfg, enums.DailyTask)

	if daily.Model != enums.Opus || daily.ThinkingLevel != enums.HighThinking {
		t.Fatalf("upgrade lost the stored model: %+v", daily)
	}

	if daily.Prices != nil {
		t.Fatalf("a config with no prices key loaded prices %+v", daily.Prices)
	}

	if !cfg.Onboarded() || !cfg.Autopilot() {
		t.Fatal("upgrade lost the other stored settings")
	}

	if err := cfg.SetAgentDefaultPrices(enums.DailyTask, &output_itf.TokenPrices{
		Input:  price(3),
		Output: price(15),
	}); err != nil {
		t.Fatalf("set prices on the upgraded config: %v", err)
	}

	if stored := agentDefault(t, openConfig(t, path), enums.DailyTask); stored.Prices == nil {
		t.Fatal("prices set on an upgraded config did not persist")
	}
}

func TestUnreadablePricesKeepTheStoredModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	broken := `{
  "agent_defaults": {
    "daily_task": {
      "model": "opus",
      "thinking_level": "high",
      "prices": {"input_per_mtok": -4}
    }
  }
}`

	if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
		t.Fatalf("write broken config: %v", err)
	}

	stored := agentDefault(t, openConfig(t, path), enums.DailyTask)

	if stored.Model != enums.Opus || stored.ThinkingLevel != enums.HighThinking {
		t.Fatalf("an unusable price reset the whole default: %+v", stored)
	}

	if stored.Prices != nil {
		t.Fatalf("an unusable price was kept as %+v", stored.Prices)
	}
}

func TestSeededDefaultsCarryNoPrices(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	for level, stored := range cfg.AgentDefaults() {
		if stored.Prices != nil {
			t.Fatalf("fresh install priced %s at %+v", level, stored.Prices)
		}
	}
}

func TestReturnedPricesCannotBeMutatedByTheCaller(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.DailyTask, &output_itf.TokenPrices{
		Input:  price(3),
		Output: price(15),
	}); err != nil {
		t.Fatalf("set prices: %v", err)
	}

	handed := agentDefault(t, cfg, enums.DailyTask)
	*handed.Prices.Input = 999
	handed.Prices.Output = price(0)
	handed.Model = enums.Haiku

	stored := agentDefault(t, cfg, enums.DailyTask)

	if *stored.Prices.Input != 3 || *stored.Prices.Output != 15 {
		t.Fatalf("a caller mutated the stored prices: %+v", stored.Prices)
	}

	if stored.Model != enums.Sonnet {
		t.Fatalf("a caller mutated the stored model: %s", stored.Model)
	}

	fromMap := cfg.AgentDefaults()[enums.DailyTask]
	*fromMap.Prices.Input = 111

	if *agentDefault(t, cfg, enums.DailyTask).Prices.Input != 3 {
		t.Fatal("a caller mutated the stored prices through AgentDefaults")
	}
}

func TestSetPricesOnUnknownTaskLevelFails(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	if err := cfg.SetAgentDefaultPrices(enums.TaskLevel("bogus_task"), nil); err == nil {
		t.Fatal("an unknown task level was accepted")
	}
}

func TestNegativePriceIsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetAgentDefaultPrices(enums.DailyTask, &output_itf.TokenPrices{
		Input:  price(-1),
		Output: price(15),
	}); err == nil {
		t.Fatal("a negative price was accepted")
	}

	if stored := agentDefault(t, cfg, enums.DailyTask); stored.Prices != nil {
		t.Fatalf("a rejected write left prices %+v behind", stored.Prices)
	}
}
