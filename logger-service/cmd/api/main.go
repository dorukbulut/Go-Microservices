package main

import (
	"context"
	"fmt"
	"github.com/dorukbulut/log-service/data"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
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
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", PORT),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
	//go app.serve()
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
