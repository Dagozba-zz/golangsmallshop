package parser

import "testing"

func TestParseItemsDefinitions(t *testing.T) {

	//ARRANGE
	c := ConfiguredItems{
		"VOUCHER": ItemDefinition {
			Name:  "Company Voucher",
			Price: 5.00,
		},
		"TSHIRT": ItemDefinition {
			Name:  "Company T-Shirt",
			Price: 20.00,
		},
		"MUG": ItemDefinition {
			Name:  "Company Coffee Mug",
			Price: 7.5,
		},
	}
	itemsParser := &ItemsParser{}

	//ACT
	pc, _ := itemsParser.ParseItemsDefinitions("../../configs/item_definitions.yaml")

	//ASSERT
	for k, v := range c {
		if pcv, exs := pc[k]; !exs {
			t.Errorf("Item key %s missing from the parsed map", k)
		} else {
			if v != pcv {
				t.Errorf("The contents of the item with key %s don't match, expected: %+v, got: %+v", k, v, pcv)
			}
		}
	}

}
