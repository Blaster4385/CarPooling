package util

import "go.mongodb.org/mongo-driver/mongo"

type Collection struct {
	driver *mongo.Collection
	passenger *mongo.Collection
	rides *mongo.Collection
}

func NewCollection(client *mongo.Client, config Config) *Collection {
	return &Collection{
		driver: client.Database(config.DBName).Collection("drivers"),
		passenger: client.Database(config.DBName).Collection("passengers"),
		rides: client.Database(config.DBName).Collection("rides"),
	}
}