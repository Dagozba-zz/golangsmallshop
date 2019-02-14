package parser

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type BulkRule struct {
	RuleName           string `yaml:"ruleName"`
	AffectedItem       string `yaml:"affectedItem"`
	TriggerAmount      int    `yaml:"triggerAmount"`
	DiscountPercentage int    `yaml:"discountPercentage"`
}

type NxMRule struct {
	RuleName     string `yaml:"ruleName"`
	AffectedItem string `yaml:"affectedItem"`
	BuyN         int    `yaml:"buyN"`
	PayM         int    `yaml:"payM"`
}

type Rules struct {
	NxmRules  []NxMRule  `yaml:"nxmRules"`
	BulkRules []BulkRule `yaml:"bulkRules"`
}

type generatedRules struct {
	Rules `yaml:"rules"`
}

//Visible for mocking
type IRuleParser interface {
	ParseRulesFile(p string) (Rules, error)
}

type RuleParser struct {}

//Parses the given yaml in the given path in order to add it to the microservice configuration
func (pa RuleParser) ParseRulesFile(p string) (Rules, error) {
	path, _ := filepath.Abs(p)
	d, err := ioutil.ReadFile(path)
	var rules Rules
	if err != nil {
		return rules, fmt.Errorf("the rules file couldn't be loaded: %v", err)
	}
	var a generatedRules
	yaml.Unmarshal(d, &a)

	return pa.validateAndReturnRules(a.Rules), nil
}

//Validates that the data in the yaml file makes sense
func (RuleParser) validateAndReturnRules(rules Rules) Rules {
	var validatedNxMRules []NxMRule
	for _, v := range rules.NxmRules {
		if err := v.validateNxMRuleInput(); err != nil {
			logrus.Warn(fmt.Errorf("the rule %s failed to be validated: , %v", v.RuleName, err))
		} else {
			validatedNxMRules = append(validatedNxMRules, v)
		}
	}

	var validatedBulkRules []BulkRule
	for _, v := range rules.BulkRules {
		if err := v.validateBulkRuleInput(); err != nil {
			logrus.Warn(fmt.Errorf("the rule %s failed to be validated: , %v", v.RuleName, err))
		} else {
			validatedBulkRules = append(validatedBulkRules, v)
		}
	}

	return Rules{BulkRules: validatedBulkRules, NxmRules: validatedNxMRules}
}

//TODO: Rules structs and validations should go on separate files to avoid clumping everything up in the same file
//Validates the given NxMRule, returns an error othweise
func (r NxMRule) validateNxMRuleInput() error {

	if r.AffectedItem == "" {
		return errors.New("the affected item can't be nil")
	}

	if r.PayM <= 0 {
		return errors.New("the pay amount can't be zero or below")
	}

	if r.BuyN <= 0 {
		return errors.New("the buy amount can't be zero or below")
	}

	if r.BuyN < r.PayM {
		return errors.New("the amount to pay can't be higher than the amount to buy")
	}

	return nil
}

//Validates the given BulkRule, returns an error otherwise
func (r BulkRule) validateBulkRuleInput() error {

	if r.AffectedItem == "" {
		return errors.New("the affected item can't be nil")
	}

	if r.TriggerAmount == 0 {
		return errors.New("the discount trigger amount can't be zero")
	}

	if r.DiscountPercentage < 0 || r.DiscountPercentage >= 100 {
		return errors.New("the discount percentage can't be lower than 0% or higher or equals than 100%")
	}

	return nil
}
