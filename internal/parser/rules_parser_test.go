package parser

import "testing"

func TestParseRulesFile(t *testing.T) {

	//ARRANGE
	c := Rules{
		BulkRules: []BulkRule{
			{
				RuleName:           "Bulk Rule",
				AffectedItem:       "TSHIRT",
				DiscountPercentage: 5,
				TriggerAmount:      3,
			},
		},
		NxmRules: []NxMRule{
			{
				RuleName:     "Buy N pay M Rule",
				AffectedItem: "VOUCHER",
				BuyN:         2,
				PayM:         1,
			},
		},
	}
	rulesParser := &RuleParser{}

	//ACT
	pc, _ := rulesParser.ParseRulesFile("../../configs/rules.yaml")

	//ASSERT
	for i := range c.BulkRules {
		if c.BulkRules[i] != pc.BulkRules[i] {
			t.Errorf("BulkRules slices are not equal")
		}
	}

	for i := range c.NxmRules {
		if c.NxmRules[i] != pc.NxmRules[i] {
			t.Errorf("NxMRules slices are not equal")
		}
	}

}
