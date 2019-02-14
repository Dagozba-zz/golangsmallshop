package main

import (
	"fmt"
	grpcClient "github.com/dagozba/golangsmallshop/internal/client"
	"gopkg.in/urfave/cli.v1"
	"os"
	"time"
)

//This is the CLI that will interact with the GRPC server.
func main() {

	app := cli.NewApp()
	app.Name = "cli"
	app.Usage = `This CLI helps you interact with the Checkout GRPC server by providing certain commands and information about the current state of the service for a given basket.`
	app.Author = "Daniel Gozalo"
	app.Compiled = time.Now()
	app.Email = "Dagozba@gmail.com"
	app.Version = "1.0.0"

	app.Commands = []cli.Command{
		{
			Name:    "basket",
			Aliases: []string{"b"},
			Usage:   "Interacts with baskets",
			Subcommands: []cli.Command{
				{
					Name: "create",
					Action: func(c *cli.Context) {
						id := grpcClient.CreateBasketCall()
						fmt.Println("Created Basket with id: ", id)
					},
				},
				{
					Name:  "delete",
					Usage: "BASKETID",
					Action: func(c *cli.Context) {
						basketId := c.Args().First()
						fmt.Println("Basket to delete: ", basketId)
						grpcClient.RemoveBasketCall(basketId)
					},
				},
			},
		},
		{
			Name:    "scan",
			Aliases: []string{"s"},
			Usage:   "Scans an item to add it to the given basket",
			Action: func(c *cli.Context) {
				basketId := c.Args().First()
				item := c.Args().Get(1)
				fmt.Println("Basket id: ", basketId)
				fmt.Println("Item to Assign: ", item)
				result, err := grpcClient.ScanItemCall(basketId, item)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				if result {
					fmt.Printf("Item %s correctly scanned\n", item)
				}

			},
		},
		{
			Name:    "get-price",
			Aliases: []string{"g"},
			Usage:   "Gets the accumulated price of a given basket",
			Action: func(c *cli.Context) {
				basketId := c.Args().First()
				fmt.Println("Basket id: ", basketId)
				p, err := grpcClient.GetTotalAmountCall(basketId)
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				fmt.Printf("Obtained price is %.2f\n: ", float64(p / 100))
			},
		},
	}

	app.Run(os.Args)

}
