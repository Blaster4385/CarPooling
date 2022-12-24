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
	collection *util.Collection
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

}

func (server *Server) Start(serverAddress string) error {
	return server.router.Run(serverAddress)
}

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
