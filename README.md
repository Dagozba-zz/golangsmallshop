Small Shop developed in Golang
============================================

This was my first ever development using Go, I've tried to do it as idiomatic as possible considering how
opinionated Go is. I've mainly used the official Go documentation and Effective Go to support this development.

**Features:**

* GRPC based communication between CLI and Server using Protocol Buffers.
* Docker ready server.
* Makefile available to automate building, testing and vetting.
* Pricing Rules and Configured Items loaded from .yaml files (Not hardcoded).

Solution
--------

Two pieces have been developed, a Server and a CLI to interact with the server.

The server is a GRPC server that listens to requests made to the CheckoutService, exposing the following actions:

* CreateBasket
* DeleteBasket
* Scan item
* CalculateTotal

It makes use of pricing rules in order to apply different discounts and promotions on configured items.

The Factory design pattern is used to generate different rule strategies, loaded from the configs/rules.yaml file. All these rules implement the
RuleStrategyExecutor interface in order to let the Pricer execute each rule's logic, passing them the configured items and the scanned items.

The baskets are stored in an in a map in memory. I didn't want to use databases or any other storage as I wanted to test concurrency, so I decided that using
mutexes was the right way to do this as it would lock the whole basket and not the whole methods, so it's concurrent and thread safe at the same time.

A Basket needs to be created before scanning items to a basket

The client is a simple Go CLI that interacts with the server using the GRPC client generated from the .proto file.

There's a way to configure the number of items of the promotion in the /configs/rules.yaml:

    rules:
      nxmRules:
      - affectedItem: VOUCHER
        buyN: 2
        payM: 1
        
In this case, it comes preconfigured with a 2x1 (Buy 2, pay 1, as expressed in the examples), but it could be extended to a 3x2:

    rules:
      nxmRules:
      - affectedItem: VOUCHER
        buyN: 3
        payM: 2

Please note that a restart to the Server would be needed for this change to take effect. This wouldn't be a problem if this configuration was loaded at runtime, it could be
accomplished by storing it in either a database, a configuration store, or an objects store like S3.

Getting Started
---------------

### Installation

Makefile is provided in order to regenerate the binaries, simply by calling:

    $ make

This will trigger the LINK, TEST, VET and BUILD commands of the Makefile.

The protocol buffers are already compiled and ready to be used by the application.

Usage
-----

### Server
The server binary is located in cmd/server. When started, the gRPC server will start listening on localhost:50051 by default but it can be changed by using the -host flag
It will also need the flags "-items-path" and "-rules-path" to find the configuration yaml files.

    $ cd cmd/server
    $ ./server-<CHOSEN_ARCHITECTURE>

Alternatively, a Dockefile has been provided in order to generate a Docker image for the server.

**Usage:**

To generate the image:

    $ docker build -t golang_small_shop_server:1.0.0 .

To execute it:

    $ docker run -d -p 50051:50051 golang_small_shop_server:1.0.0

The internal container port 50051 is being published to the host's 50051 port, so it must be free.

### CLI
A CLI is provided to interact with the server, usage can be checked by executing:
    
    $ cd cmd/cli
    $ ./cli-<CHOSEN_ARCHITECTURE> help

**Commands:**

* basket create -> Creates a basket in the server and returns its identifier for later use
* basket delete BASKET_ID -> Deletes the basket in the server. Must be provided with a basket id.
* scan [BASKET_ID, ITEM_ID] -> Scans an item, inserting it in the provided basket. Must be provided with a basket id and an item id.
* get-price [BASKET_ID] -> Calculates the total price of all scanned items within a basket, using the configured pricing rules. Must be provided with a basket id.

Commands example:

    $ ./cli-linux-amd64 basket create

This generates a basket id, say: **12456789**.
    
    $ ./cli-linux-amd64 basket delete 12456789

    $ ./cli-linux-amd64 scan 12456789 VOUCHER
    
    $ ./cli-linux-amd64 get-price 12456789

This should return the calculated price of the basket.


### Thread safety considerations for the in memory map

I've created a new Struct called BasketSession with the baskets map and a pointer to a RWMutex.

I've used a pointer to a RWMutex in the Basket struct as well.

Methods are receiver methods, meaning they only apply to the structs that are being used in that thread,
it also helps keep code clean as locking / unlocking in the main methods would get the code dirty.

I did this in order to only lock the underlying data structure behind the maps inside each instance of these structs, in order to
avoid locking the incoming threads of different baskets.

Like this, instead of locking all incoming threads with a global RWMutex, if, say we have two baskets performing costly price calculations
and we try to add a new item to that basket, the thread will wait until the calculations are done, however, if you try to add an item to another
basket that isn't locked, you will be able to, and even calculate the amount of that basket.

I've also added a test to check concurrency to pricer_test.go.


### Lessons learned when refactoring 1 year later

- Panics are evil if not used when they have to be used, I had to refactor a lot to control the errors
- init() method is good for small initializations, not to be used for error prone executions as you can't control the errors
- As I have not used a dependency injection library, it's always a good idea to use structs and "constructors" to define the behaviours and not just functions within files, this way its dependencies can be passed and be way more testable and mockable
- Go modules weren't around a year ago, it's totally game changing and I love it.
- I added mocking and improved the tests A LOT
- Flags, use flags
- Versioned the API
- Control errors, it's the idiomatic way!

### Things to improve

- My model should be separate from the code using them. For example, the Baskets, Items etc should have their own files and define their behaviour there, this leads to cleaner code
- Improve tests even more
- Use a CI server to automate everything when uploaded to Github
- Sonar !!
- Control errors even more
- Add debug traces