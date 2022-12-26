package api

import (
	"fmt"

	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/achintya-7/car_pooling_backend/util"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

type Server struct {
	config     util.Config
	router     *gin.Engine
	tokenMaker token.Maker
	client     *mongo.Client
	collection util.Collection
}

func NewServer(config util.Config, client *mongo.Client) (*Server, error) {
	tokenMaker, err := token.NewPasetoMaker(config.TokenSecret)
	if err != nil {
		return nil, fmt.Errorf("cannot create token make : %v", err)
	}

	collection := util.NewCollection(client, config)

	server := &Server{
		config:     config,
		tokenMaker: tokenMaker,
		client:     client,
		collection: collection,
	}

	server.setupRoutes()
	return server, nil
}

func (server *Server) setupRoutes() {
	r := gin.Default()
	defer func() {
		server.router = r
	}()

	r.GET("/", server.entryPoint)
	r.POST("/passengers", server.createPassenger)

	// * AUTHENTICATION
	authRoute := r.Group("/").Use(authMiddleware(server.tokenMaker))

	// * PASSENGERS
	authRoute.PUT("/passengers", server.updatePassenger)
	authRoute.GET("/passengers", server.getPassenger)

	// * DRIVERS
	authRoute.POST("/drivers", server.cretaeDriver)
	authRoute.PUT("/drivers", server.updateDriver)
	authRoute.GET("/drivers", server.getDriver)

	// * API
	authRoute.GET("/api/placePredictions/:place", server.placePredicions)
	authRoute.POST("/api/route", server.placeRoute)

	// * RIDES
	authRoute.POST("/rides", server.createRide)
	authRoute.DELETE("/rides", server.deleteRide)
	authRoute.GET("/rides/all", server.getAllRides)
	authRoute.PUT("/rides", server.updateRide)
	authRoute.GET("/rides", server.getCurrentRide)
	authRoute.GET("rides/complete", server.completeRide)
	// authRoute.GET("/rides/requests", server.getRideRequests)

	// * REQUESTS
	authRoute.POST("/requests", server.createRequest)
	// authRoute.GET("/requests", server.getRequest)

}

func (server *Server) Start(serverAddress string) error {
	return server.router.Run(serverAddress)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
