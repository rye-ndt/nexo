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

func pricedModels(cfg output_itf.UserConfig) []enums.ModelName {
	priced := []enums.ModelName{}

	for _, model := range enums.ModelNames() {
		if cfg.ModelPrice(model) != nil {
			priced = append(priced, model)
		}
	}

	return priced
}

func setPrices(t *testing.T, cfg output_itf.UserConfig, model enums.ModelName, prices *output_itf.TokenPrices) {
	t.Helper()

	if err := cfg.SetModelPrices(model, prices); err != nil {
		t.Fatalf("set prices for %s: %v", model, err)
	}
}

func TestPricesRoundTripThroughTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	setPrices(t, openConfig(t, path), enums.Opus, &output_itf.TokenPrices{
		Input:       price(15),
		CachedInput: price(1.5),
		Output:      price(75),
	})

	reopened := openConfig(t, path).ModelPrice(enums.Opus)

	if reopened == nil {
		t.Fatal("prices are gone after reopening the config")
	}

	if *reopened.Input != 15 || *reopened.CachedInput != 1.5 || *reopened.Output != 75 {
		t.Fatalf("prices came back as %+v", reopened)
	}
}

func TestPricesAreReadBackByModelPrices(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	setPrices(t, cfg, enums.Sonnet, &output_itf.TokenPrices{
		Input:  price(3),
		Output: price(15),
	})

	stored := cfg.ModelPrice(enums.Sonnet)
	if stored == nil {
		t.Fatal("model prices dropped the prices")
	}

	if *stored.Input != 3 || *stored.Output != 15 {
		t.Fatalf("prices came back as %+v", stored)
	}

	if stored.CachedInput != nil {
		t.Fatalf("blank cached price came back as %v", *stored.CachedInput)
	}
}

func TestZeroPriceIsNotABlankPrice(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	setPrices(t, openConfig(t, path), enums.Deepseek4Flash, &output_itf.TokenPrices{
		Input:  price(0),
		Output: price(0),
	})

	stored := openConfig(t, path).ModelPrice(enums.Deepseek4Flash)

	if stored == nil {
		t.Fatal("a free model's prices were stored as blank")
	}

	if stored.Input == nil || *stored.Input != 0 {
		t.Fatalf("input price came back as %v, want a set 0", stored.Input)
	}

	if stored.Output == nil || *stored.Output != 0 {
		t.Fatalf("output price came back as %v, want a set 0", stored.Output)
	}

	if stored.CachedInput != nil {
		t.Fatalf("cached price came back as %v, want blank", *stored.CachedInput)
	}
}

func TestOneModelIsPricedTheSameAtEveryTaskLevel(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	for _, level := range []enums.TaskLevel{enums.HeavyTask, enums.MaximumEffortTask} {
		if err := cfg.SetAgentDefault(level, &output_itf.AgentDefault{
			Model:         enums.Opus,
			ThinkingLevel: enums.HighThinking,
		}); err != nil {
			t.Fatalf("set agent default for %s: %v", level, err)
		}
	}

	setPrices(t, cfg, enums.Opus, &output_itf.TokenPrices{Input: price(15), Output: price(75)})

	heavy := agentDefault(t, cfg, enums.HeavyTask)
	maximum := agentDefault(t, cfg, enums.MaximumEffortTask)

	if cfg.ModelPrice(heavy.Model) == nil || cfg.ModelPrice(maximum.Model) == nil {
		t.Fatal("two levels on one model do not share its prices")
	}
}

