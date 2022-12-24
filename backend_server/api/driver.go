package api

import (
	"fmt"
	"net/http"

	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func (server *Server) cretaeDriver(c *gin.Context) {
	var req models.CreateDriverRequest
	var result models.CreatePassengerResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	fmt.Println(authPayload.Email)
	filter := bson.M{"user.email": authPayload.Email}

	err := server.collection.Passenger.FindOne(c, filter).Decode(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if authPayload.Email != result.User.Email {
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	resp := models.CreateDriverResponse{
		Car:        req.Car,
		Seats:      req.Seats,
		Experience: req.Experience,
		Email:      result.User.Email,
		Phone:      result.User.Phone,
		Name:       result.User.Name,
	}

	_, err = server.collection.Driver.InsertOne(c, resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)

}
