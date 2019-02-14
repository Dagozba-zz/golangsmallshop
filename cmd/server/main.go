package main

import (
	"flag"
	pb "github.com/dagozba/golangsmallshop/internal/generated/api/v1"
	"github.com/dagozba/golangsmallshop/internal/parser"
	"github.com/dagozba/golangsmallshop/internal/pricer"
	"github.com/dagozba/golangsmallshop/internal/rules"
	"github.com/golang/protobuf/ptypes/empty"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
	"os"
)

type server struct {
	pricer pricer.Pricer
}

func (s *server) CreateBasket(context.Context, *empty.Empty) (*pb.BasketReply, error) {
	id := s.pricer.CreateBasket()
	return &pb.BasketReply{BasketId: id}, nil
}

func (s *server) ScanItem(context context.Context, request *pb.ItemRequest) (*pb.ItemReply, error) {
	result, err := s.pricer.ScanItem(request.ItemId, request.BasketId)
	return &pb.ItemReply{Result: result}, err
}

func (s *server) GetTotalAmount(context context.Context, request *pb.TotalAmountRequest) (*pb.TotalAmountReply, error) {
	totalAmount, err := s.pricer.GetTotalAmount(request.BasketId)
	return &pb.TotalAmountReply{TotalAmount: totalAmount}, err
}

func (s *server) RemoveBasket(context context.Context, request *pb.RemoveBasketRequest) (*pb.RemoveBasketReply, error) {
	result := s.pricer.RemoveBasket(request.BasketId)
	return &pb.RemoveBasketReply{Result: result}, nil
}

//It starts the GRPC server that will listen to requests to the CheckoutService
func main() {

	var (
		port                    = flag.String("host", ":50051", "GRPC service address")
		rulesFilePath           = flag.String("rules-path", "", "The path to the Rules yaml config file")
		itemDefinitionsFilePath = flag.String("items-path", "", "The path to the item definitions yaml config file")
	)

	flag.Parse()

	log.Info("Starting GRPC server listening on port: ", *port)
	lis, err := net.Listen("tcp", *port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		os.Exit(1)
	}

	ruleFactory := &rules.RuleStrategyFactory{RuleParser: parser.RuleParser{}}
	if err := ruleFactory.LoadRules(*rulesFilePath); err != nil {
		log.Fatal("There was a problem loading the pricing rules for the service - ", err)
		os.Exit(1)
	}

	basketPricer := pricer.Pricer{StrategyFactory: *ruleFactory, ItemsParser: parser.ItemsParser{}}
	if err := basketPricer.LoadItems(*itemDefinitionsFilePath); err != nil {
		log.Fatal("There was a problem loading the item definitions for the service - ", err)
		os.Exit(1)
	}

	s := grpc.NewServer()
	log.Info("Registering Checkout GRPC Service")
	pb.RegisterCheckoutServer(s, &server{pricer: basketPricer})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
		os.Exit(1)
	}
}