func TestSetAgentDefaultLeavesModelPricesAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	setPrices(t, cfg, enums.Sonnet, &output_itf.TokenPrices{Input: price(3), Output: price(15)})

	if err := cfg.SetAgentDefault(enums.DailyTask, &output_itf.AgentDefault{
		Model:         enums.Opus,
		ThinkingLevel: enums.HighThinking,
	}); err != nil {
		t.Fatalf("set agent default: %v", err)
	}

	reopened := openConfig(t, path)

	if stored := agentDefault(t, reopened, enums.DailyTask); stored.Model != enums.Opus {
		t.Fatalf("model change did not stick: %+v", stored)
	}

	stored := reopened.ModelPrice(enums.Sonnet)
	if stored == nil || *stored.Input != 3 || *stored.Output != 15 {
		t.Fatalf("changing a level's model wiped the model's prices: %+v", stored)
	}
}

func TestNilPricesClearStoredPrices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	setPrices(t, cfg, enums.Fable, &output_itf.TokenPrices{Input: price(15), Output: price(75)})
	setPrices(t, cfg, enums.Fable, nil)

	if stored := cfg.ModelPrice(enums.Fable); stored != nil {
		t.Fatalf("prices are still %+v after clearing them", stored)
	}

	if stored := openConfig(t, path).ModelPrice(enums.Fable); stored != nil {
		t.Fatalf("cleared prices came back from the file as %+v", stored)
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

	if priced := pricedModels(cfg); len(priced) != 0 {
		t.Fatalf("a config with no prices key loaded prices for %v", priced)
	}

	if !cfg.Onboarded() || !cfg.Autopilot() {
		t.Fatal("upgrade lost the other stored settings")
	}

	setPrices(t, cfg, enums.Opus, &output_itf.TokenPrices{Input: price(3), Output: price(15)})

	if stored := openConfig(t, path).ModelPrice(enums.Opus); stored == nil {
		t.Fatal("prices set on an upgraded config did not persist")
	}
}

func TestUnreadablePricesKeepTheStoredModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	broken := `{
  "agent_defaults": {
    "daily_task": {
      "model": "opus",
      "thinking_level": "high"
    }
  },
  "model_prices": {
    "opus": {"input_per_mtok": -4},
    "bogus": {"input_per_mtok": 1}
  }
}`

	if err := os.WriteFile(path, []byte(broken), 0o600); err != nil {
		t.Fatalf("write broken config: %v", err)
	}

	cfg := openConfig(t, path)

	if stored := agentDefault(t, cfg, enums.DailyTask); stored.Model != enums.Opus {
		t.Fatalf("an unusable price reset the agent default: %+v", stored)
	}

	if stored := cfg.ModelPrice(enums.Opus); stored != nil {
		t.Fatalf("an unusable price was kept as %+v", stored)
	}

	if priced := pricedModels(cfg); len(priced) != 0 {
		t.Fatalf("a price under an unknown model was kept for %v", priced)
	}
}

func TestFreshInstallPricesNothing(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	if priced := pricedModels(cfg); len(priced) != 0 {
		t.Fatalf("fresh install priced %v", priced)
	}
}

func TestReturnedPricesCannotBeMutatedByTheCaller(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	setPrices(t, cfg, enums.Sonnet, &output_itf.TokenPrices{Input: price(3), Output: price(15)})

	handed := cfg.ModelPrice(enums.Sonnet)
	*handed.Input = 999
	handed.Output = price(0)

	stored := cfg.ModelPrice(enums.Sonnet)

	if *stored.Input != 3 || *stored.Output != 15 {
		t.Fatalf("a caller mutated the stored prices: %+v", stored)
	}
}

func TestSetPricesOnUnknownModelFails(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	if err := cfg.SetModelPrices(enums.ModelName("bogus"), nil); err == nil {
		t.Fatal("an unknown model was accepted")
	}
}

func TestNegativePriceIsRejected(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	if err := cfg.SetModelPrices(enums.Sonnet, &output_itf.TokenPrices{
		Input:  price(-1),
		Output: price(15),
	}); err == nil {
		t.Fatal("a negative price was accepted")
	}

	if stored := cfg.ModelPrice(enums.Sonnet); stored != nil {
		t.Fatalf("a rejected write left prices %+v behind", stored)
	}
}
