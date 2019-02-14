package rules

import (
	"github.com/dagozba/golangsmallshop/internal/parser"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockedRulesParser struct {
	mock.Mock
}

func (m MockedRulesParser) ParseRulesFile(p string) (parser.Rules, error) {
	return parser.Rules{
		NxmRules:  []parser.NxMRule{{
			RuleName:     "NxM Rule",
			AffectedItem: "VOUCHER",
			BuyN:         2,
			PayM:         1,
		}},
		BulkRules: []parser.BulkRule{{
			RuleName: "BulkRule",
			AffectedItem: "TSHIRT",
			TriggerAmount: 3,
			DiscountPercentage: 5,
		}},
	}, nil
}


func getConfiguredItems() parser.ConfiguredItems {
	return parser.ConfiguredItems{
		"VOUCHER": {
			Name:  "Company Voucher",
			Price: 5.00,
		},
		"TSHIRT": {
			Name:  "Company T-Shirt",
			Price: 20.00,
		},
		"MUG": {
			Name:  "Company Coffee Mug",
			Price: 7.5,
		},
	}
}

func TestLoadConfiguredRules(t *testing.T) {

	//ARRANGE
	expectedRules := 3
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}

	//ACT
	rulesFactory.LoadRules("PATH")

	//ASSERT
	if actualLength := len(rulesFactory.RuleExecutors); actualLength != expectedRules {
		t.Errorf("There should be %d configured rules (1 Bulk, 1 NxM and Default), got: %d", expectedRules, actualLength)
	}

}

func TestNxMRuleStrategy_ExecuteRuleExactBundle(t *testing.T) {

	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("PATH")
	var nxmRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case NxMRuleStrategy:
			nxmRuleStrategy = v
		}
	}

	var expectedCalc int64 = 500

	c := getConfiguredItems()

	//ACT
	result := nxmRuleStrategy.ExecuteRule(c, map[string]int{
		"VOUCHER": 2,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("NxMRule should have been applied once, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

func TestNxMRuleStrategy_ExecuteRuleBundleAndRemainder(t *testing.T) {

	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var nxmRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case NxMRuleStrategy:
			nxmRuleStrategy = v
		}
	}

	var expectedCalc int64 = 1000

	c := getConfiguredItems()

	//ACT
	result := nxmRuleStrategy.ExecuteRule(c, map[string]int{
		"VOUCHER": 3,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("NxMRule should have been applied once, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}


func TestNxMRuleStrategy_ExecuteRuleNoAffectedItems(t *testing.T) {

	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var nxmRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case NxMRuleStrategy:
			nxmRuleStrategy = v
		}
	}

	var expectedCalc int64 = 0

	c := getConfiguredItems()

	//ACT
	result := nxmRuleStrategy.ExecuteRule(c, map[string]int{
		"MUG":    2,
		"TSHIRT": 1,
	})

	//ASSERT
	//ASSERT
	if result != expectedCalc {
		t.Errorf("NxMRule should have not been applied, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

func TestBulkRuleStrategy_ExecuteRule(t *testing.T) {
	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var bulkRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case BulkRuleStrategy:
			bulkRuleStrategy = v
		}
	}

	var expectedCalc int64 = 5700

	c := getConfiguredItems()

	//ACT
	result := bulkRuleStrategy.ExecuteRule(c, map[string]int{
		"TSHIRT": 3,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("BUlkRule should have been applied once, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

func TestBulkRuleStrategy_ExecuteRuleCorrectItemDiscountNotTriggered(t *testing.T) {
	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var bulkRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case BulkRuleStrategy:
			bulkRuleStrategy = v
		}
	}

	var expectedCalc int64 = 4000

	c := getConfiguredItems()

	//ACT
	result := bulkRuleStrategy.ExecuteRule(c, map[string]int{
		"TSHIRT": 2,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("BUlkRule should have not been applied, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

func TestBulkRuleStrategy_ExecuteRuleNotAffectedItems(t *testing.T) {
	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var bulkRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case BulkRuleStrategy:
			bulkRuleStrategy = v
		}
	}

	var expectedCalc int64 = 0

	c := getConfiguredItems()

	//ACT
	result := bulkRuleStrategy.ExecuteRule(c, map[string]int{
		"VOUCHER": 1,
		"MUG":     1,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("BUlkRule should have not been applied, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

func TestDefaultRuleStrategy_ExecuteRule(t *testing.T) {
	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var defaultRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case DefaultRuleStrategy:
			defaultRuleStrategy = v
		}
	}

	var expectedCalc int64 = 2250

	c := getConfiguredItems()

	//ACT
	result := defaultRuleStrategy.ExecuteRule(c, map[string]int{
		"MUG": 3,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("Default rule should have been applied once, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}


func TestDefaultRuleStrategy_ExecuteRuleOnlyNonAffectedItems(t *testing.T) {
	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var defaultRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case DefaultRuleStrategy:
			defaultRuleStrategy = v
		}
	}

	var expectedCalc int64 = 0

	c := getConfiguredItems()

	//ACT
	result := defaultRuleStrategy.ExecuteRule(c, map[string]int{
		"VOUCHER": 1,
		"TSHIRT":  1,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("BUlkRule should have not been applied, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

func TestDefaultRuleStrategy_ExecuteRuleMixedItemsNotAffectedIgnored(t *testing.T) {
	//ARRANGE
	rulesFactory := RuleStrategyFactory{RuleParser: MockedRulesParser{}}
	rulesFactory.LoadRules("DUMMY_PATH")
	var defaultRuleStrategy RuleStrategyExecutor
	for _, v := range rulesFactory.RuleExecutors {
		switch v.(type) {
		case DefaultRuleStrategy:
			defaultRuleStrategy = v
		}
	}

	var expectedCalc int64 = 750

	c := getConfiguredItems()

	//ACT
	result := defaultRuleStrategy.ExecuteRule(c, map[string]int{
		"VOUCHER": 1,
		"MUG":     1,
		"TSHIRT":  1,
	})

	//ASSERT
	if result != expectedCalc {
		t.Errorf("BUlkRule should been applied once, ignoring non affected items, expected: %.2f, got %.2f", float64(expectedCalc), float64(result))
	}

}

