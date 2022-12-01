package main

import (
	"context"
	"fmt"
	"github.com/dorukbulut/log-service/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

const (
	PORT  = "80"
	RPC   = "5001"
	MONGO = "mongodb://mongo:27017"
	GRPC  = "50001"
)

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	// connect to mongo

	mongoClient, err := connectMongo()

	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	// create a context in order to disconnect

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)

	defer cancel()

	// close connection
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	app := Config{
		Models: data.New(client),
	}
	//Register the RPC Server
	err = rpc.Register(new(RPCServer))
	if err != nil {
		panic(err)
		return
	}
	// start rpc server
	go app.rpcListen()

	// start the gRPC server
	go app.gRPCListen()
	//start the webServer
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}

}

func (app *Config) rpcListen() {
	log.Println("Starting RPC Server on port", RPC)
	listen, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", RPC))

	if err != nil {
		panic(err)
		return
	}

	defer listen.Close()

	for {
		rpcConn, err := listen.Accept()
		if err != nil {
			continue
		}

		go rpc.ServeConn(rpcConn)
	}
}

func connectMongo() (*mongo.Client, error) {
	// create connection options
	clientOptions := options.Client().ApplyURI(MONGO)
	clientOptions.SetAuth(options.Credential{
		Username: "admin",
		Password: "password",
	})

	// connect

	c, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Println("Error connecting : ", err)
		return nil, err
	}

	return c, nil
}
