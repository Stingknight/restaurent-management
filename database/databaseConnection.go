package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var Client *mongo.Client = DBInstance()

func DBInstance() *mongo.Client{
	err := godotenv.Load(".env")
	if err != nil{
		log.Fatalf("Error loading .env file: %s", err)
	}
	
	
	var MongoDb string = os.Getenv("MONGODB_DATABASE")

	fmt.Println(MongoDb)

	ctx,cancel:=context.WithTimeout(context.Background(),10*time.Second)

	defer cancel()

	client,err := mongo.Connect(ctx,options.Client().ApplyURI(MongoDb))
	if err != nil {
		log.Fatal(err)
	}

	if err=client.Ping(ctx,readpref.Primary());err!=nil{
		log.Fatal(err)
	}

	fmt.Println("Connected to database -------------------->")
	return client
}


func OpenCollection(client *mongo.Client,collectionName string)*mongo.Collection{
	var collection *mongo.Collection =client.Database("restaurent").Collection(collectionName)

	return collection
}
