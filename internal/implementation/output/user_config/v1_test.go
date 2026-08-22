package user_config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"
)

func price(v float64) *float64 {
	return &v
}

func effortDefault(model enums.ModelName, thinking enums.ThinkingLevel) *input_itf.EffortDefault {
	return &input_itf.EffortDefault{Model: model, ThinkingLevel: thinking}
}

func shippedDefaults() []*input_itf.HarnessDefaults {
	return []*input_itf.HarnessDefaults{
		{
			Harness: enums.ClaudeCode,
			Models: map[enums.Effort]*input_itf.EffortDefault{
				enums.EffortQuick:      effortDefault(enums.Haiku, enums.LowThinking),
				enums.EffortStandard:   effortDefault(enums.Sonnet, enums.MedThinking),
				enums.EffortDeep:       effortDefault(enums.Opus, enums.HighThinking),
				enums.EffortExhaustive: effortDefault(enums.Fable, enums.MaxThinking),
			},
		},
		{
			Harness: enums.Codex,
			Models: map[enums.Effort]*input_itf.EffortDefault{
				enums.EffortQuick:      effortDefault(enums.GPT56Luna, enums.LowThinking),
				enums.EffortStandard:   effortDefault(enums.GPT56Luna, enums.MedThinking),
				enums.EffortDeep:       effortDefault(enums.GPT56Terra, enums.HighThinking),
				enums.EffortExhaustive: effortDefault(enums.GPT56Sol, enums.MaxThinking),
			},
		},
		{
			Harness: enums.OpenCode,
			Models: map[enums.Effort]*input_itf.EffortDefault{
				enums.EffortQuick:      effortDefault(enums.Deepseek4Flash, enums.LowThinking),
				enums.EffortStandard:   effortDefault(enums.Deepseek4Flash, enums.MedThinking),
				enums.EffortDeep:       effortDefault(enums.Deepseek4Flash, enums.HighThinking),
				enums.EffortExhaustive: effortDefault(enums.Deepseek4Flash, enums.MaxThinking),
			},
		},
	}
}

var enabledModels = map[enums.AgentHarness][]enums.ModelName{
	enums.ClaudeCode: {enums.Fable, enums.Opus, enums.Sonnet, enums.Haiku},
	enums.Codex:      {enums.GPT56Sol, enums.GPT56Terra, enums.GPT56Luna},
	enums.OpenCode:   {enums.Deepseek4Flash},
}

func readyFrom(harnesses ...enums.AgentHarness) output_itf.ModelReady {
	return func(model enums.ModelName) bool {
		for _, harness := range harnesses {
			if slices.Contains(enabledModels[harness], model) {
				return true
			}
		}

		return false
	}
}

func openConfig(t *testing.T, path string) output_itf.UserConfig {
	t.Helper()

	return openConfigLoggedInto(t, path, enums.ClaudeCode)
}

func openConfigLoggedInto(t *testing.T, path string, harnesses ...enums.AgentHarness) output_itf.UserConfig {
	t.Helper()

	cfg, err := InitV1(path, shippedDefaults(), readyFrom(harnesses...), 3)
	if err != nil {
		t.Fatalf("open user config: %v", err)
	}

	return cfg
}

func agentDefault(t *testing.T, cfg output_itf.UserConfig, level enums.Effort) *output_itf.AgentDefault {
	t.Helper()

	stored, err := cfg.AgentDefault(level)
	if err != nil {
		t.Fatalf("read agent default for %s: %v", level, err)
	}

	return stored
}

