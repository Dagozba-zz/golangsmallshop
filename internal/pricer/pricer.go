package pricer

import (
	"errors"
	"github.com/dagozba/golangsmallshop/internal/parser"
	"github.com/dagozba/golangsmallshop/internal/rules"
	"github.com/segmentio/ksuid"
	log "github.com/sirupsen/logrus"
	"sync"
)

//Trying to follow the Inversion of Control principle through Dependency Injection using the "Constructor"
//This allows for better unit testing as dependencies can be mocked or dummies can be created
type Pricer struct {
	StrategyFactory rules.RuleStrategyFactory
	ItemsParser     parser.IParser
	ConfiguredItems parser.ConfiguredItems
}

type Item struct {
	id    string
	name  string
	price float32
}

type Basket struct {
	items     map[string]int
	itemsLock *sync.RWMutex
}

type BasketSession struct {
	baskets     map[string]*Basket
	basketsLock *sync.RWMutex
}

//The map where the created baskets is stored. In a normal microservices environment, this would not be ideal at all
//as we would want our microservices to be stateless and all the session state should be stored in an session store
//In AWS, it would probably be DynamoDB or ElasticCache.
//The whole idea of this application was to test as many things as possible from just Golang, so I went for concurrent
//access to an in memory map
var basketSession = BasketSession{baskets: make(map[string]*Basket), basketsLock: new(sync.RWMutex)}

//It begins parsing the item definitions defined in /configs/item_definitions.yaml
//This would be stored in a database or a cloud configuration service so it could be modified at runtime, but I didn't wan
//to include a Database access for this exercise as I wanted to try concurrent access to an in memory map
func (p *Pricer) LoadItems(itemsFilePath string) error {
	log.Info("Parsing initial Item Definitions for Pricer")
	configuredItems, err := p.ItemsParser.ParseItemsDefinitions(itemsFilePath)
	p.ConfiguredItems = configuredItems
	return err
}

//Executes all the rules loaded from the yaml file on the basket items
func (b Basket) executeRules(executors []rules.RuleStrategyExecutor, configuredItems parser.ConfiguredItems) int64 {
	var total int64
	b.itemsLock.Lock()
	defer b.itemsLock.Unlock()
	for _, executor := range executors {
		total += executor.ExecuteRule(configuredItems, b.items)
	}
	return total
}

//Adds an item to the given basket
func (b *Basket) addItemToBasket(i string) {
	b.itemsLock.Lock()
	defer b.itemsLock.Unlock()
	if _, exs := b.items[i]; exs {
		b.items[i]++
	} else {
		b.items[i] = 1
	}
}

func (bs BasketSession) getBasket(key string) *Basket {
	bs.basketsLock.RLock()
	defer bs.basketsLock.RUnlock()
	return bs.baskets[key]
}

func (bs BasketSession) createBasket() string {
	id := ksuid.New().String()
	log.Infof("Generating basket with id '%s'", id)
	bs.basketsLock.Lock()
	defer bs.basketsLock.Unlock()
	bs.baskets[id] = &Basket{items: make(map[string]int), itemsLock: new(sync.RWMutex)}
	return id
}

func (bs BasketSession) deleteBasket(basketId string) {
	bs.basketsLock.Lock()
	defer bs.basketsLock.Unlock()
	delete(basketSession.baskets, basketId)
}

//It creates a new UID as the basket identifier and adds it to the basketsSession map with a pointer to a Basket struct
//where scanned items will be stored
func (p Pricer) CreateBasket() string {
	return basketSession.createBasket()
}

//Stores an item in the given basket. returns an error if the basket doesn't exist or the item has not been defined by configuration
func (p *Pricer) ScanItem(i string, basketId string) (bool, error) {
	log.Infof("Scanning item %s into basket %s", i, basketId)
	basket := basketSession.getBasket(basketId)
	if basket == nil {
		log.Errorf("The basket '%s' doesn't exist", basketId)
		return false, errors.New("the specified basket doesn't exist")
	}
	if _, prs := p.ConfiguredItems[i]; prs == false {
		log.Errorf("the item '%s' has not been configured in the server", i)
		return false, errors.New("the specified item is not configured in the server")
	}
	basket.addItemToBasket(i)
	log.Infof("Item %s added to the basket %s", i, basketId)
	return true, nil
}

//Calculates the total price for the items in the given basket by executing all the Pricing Rules in the RuleExecutors slice
//As all rules implement the RuleStrategyExecutor interface, by calling ExecuteRule any rule can be executed and the Pricer
//delegates the rules creation and execution logic to the rules strategy factory.
//if the basket doesn't exist, an error is returned
func (p Pricer) GetTotalAmount(basketId string) (int64, error) {
	log.Infof("Getting total amount of items with applied discounts in basket %s", basketId)
	basket := basketSession.getBasket(basketId)
	if basket == nil {
		log.Errorf("The basket '%s' doesn't exist", basketId)
		return 0, errors.New("the specified basket doesn't exist")
	} else {
		return basket.executeRules(p.StrategyFactory.RuleExecutors, p.ConfiguredItems), nil
	}
}

//Removes the basket from the basketSession map
func (p Pricer) RemoveBasket(basketId string) bool {
	log.Infof("Removing basket '%s'", basketId)
	basketSession.deleteBasket(basketId)
	log.Infof("Basket '%s' has been removed", basketId)
	return true
}
