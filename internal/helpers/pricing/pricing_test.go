package pricing

import (
	"math"
	"testing"
)

const epsilon = 1e-9

func price(v float64) *float64 {
	return &v
}

func closeEnough(got, want float64) bool {
	return math.Abs(got-want) < epsilon
}

func mustRates(t *testing.T, input, cachedInput, output *float64) Rates {
	t.Helper()

	rates, ok := NewRates(input, cachedInput, output)
	if !ok {
		t.Fatalf("rates are not priced")
	}

	return rates
}

func TestCostSumsEveryBucketAtItsOwnRate(t *testing.T) {
	rates := mustRates(t, price(3), price(0.3), price(15))

	got := rates.Cost(Tokens{Input: 1_000_000, Cached: 2_000_000, Output: 500_000})

	if !closeEnough(got, 11.10) {
		t.Fatalf("cost is %v, want 11.10", got)
	}
}

func TestCachedRateFallsBackToInputOnlyWhenItIsBlank(t *testing.T) {
	cases := map[string]struct {
		cachedInput *float64
		want        float64
	}{
		"blank falls back to the input rate": {cachedInput: nil, want: 3},
		"zero is a free cache read":          {cachedInput: price(0), want: 0},
		"a rate of its own is used":          {cachedInput: price(0.3), want: 0.3},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rates := mustRates(t, price(3), tc.cachedInput, price(15))

			if !closeEnough(rates.CachedInput, tc.want) {
				t.Fatalf("cached rate is %v, want %v", rates.CachedInput, tc.want)
			}

			if got := rates.Cost(Tokens{Cached: 1_000_000}); !closeEnough(got, tc.want) {
				t.Fatalf("a million cached tokens cost %v, want %v", got, tc.want)
			}
		})
	}
}

func TestALevelIsPricedOnlyWithBothRatesAndNeitherNegative(t *testing.T) {
	cases := map[string]struct {
		input  *float64
		output *float64
		priced bool
	}{
		"no input":        {input: nil, output: price(15)},
		"no output":       {input: price(3), output: nil},
		"neither":         {input: nil, output: nil},
		"negative input":  {input: price(-1), output: price(15)},
		"negative output": {input: price(3), output: price(-15)},
		"free model":      {input: price(0), output: price(0), priced: true},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rates, ok := NewRates(tc.input, price(0.3), tc.output)
			if ok != tc.priced {
				t.Fatalf("priced = %t, want %t", ok, tc.priced)
			}

			if !tc.priced {
				return
			}

			if got := rates.Cost(Tokens{Input: 4_000_000, Output: 2_000_000}); got != 0 {
				t.Fatalf("free model cost %v, want 0", got)
			}
		})
	}
}