func wantEveryLevel(t *testing.T, cfg output_itf.UserConfig, want map[enums.Effort]*output_itf.AgentDefault) {
	t.Helper()

	resolved := cfg.AgentDefaults()

	if len(resolved) != len(want) {
		t.Fatalf("resolved %d levels, want %d: %+v", len(resolved), len(want), resolved)
	}

	for level, expected := range want {
		got := resolved[level]

		if got == nil {
			t.Fatalf("level %s resolved to nothing", level)
		}

		if got.Model != expected.Model || got.ThinkingLevel != expected.ThinkingLevel {
			t.Fatalf("level %s resolved to %+v, want %+v", level, got, expected)
		}
	}
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

func TestSetAgentDefaultLeavesModelPricesAlone(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	setPrices(t, cfg, enums.Sonnet, &output_itf.TokenPrices{Input: price(3), Output: price(15)})

	if err := cfg.SetAgentDefault(enums.EffortStandard, &output_itf.AgentDefault{
		Model:         enums.Opus,
		ThinkingLevel: enums.HighThinking,
	}); err != nil {
		t.Fatalf("set agent default: %v", err)
	}

	reopened := openConfig(t, path)

	if stored := agentDefault(t, reopened, enums.EffortStandard); stored.Model != enums.Opus {
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
    "standard": {
      "model": "opus",
      "thinking_level": "high"
    },
    "deep": {
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

	daily := agentDefault(t, cfg, enums.EffortStandard)

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
    "standard": {
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

	if stored := agentDefault(t, cfg, enums.EffortStandard); stored.Model != enums.Opus {
		t.Fatalf("an unusable price reset the agent default: %+v", stored)
	}

	if stored := cfg.ModelPrice(enums.Opus); stored != nil {
		t.Fatalf("an unusable price was kept as %+v", stored)
	}

	if priced := pricedModels(cfg); len(priced) != 0 {
		t.Fatalf("a price under an unknown model was kept for %v", priced)
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

func TestAgentDefaultsSurviveTheEffortRename(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	seeded := `{"agent_defaults":{"daily_task":{"model":"opus","thinking_level":"xhigh"}}}`
	if err := os.WriteFile(path, []byte(seeded), 0o644); err != nil {
		t.Fatalf("cannot seed the config: %v", err)
	}

	carried := openConfig(t, path).AgentDefaults()[enums.EffortStandard]
	if carried == nil {
		t.Fatal("the default stored under daily_task did not carry over to standard")
	}

	if carried.Model != enums.Opus || carried.ThinkingLevel != enums.XHighThinking {
		t.Fatalf("carried default = %+v, want the stored one back", carried)
	}
}

func TestNoRunnableModelLeavesEveryLevelBlank(t *testing.T) {
	cfg := openConfigLoggedInto(t, filepath.Join(t.TempDir(), "config.json"))

	if resolved := cfg.AgentDefaults(); len(resolved) != 0 {
		t.Fatalf("levels resolved with no tool logged in: %+v", resolved)
	}

	if _, err := cfg.AgentDefault(enums.EffortStandard); err == nil {
		t.Fatal("a level answered with a model no logged-in tool can run")
	}
}

func TestClaudeCodeAloneFillsEveryLevel(t *testing.T) {
	cfg := openConfigLoggedInto(t, filepath.Join(t.TempDir(), "config.json"), enums.ClaudeCode)

	wantEveryLevel(t, cfg, map[enums.Effort]*output_itf.AgentDefault{
		enums.EffortQuick:      {Model: enums.Haiku, ThinkingLevel: enums.LowThinking},
		enums.EffortStandard:   {Model: enums.Sonnet, ThinkingLevel: enums.MedThinking},
		enums.EffortDeep:       {Model: enums.Opus, ThinkingLevel: enums.HighThinking},
		enums.EffortExhaustive: {Model: enums.Fable, ThinkingLevel: enums.MaxThinking},
	})
}

func TestCodexOutranksOpenCodeBecauseItComesFirstInTheTable(t *testing.T) {
	cfg := openConfigLoggedInto(t, filepath.Join(t.TempDir(), "config.json"), enums.OpenCode, enums.Codex)

	wantEveryLevel(t, cfg, map[enums.Effort]*output_itf.AgentDefault{
		enums.EffortQuick:      {Model: enums.GPT56Luna, ThinkingLevel: enums.LowThinking},
		enums.EffortStandard:   {Model: enums.GPT56Luna, ThinkingLevel: enums.MedThinking},
		enums.EffortDeep:       {Model: enums.GPT56Terra, ThinkingLevel: enums.HighThinking},
		enums.EffortExhaustive: {Model: enums.GPT56Sol, ThinkingLevel: enums.MaxThinking},
	})
}

func TestOpenCodeAloneRunsEveryLevelOnItsOneFreeModel(t *testing.T) {
	cfg := openConfigLoggedInto(t, filepath.Join(t.TempDir(), "config.json"), enums.OpenCode)

	wantEveryLevel(t, cfg, map[enums.Effort]*output_itf.AgentDefault{
		enums.EffortQuick:      {Model: enums.Deepseek4Flash, ThinkingLevel: enums.LowThinking},
		enums.EffortStandard:   {Model: enums.Deepseek4Flash, ThinkingLevel: enums.MedThinking},
		enums.EffortDeep:       {Model: enums.Deepseek4Flash, ThinkingLevel: enums.HighThinking},
		enums.EffortExhaustive: {Model: enums.Deepseek4Flash, ThinkingLevel: enums.MaxThinking},
	})
}

func TestAPickedModelOutranksTheTableAndSurvivesAReopen(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfigLoggedInto(t, path, enums.ClaudeCode, enums.Codex)

	if err := cfg.SetAgentDefault(enums.EffortQuick, &output_itf.AgentDefault{
		Model:         enums.GPT56Luna,
		ThinkingLevel: enums.HighThinking,
	}); err != nil {
		t.Fatalf("pick a model for quick: %v", err)
	}

	reopened := openConfigLoggedInto(t, path, enums.ClaudeCode, enums.Codex)

	picked := agentDefault(t, reopened, enums.EffortQuick)

	if picked.Model != enums.GPT56Luna || picked.ThinkingLevel != enums.HighThinking {
		t.Fatalf("the table overrode the user's pick: %+v", picked)
	}

	if untouched := agentDefault(t, reopened, enums.EffortStandard); untouched.Model != enums.Sonnet {
		t.Fatalf("picking one level moved another: %+v", untouched)
	}
}

func TestAPickedModelStopsResolvingOnceItsToolIsLoggedOut(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfigLoggedInto(t, path, enums.ClaudeCode, enums.Codex)

	if err := cfg.SetAgentDefault(enums.EffortQuick, &output_itf.AgentDefault{
		Model:         enums.GPT56Luna,
		ThinkingLevel: enums.HighThinking,
	}); err != nil {
		t.Fatalf("pick a model for quick: %v", err)
	}

	quick := agentDefault(t, openConfigLoggedInto(t, path, enums.ClaudeCode), enums.EffortQuick)

	if quick.Model != enums.Haiku || quick.ThinkingLevel != enums.LowThinking {
		t.Fatalf("quick answered %+v, want the table entry of the one logged-in tool", quick)
	}
}

func TestTheCurrentEffortNameWinsOverItsRetiredSpelling(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	seeded := `{"agent_defaults":{` +
		`"daily_task":{"model":"haiku","thinking_level":"low"},` +
		`"standard":{"model":"opus","thinking_level":"high"}}}`

	for attempt := range 32 {
		if err := os.WriteFile(path, []byte(seeded), 0o600); err != nil {
			t.Fatalf("cannot seed the config: %v", err)
		}

		standard := agentDefault(t, openConfig(t, path), enums.EffortStandard)

		if standard.Model != enums.Opus || standard.ThinkingLevel != enums.HighThinking {
			t.Fatalf("attempt %d: daily_task overwrote standard with %+v", attempt, standard)
		}
	}
}

func TestALevelAToolDoesNotOfferFallsThroughToTheNextTool(t *testing.T) {
	defaults := []*input_itf.HarnessDefaults{
		{
			Harness: enums.Codex,
			Models: map[enums.Effort]*input_itf.EffortDefault{
				enums.EffortQuick:      effortDefault(enums.GPT56Luna, enums.LowThinking),
				enums.EffortStandard:   effortDefault(enums.GPT56Luna, enums.MedThinking),
				enums.EffortExhaustive: effortDefault(enums.GPT56Sol, enums.MaxThinking),
			},
		},
		{
			Harness: enums.OpenCode,
			Models: map[enums.Effort]*input_itf.EffortDefault{
				enums.EffortQuick:      effortDefault(enums.Deepseek4Flash, enums.LowThinking),
				enums.EffortStandard:   effortDefault(enums.Deepseek4Flash, enums.MedThinking),
				enums.EffortDeep:       effortDefault(enums.Deepseek4Flash, enums.HighThinking),
				enums.EffortExhaustive: effortDefault(enums.Deepseek4Flash, enums.MaxThinking),
			},
		},
	}

	cfg, err := InitV1(filepath.Join(t.TempDir(), "config.json"), defaults, readyFrom(enums.Codex, enums.OpenCode), 3)
	if err != nil {
		t.Fatalf("open user config: %v", err)
	}

	if deep := agentDefault(t, cfg, enums.EffortDeep); deep.Model != enums.Deepseek4Flash {
		t.Fatalf("the level Codex does not list resolved to %+v", deep)
	}

	if quick := agentDefault(t, cfg, enums.EffortQuick); quick.Model != enums.GPT56Luna {
		t.Fatalf("a gap at one level moved the levels Codex does list: %+v", quick)
	}
}

func TestInitRefusesATableItCannotResolveFrom(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if _, err := InitV1(path, nil, readyFrom(enums.ClaudeCode), 3); err == nil {
		t.Fatal("an empty defaults table was accepted")
	}

	if _, err := InitV1(path, shippedDefaults(), nil, 3); err == nil {
		t.Fatal("a missing readiness check was accepted")
	}

	if _, err := InitV1(path, shippedDefaults(), readyFrom(enums.ClaudeCode), 0); err == nil {
		t.Fatal("a default of no running agents was accepted")
	}
}

func TestLanguageRoundTripsThroughTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := openConfig(t, path).SetLanguage(enums.Vietnamese); err != nil {
		t.Fatalf("set language: %v", err)
	}

	if spoken := openConfig(t, path).Language(); spoken != enums.Vietnamese {
		t.Fatalf("reopened config speaks %s, want %s", spoken, enums.Vietnamese)
	}
}

func TestAConfigWrittenBeforeLanguageExistedSpeaksEnglish(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	old := `{"onboarded": true, "autopilot": true}`

	if err := os.WriteFile(path, []byte(old), 0o600); err != nil {
		t.Fatalf("write old config: %v", err)
	}

	if spoken := openConfig(t, path).Language(); spoken != enums.English {
		t.Fatalf("a config with no language key speaks %s, want %s", spoken, enums.English)
	}
}

func TestAnUnknownLanguageIsRejected(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	cfg := openConfig(t, path)

	if err := cfg.SetLanguage("kl"); err == nil {
		t.Fatal("a language the app does not speak was accepted")
	}

	if spoken := cfg.Language(); spoken != enums.English {
		t.Fatalf("a refused language changed the config to %s", spoken)
	}
}

func TestMaxRunningAgentsFallsBackToTheShippedDefault(t *testing.T) {
	if limit := openConfig(t, filepath.Join(t.TempDir(), "config.json")).MaxRunningAgents(); limit != 3 {
		t.Fatalf("a config with no stored limit allows %d agents, want the shipped 3", limit)
	}
}

func TestMaxRunningAgentsRoundTripsThroughTheFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")

	if err := openConfig(t, path).SetMaxRunningAgents(7); err != nil {
		t.Fatalf("set max running agents: %v", err)
	}

	if limit := openConfig(t, path).MaxRunningAgents(); limit != 7 {
		t.Fatalf("reopened config allows %d agents, want 7", limit)
	}
}

func TestAMaxRunningAgentsBelowOneIsRejected(t *testing.T) {
	cfg := openConfig(t, filepath.Join(t.TempDir(), "config.json"))

	if err := cfg.SetMaxRunningAgents(0); err == nil {
		t.Fatal("a limit of no running agents was accepted")
	}

	if limit := cfg.MaxRunningAgents(); limit != 3 {
		t.Fatalf("a refused limit changed the config to %d", limit)
	}
}
