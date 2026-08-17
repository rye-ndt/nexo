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

func TestCachedRateFallsBackToInputWhenBlank(t *testing.T) {
	rates := mustRates(t, price(3), nil, price(15))

	if !closeEnough(rates.CachedInput, 3) {
		t.Fatalf("cached rate is %v, want the input rate 3", rates.CachedInput)
	}

	got := rates.Cost(Tokens{Cached: 1_000_000})

	if !closeEnough(got, 3) {
		t.Fatalf("cached tokens cost %v, want 3", got)
	}
}

func TestCachedRateIsUsedWhenSet(t *testing.T) {
	rates := mustRates(t, price(3), price(0.3), price(15))

	got := rates.Cost(Tokens{Cached: 1_000_000})

	if !closeEnough(got, 0.3) {
		t.Fatalf("cached tokens cost %v, want 0.3", got)
	}
}

func TestCachedRateOfZeroIsNotTreatedAsBlank(t *testing.T) {
	rates := mustRates(t, price(3), price(0), price(15))

	got := rates.Cost(Tokens{Input: 1_000_000, Cached: 9_000_000})

	if !closeEnough(got, 3) {
		t.Fatalf("cost is %v, want 3 with free cache reads", got)
	}
}

func TestLevelIsUnpricedWithoutBothInputAndOutput(t *testing.T) {
	cases := map[string]struct {
		input  *float64
		output *float64
	}{
		"no input":  {input: nil, output: price(15)},
		"no output": {input: price(3), output: nil},
		"neither":   {input: nil, output: nil},
	}

	for name, tc := range cases {
		if _, ok := NewRates(tc.input, price(0.3), tc.output); ok {
			t.Fatalf("%s: rates report priced", name)
		}
	}
}

func TestFreeModelIsPricedAtZero(t *testing.T) {
	rates, ok := NewRates(price(0), nil, price(0))
	if !ok {
		t.Fatal("a model priced at zero reports unpriced")
	}

	got := rates.Cost(Tokens{Input: 4_000_000, Cached: 1_000_000, Output: 2_000_000})

	if got != 0 {
		t.Fatalf("free model cost %v, want 0", got)
	}
}

func TestNegativeRateIsUnpriced(t *testing.T) {
	if _, ok := NewRates(price(-1), nil, price(15)); ok {
		t.Fatal("a negative input rate reports priced")
	}

	if _, ok := NewRates(price(3), nil, price(-15)); ok {
		t.Fatal("a negative output rate reports priced")
	}
}

func TestPricedLevelWithNoSpendCostsNothing(t *testing.T) {
	rates := mustRates(t, price(3), price(0.3), price(15))

	got := rates.Cost(Tokens{})

	if got != 0 {
		t.Fatalf("empty spend cost %v, want 0", got)
	}
}

func TestRealisticMagnitudesKeepTheirPrecision(t *testing.T) {
	rates := mustRates(t, price(3), price(0.3), price(15))

	got := rates.Cost(Tokens{Input: 1_234_567, Cached: 48_000_000, Output: 987_654})

	want := 3*1.234567 + 0.3*48 + 15*0.987654

	if !closeEnough(got, want) {
		t.Fatalf("cost is %v, want %v", got, want)
	}

	if !closeEnough(rates.Cost(Tokens{Cached: 48_000_000}), 14.4) {
		t.Fatalf("48M cached tokens cost %v, want 14.4", rates.Cost(Tokens{Cached: 48_000_000}))
	}
}

func TestFractionalRatesAddUpOverManySmallSpends(t *testing.T) {
	rates := mustRates(t, price(0.25), price(0.03), price(1.25))

	total := 0.0
	for range 1000 {
		total += rates.Cost(Tokens{Input: 1_000, Cached: 10_000, Output: 500})
	}

	want := 1000 * (0.25*0.001 + 0.03*0.01 + 1.25*0.0005)

	if !closeEnough(total, want) {
		t.Fatalf("total is %v, want %v", total, want)
	}
}
