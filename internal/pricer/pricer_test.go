package pricer

import (
	"github.com/dagozba/golangsmallshop/internal/parser"
	"github.com/dagozba/golangsmallshop/internal/rules"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockedItemsParser struct {
	mock.Mock
}

func (m MockedItemsParser) ParseItemsDefinitions(p string) (parser.ConfiguredItems, error) {
	return parser.ConfiguredItems{
		"VOUCHER": parser.ItemDefinition{
			Name:  "Company Voucher",
			Price: 5.00,
		},
		"TSHIRT": parser.ItemDefinition{
			Name:  "Company T-Shirt",
			Price: 20.00,
		},
		"MUG": parser.ItemDefinition{
			Name:  "Company Coffee Mug",
			Price: 7.5,
		},
	}, nil
}

func TestCreateBasket(t *testing.T) {

	//ARRANGE
	pricer := &Pricer{}

	//ACT
	id := pricer.CreateBasket()

	//ASSERT
	if l := len(basketSession.baskets); l != 1 {
		t.Errorf("No basket has been inserted in the sessionBasket, expected: %d, got: %d", 1, l)
	}

	if _, exs := basketSession.baskets[id]; !exs {
		t.Errorf("The returned basket id doesn't match the inserted one")
	}
}

func TestCreateAccessRemoveBasketConcurrent(t *testing.T) {
	pricer := &Pricer{}
	for i := 0; i < 10; i++ {
		id := pricer.CreateBasket()
		go pricer.ScanItem("VOUCHER", id)
		go pricer.GetTotalAmount(id)
		go pricer.RemoveBasket(id)
	}
}

func TestScanItem(t *testing.T) {
	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	pricer := &Pricer{ItemsParser: itemsParserMock}
	pricer.LoadItems("DUMMYPATH")
	bId := pricer.CreateBasket()

	//ACT
	_, err := pricer.ScanItem("VOUCHER", bId)

	//ASSERT
	if err != nil {
		t.Errorf("Item scanning failed when it should've worked, err: %+v", err)
	}

	if v, _ := basketSession.baskets[bId].items["VOUCHER"]; v != 1 {
		t.Errorf("The item inserted in the basket doesn't match the provided one, expected: %s %d, got: %d", "VOUCHER", 1, v)
	}

}

func TestScanItemNonExistentItem(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	pricer := &Pricer{ItemsParser: itemsParserMock}
	pricer.LoadItems("DUMMYPATH")
	bId := pricer.CreateBasket()

	//ACT
	_, err := pricer.ScanItem("NONEXISTENT", bId)

	//ASSERT
	if err == nil {
		t.Errorf("There should've been an error produced by inserting a non configured item")
	}

}

func TestScanItemNonExistentBasket(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	pricer := &Pricer{ItemsParser: itemsParserMock}
	pricer.LoadItems("DUMMYPATH")

	//ACT
	_, err := pricer.ScanItem("VOUCHER", "FAKEBASKETID")

	//ASSERT
	if err == nil {
		t.Errorf("There should've been an error produced by accessing a non existent basket")
	}

}

func TestRemoveBasket(t *testing.T) {

	//ARRANGE
	pricer := &Pricer{}
	bId := pricer.CreateBasket()

	//ACT
	pricer.RemoveBasket(bId)

	//ASSERT
	if _, exs := basketSession.baskets[bId]; exs {
		t.Errorf("The current basket shouldn't be present in tbe baskets session map")
	}

}

func TestGetTotalAmountNoItems(t *testing.T) {

	//ASSERT
	pricer := &Pricer{}
	bId := pricer.CreateBasket()

	//ACT
	amount, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if amount != 0 {
		t.Errorf("The amount calculated should be 0 if there are no items scanned, got: %f", float64(amount))
	}

	if err != nil {
		t.Errorf("The calculation shouldn't have produced an error, got: %+v", err)
	}

}

func TestGetTotalAmountWithItems(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{rules.DefaultRuleStrategy{}}}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")


	bId := pricer.CreateBasket()
	pricer.ScanItem("VOUCHER", bId)

	//ACT
	amount, err := pricer.GetTotalAmount(bId)
	amount = amount / 100
	//ASSERT
	if amount != 5.0 {
		t.Errorf("The amount calculated should be VOUCHER's value, expected: %f, got: %f", 5.0, float64(amount))
	}

	if err != nil {
		t.Errorf("The calculation shouldn't have produced an error, got: %+v", err)
	}

}

func TestGetTotalAmountNonExistentBasket(t *testing.T) {

	//ARRANGE
	pricer := &Pricer{}

	//ACT
	_, err := pricer.GetTotalAmount("FAKEBASKETID")

	//ASSERT
	if err == nil {
		t.Errorf("An error should've been produced by getting the calculated amount of a non existent basket")
	}

}

