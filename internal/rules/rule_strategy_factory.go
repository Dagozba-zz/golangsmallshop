package rules

import (
	"github.com/dagozba/golangsmallshop/internal/parser"
	log "github.com/sirupsen/logrus"
)

type RuleStrategyFactory struct {
	RuleExecutors []RuleStrategyExecutor
	RuleParser    parser.IRuleParser
}

//Interface that serves as an abstraction layer for the Pricer, executing this method for any struct that implements this interface
type RuleStrategyExecutor interface {
	ExecuteRule(conf parser.ConfiguredItems, scannedItems map[string]int) int64
}

type BulkRuleStrategy struct {
	Rule parser.BulkRule
}

type NxMRuleStrategy struct {
	Rule parser.NxMRule
}

type DefaultRuleStrategy struct{}


var IncludedItems map[string]bool

//It begins parsing the rules defined in the /configs/rules.yaml file
//then it creates a matching rule strategy and adds it to the executors slice
//all the items affected by any promotion are added to a map so the default rule can apply to the items not included in it
func (f *RuleStrategyFactory) LoadRules(filePath string) error {
	log.Info("Parsing initial Rules for Rule Strategy Factory")
	rules, err := f.RuleParser.ParseRulesFile(filePath)
	if err != nil {
		return err
	}
	IncludedItems = make(map[string]bool)
	for _, v := range rules.BulkRules {
		f.RuleExecutors = append(f.RuleExecutors, BulkRuleStrategy{Rule: v})
		log.Infof("Applying BulkRule for item: %s - Default rule will not be applied to this item", v.AffectedItem)
		IncludedItems[v.AffectedItem] = true
	}

	for _, v := range rules.NxmRules {
		f.RuleExecutors = append(f.RuleExecutors, NxMRuleStrategy{Rule: v})
		log.Infof("Applying Bundle (NxMRule) for item: %s - Default rule will not be applied to this item", v.AffectedItem)
		IncludedItems[v.AffectedItem] = true
	}

	f.RuleExecutors = append(f.RuleExecutors, DefaultRuleStrategy{})

	return err
}

//Executes the BulkRule calculation
//It gets the number of items affected by this rule in the scanned items map
//if the number of items affected is equal or higher than the configured trigger amount (ie: if you buy 10 and trigger amount is 5)
//else, it applies the default formula
//As the discount % is given as an integer, it has to be converted into a decimal in order to apply it to the configured price
//final calculation formula is: number of items * configured price * % in decimal (if 20, then 0.20)
func (s BulkRuleStrategy) ExecuteRule(conf parser.ConfiguredItems, scannedItems map[string]int) int64 {
	if a, exs := scannedItems[s.Rule.AffectedItem]; exs && int(a) >= s.Rule.TriggerAmount {
		result := float32(a) * conf[s.Rule.AffectedItem].Price * float32(1.0-float32(s.Rule.DiscountPercentage)/100)
		return int64(result * 100)
	} else if exs {
		return int64((conf[s.Rule.AffectedItem].Price * float32(a)) * 100)
	} else {
		return 0
	}
}

//Executes a Bundle Rule calculation
//It gets the number of items affected by this rule in the scanned items map
//then it calculates the number of bundles (ie: for a Buy 2 pay 1 Rule and 3 items, the number of bundles is 2)
//then it calculates the reminder, which is the item left over from the bundle that is not affected by the promotion
//finally, it applies the formula: (number of bundles * items_to_pay) * item price
func (s NxMRuleStrategy) ExecuteRule(conf parser.ConfiguredItems, scannedItems map[string]int) int64 {
	a, _ := scannedItems[s.Rule.AffectedItem]
	bundles := a / s.Rule.BuyN
	remainder := a % s.Rule.BuyN
	return int64((float32(bundles*s.Rule.PayM+remainder) * conf[s.Rule.AffectedItem].Price) * 100)
}

//Executes the default rule for all items not affected by pricing rules
//default rule is just the items' configured price
func (DefaultRuleStrategy) ExecuteRule(conf parser.ConfiguredItems, scannedItems map[string]int) int64 {
	var totalAmount float32
	for k, v := range scannedItems {
		if _, ok := IncludedItems[k]; !ok {
			totalAmount += conf[k].Price * float32(v)
		}
	}
	return int64(totalAmount * 100)
}
