// Package pricing turns token counts into US dollars. It knows nothing about where
// a rate was configured or where a count came from, so the arithmetic that decides
// what a run cost can be read and tested on its own.
package pricing

// Rates are US dollars per million tokens.
type Rates struct {
	Input       float64
	CachedInput float64
	Output      float64
}

// Tokens is what one agent spent: Input is prompt read at full price, Cached is
// prompt read back out of the vendor's cache, Output is what the agent wrote.
type Tokens struct {
	Input  int
	Cached int
	Output int
}

const perMillion = 1_000_000

// NewRates reads the three optional prices a model carries. A model is priced only
// once it names both an input and an output rate — a run half-priced would report a
// number lower than the truth, which is worse than reporting none. The cached rate is
// allowed to be missing and falls back to the input rate.
func NewRates(input, cachedInput, output *float64) (Rates, bool) {
	if input == nil || output == nil || *input < 0 || *output < 0 {
		return Rates{}, false
	}

	cached := *input
	if cachedInput != nil && *cachedInput >= 0 {
		cached = *cachedInput
	}

	return Rates{Input: *input, CachedInput: cached, Output: *output}, true
}

// Cost is what these tokens cost at these rates.
func (r Rates) Cost(t Tokens) float64 {
	dollars := float64(t.Input)*r.Input +
		float64(t.Cached)*r.CachedInput +
		float64(t.Output)*r.Output

	return dollars / perMillion
}
