package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/reflection"
)

var collection *mongo.Collection

func main() {
	fmt.Println("Run Checklist server")
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)
	addr := "0.0.0.0:10000"

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("Error create client for mongodb: %v", err)
	}

	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal("Can not connect to mongodb: %v", err)
	}

	collection = client.Database("checklistDB").Collection("checklist")

	// Listen for grpc
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Failed to listen: %v", err)
	}
	
	s := grpc.NewServer(
		// grpc.Creds(credentials.NewServerTLSFromCert(&insecure.Cert)),
	)

	checklistServer(s)

	reflection.Register(s)

	go func() {
		if err := s.Serve(listen); err != nil {
			log.Fatal("fail to start server: %v", err)
		}
	}()

	// err = gateway.Run("dns:///" + addr)
	// log.Fatalf("Gate Error %v", err)

	ch := make(chan os.Signal, 1)

	signal.Notify(ch, os.Interrupt)

	<-ch
	fmt.Println("Stopping the server.")
	s.Stop()
	fmt.Println("Closing the listener.")
	listen.Close()
	fmt.Println("Closing MongoDB Connection")
	client.Disconnect(context.TODO())
	fmt.Println("Program down gracefully.")
}