func TestGetTotalAmountApplyingBulkRule(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{
		rules.BulkRuleStrategy{
			Rule: parser.BulkRule {RuleName: "Bulk Rule", AffectedItem: "TSHIRT", TriggerAmount: 3, DiscountPercentage: 5}}}}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("TSHIRT", bId)

	var expectedCalc int64 = 5700

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Applying the BulkRule to TSHIRT should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountApplyingBundleNxMRuleExactBundle(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{
		rules.NxMRuleStrategy{Rule: parser.NxMRule{RuleName: "NxM Rule", AffectedItem: "VOUCHER", BuyN: 2, PayM: 1}}}}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("VOUCHER", bId)

	var expectedCalc int64 = 500

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Applying the NxM to VOUCH should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountApplyingBundleNxMRuleBundleAndRemainder(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{
		RuleExecutors: []rules.RuleStrategyExecutor{
		rules.NxMRuleStrategy{
			Rule: parser.NxMRule{
				RuleName: "NxM Rule",
				AffectedItem: "VOUCHER",
				BuyN: 2,
				PayM: 1}}},
	}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")


	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("VOUCHER", bId)

	var expectedCalc int64 = 1000

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Applying the NxM to VOUCH should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountApplyingDefaultRule(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{rules.DefaultRuleStrategy{}}}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("MUG", bId)
	pricer.ScanItem("MUG", bId)
	pricer.ScanItem("MUG", bId)

	var expectedCalc int64 = 2250

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Applying the NxM to VOUCH should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountCompleteChallengeExample1(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{
		rules.DefaultRuleStrategy{}},
	}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("MUG", bId)

	var expectedCalc int64 = 3250

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Mixed items calculation should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountCompleteChallengeExample2(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{
		rules.DefaultRuleStrategy{},
		rules.NxMRuleStrategy{
			Rule: parser.NxMRule{
				RuleName: "NxM Rule",
				AffectedItem: "VOUCHER",
				BuyN: 2,
				PayM: 1}}},
	}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")
	rules.IncludedItems = map[string]bool{
		"VOUCHER": true,
	}

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("VOUCHER", bId)

	var expectedCalc int64 = 2500

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Mixed items calculation should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountCompleteChallengeExample3(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{
		rules.DefaultRuleStrategy{},
		rules.BulkRuleStrategy {
			Rule: parser.BulkRule{
				RuleName: "Bulk Rule",
				AffectedItem: "TSHIRT",
				TriggerAmount: 3,
				DiscountPercentage: 5}}},
	}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")
	rules.IncludedItems = map[string]bool{
		"TSHIRT": true,
	}

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("TSHIRT", bId)

	var expectedCalc int64 = 8100

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Mixed items calculation should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}

func TestGetTotalAmountCompleteChallengeExample4(t *testing.T) {

	//ARRANGE
	itemsParserMock := new(MockedItemsParser)
	rulesStrategyFactory := rules.RuleStrategyFactory{RuleExecutors: []rules.RuleStrategyExecutor{
		rules.DefaultRuleStrategy{},
		rules.NxMRuleStrategy{
			Rule: parser.NxMRule{
				RuleName: "NxM Rule",
				AffectedItem: "VOUCHER",
				BuyN: 2,
				PayM: 1}},
		rules.BulkRuleStrategy {
			Rule: parser.BulkRule{
				RuleName: "Bulk Rule",
				AffectedItem: "TSHIRT",
				TriggerAmount: 3,
				DiscountPercentage: 5}}},
	}
	pricer := &Pricer{ItemsParser: itemsParserMock, StrategyFactory: rulesStrategyFactory}
	pricer.LoadItems("DUMMYPATH")
	rules.IncludedItems = map[string]bool{
		"TSHIRT": true,
		"VOUCHER": true,
	}

	//ASSERT
	bId := pricer.CreateBasket()
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("VOUCHER", bId)
	pricer.ScanItem("MUG", bId)
	pricer.ScanItem("TSHIRT", bId)
	pricer.ScanItem("TSHIRT", bId)

	var expectedCalc int64 = 7450

	//ACT
	calc, err := pricer.GetTotalAmount(bId)

	//ASSERT
	if calc != expectedCalc {
		t.Errorf("Mixed items calculation should've produced %0.2f, got: %0.2f", float64(expectedCalc), float64(calc))
	}

	if err != nil {
		t.Errorf("No errors should've been produced by calculating these items, got: %+v", err)
	}

}
