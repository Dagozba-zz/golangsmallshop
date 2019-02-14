package client

import (
	"flag"
	pb "github.com/dagozba/golangsmallshop/internal/generated/api/v1"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

func InitializeConnection() *grpc.ClientConn {
	// Set up a connection to the server.
	address := flag.String("address", "localhost:50051", "The remote GRPC server address")
	conn, err := grpc.Dial(*address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
		//Panicking as the connection not being initialized is a non recoverable condition
		panic(err)
	}
	return conn
}

func CreateBasketCall() string {
	conn := InitializeConnection()
	defer conn.Close()
	c := pb.NewCheckoutClient(conn)
	r, _ := c.CreateBasket(context.Background(), &empty.Empty{})
	return r.BasketId
}

func ScanItemCall(basketId string, item string) (bool, error) {
	conn := InitializeConnection()
	defer conn.Close()
	c := pb.NewCheckoutClient(conn)
	r, err := c.ScanItem(context.Background(), &pb.ItemRequest{BasketId: basketId, ItemId: item})
	if err != nil {
		return false, err
	}
	return r.Result, nil
}

func GetTotalAmountCall(basketId string) (int64, error) {
	conn := InitializeConnection()
	defer conn.Close()
	c := pb.NewCheckoutClient(conn)
	r, err := c.GetTotalAmount(context.Background(), &pb.TotalAmountRequest{BasketId: basketId})
	if err != nil {
		return 0, err
	}
	return r.TotalAmount, nil
}

func RemoveBasketCall(basketId string) bool {
	conn := InitializeConnection()
	defer conn.Close()
	c := pb.NewCheckoutClient(conn)
	r, _ := c.RemoveBasket(context.Background(), &pb.RemoveBasketRequest{BasketId: basketId})
	return r.Result
}
