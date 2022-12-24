package util

import (
	"go.mongodb.org/mongo-driver/mongo"
)

type Collection struct {
	Driver    *mongo.Collection
	Passenger *mongo.Collection
	Rides     *mongo.Collection
}

func NewCollection(client *mongo.Client, config Config) Collection {
	return Collection{
		Driver:    client.Database(config.DBName).Collection("drivers"),
		Passenger: client.Database(config.DBName).Collection("passengers"),
		Rides:     client.Database(config.DBName).Collection("rides"),
	}
}
