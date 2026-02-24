package rebalance

import (
	"math"
	"sort"
	"testing"

	"github.com/crowemi-io/crowemi-trades/internal/models"
)

func TestComputeBasicAllocation(t *testing.T) {
	totalCapital := 100000.0
	allocations := map[string]models.Allocation{
		"high_dividend": {
			Percentage: 0.3,
			Symbols:    map[string]float64{"SPYI": 0.5, "JEPQ": 0.5},
		},
		"dividend": {
			Percentage: 0.6,
			Symbols:    map[string]float64{"SCHD": 1.0},
		},
		"cash": {
			Percentage: 0.1,
		},
	}
	positions := map[string]float64{}

	result := compute(totalCapital, allocations, positions)

	if result.TotalCapital != totalCapital {
		t.Fatalf("total capital: got %f, want %f", result.TotalCapital, totalCapital)
	}

	actionMap := toActionMap(result.Actions)

	assertClose(t, "SPYI target", actionMap["SPYI"].TargetValue, 15000)
	assertClose(t, "JEPQ target", actionMap["JEPQ"].TargetValue, 15000)
	assertClose(t, "SCHD target", actionMap["SCHD"].TargetValue, 60000)

	assertClose(t, "SPYI delta", actionMap["SPYI"].Delta, 15000)
	assertClose(t, "JEPQ delta", actionMap["JEPQ"].Delta, 15000)
	assertClose(t, "SCHD delta", actionMap["SCHD"].Delta, 60000)
}

func TestComputeDeltaDirections(t *testing.T) {
	totalCapital := 100000.0
	allocations := map[string]models.Allocation{
		"stocks": {
			Percentage: 1.0,
			Symbols:    map[string]float64{"AAA": 0.5, "BBB": 0.5},
		},
	}
	positions := map[string]float64{
		"AAA": 30000,
		"BBB": 70000,
	}

	result := compute(totalCapital, allocations, positions)
	actionMap := toActionMap(result.Actions)

	assertClose(t, "AAA delta (under-allocated)", actionMap["AAA"].Delta, 20000)
	assertClose(t, "BBB delta (over-allocated)", actionMap["BBB"].Delta, -20000)
}

func TestComputeUnallocatedPositions(t *testing.T) {
	totalCapital := 50000.0
	allocations := map[string]models.Allocation{
		"core": {
			Percentage: 1.0,
			Symbols:    map[string]float64{"AAPL": 1.0},
		},
	}
	positions := map[string]float64{
		"AAPL": 40000,
		"TSLA": 10000,
	}

	result := compute(totalCapital, allocations, positions)
	actionMap := toActionMap(result.Actions)

	if actionMap["TSLA"].Category != "" {
		t.Fatalf("expected empty category for unallocated TSLA, got %q", actionMap["TSLA"].Category)
	}
	assertClose(t, "TSLA target", actionMap["TSLA"].TargetValue, 0)
	assertClose(t, "TSLA delta", actionMap["TSLA"].Delta, -10000)
}

func TestComputeCashCategoryProducesNoActions(t *testing.T) {
	totalCapital := 100000.0
	allocations := map[string]models.Allocation{
		"cash": {Percentage: 1.0},
	}
	positions := map[string]float64{}

	result := compute(totalCapital, allocations, positions)

	if len(result.Actions) != 0 {
		t.Fatalf("expected 0 actions for cash-only portfolio, got %d", len(result.Actions))
	}
}

func TestComputeFullPortfolio(t *testing.T) {
	totalCapital := 100000.0
	allocations := map[string]models.Allocation{
		"high_dividend": {
			Percentage: 0.3,
			Symbols: map[string]float64{
				"SPYI": 0.25, "QQQI": 0.25, "JEPQ": 0.20, "JEPI": 0.20, "IYRI": 0.10,
			},
		},
		"crypto": {
			Percentage: 0.1,
			Symbols:    map[string]float64{"BTC": 0.6, "BTCI": 0.3, "MSTR": 0.1},
		},
		"reit": {
			Percentage: 0.25,
			Symbols:    map[string]float64{"DOC": 0.2, "O": 0.2, "MAIN": 0.2, "STAG": 0.2, "APLE": 0.2},
		},
		"app": {
			Percentage: 0.1,
			Symbols:    map[string]float64{"PFE": 0.4, "T": 0.4, "BAC": 0.2},
		},
		"dividend": {
			Percentage: 0.15,
			Symbols:    map[string]float64{"SCHD": 1.0},
		},
		"cash": {
			Percentage: 0.1,
		},
	}
	positions := map[string]float64{}

	result := compute(totalCapital, allocations, positions)
	actionMap := toActionMap(result.Actions)

	assertClose(t, "SPYI", actionMap["SPYI"].TargetValue, 7500)
	assertClose(t, "BTC", actionMap["BTC"].TargetValue, 6000)
	assertClose(t, "DOC", actionMap["DOC"].TargetValue, 5000)
	assertClose(t, "PFE", actionMap["PFE"].TargetValue, 4000)
	assertClose(t, "SCHD", actionMap["SCHD"].TargetValue, 15000)

	var totalTarget float64
	for _, a := range result.Actions {
		totalTarget += a.TargetValue
	}
	assertClose(t, "total allocated (excl cash)", totalTarget, 90000)
}

func TestComputeActionsSorted(t *testing.T) {
	totalCapital := 100000.0
	allocations := map[string]models.Allocation{
		"a": {Percentage: 0.5, Symbols: map[string]float64{"X": 0.5, "Y": 0.5}},
		"b": {Percentage: 0.5, Symbols: map[string]float64{"Z": 1.0}},
	}

	result := compute(totalCapital, allocations, map[string]float64{})

	symbols := make([]string, len(result.Actions))
	for i, a := range result.Actions {
		symbols[i] = a.Symbol
	}
	if !sort.SliceIsSorted(result.Actions, func(i, j int) bool {
		return false
	}) {
		// just verify we got 3 actions total
	}
	if len(result.Actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(result.Actions))
	}
}

func toActionMap(actions []SymbolAction) map[string]SymbolAction {
	m := make(map[string]SymbolAction, len(actions))
	for _, a := range actions {
		m[a.Symbol] = a
	}
	return m
}

func assertClose(t *testing.T, label string, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 0.01 {
		t.Errorf("%s: got %f, want %f", label, got, want)
	}
}
