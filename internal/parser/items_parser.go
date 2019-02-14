package parser

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type ItemDefinition struct {
	Name  string  `yaml:"name"`
	Price float32 `yaml:"price"`
}

type generatedItemDefinitions struct {
	Items ConfiguredItems `yaml:"items"`
}

type ConfiguredItems map[string]ItemDefinition

type IParser interface {
  ParseItemsDefinitions(p string) (ConfiguredItems, error)
}

type ItemsParser struct {}

//Parses the configs/item_definitions.yaml to configure the system with available products.
//If the yaml format is not valid, it will not populate the Items map
func (pa ItemsParser) ParseItemsDefinitions(p string) (ConfiguredItems, error) {
	path, _ := filepath.Abs(p)
	d, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g generatedItemDefinitions
	yaml.Unmarshal(d, &g)
	return pa.validateInput(g.Items), nil
}

//Validates the input and discards any non valid items
func (pa ItemsParser) validateInput(items ConfiguredItems) ConfiguredItems {
	validatedItems := ConfiguredItems{}
	for k, v := range items {
		if err := v.validateItemInput(); err != nil {
			logrus.Warn(fmt.Errorf("the item %s failed to be validated: , %v", k, err))
		} else {
			logrus.Infof("Loaded item %s with price %.2f from configuration file", k, v.Price)
			validatedItems[k] = v
		}
	}
	return validatedItems
}

//Checks whether the given ItemDefinition is valid, returns an error otherwise
func (i ItemDefinition) validateItemInput() error {

	if i.Name == "" {
		return errors.New("the name of the configured item can't be nil or empty")
	}

	if i.Price <= 0 {
		return errors.New("the price of a product can't be 0 or lower")
	}

	return nil
}
