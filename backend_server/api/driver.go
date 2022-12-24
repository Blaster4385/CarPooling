package api

import (
	"net/http"

	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (server *Server) cretaeDriver(c *gin.Context) {
	var req models.CreateDriverRequest
	var result models.CreatePassengerResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email}

	err := server.collection.Passenger.FindOne(c, filter).Decode(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if authPayload.Email != result.Email {
		c.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	resp := models.CreateDriverResponse{
		Car:        req.Car,
		Seats:      req.Seats,
		Experience: req.Experience,
		Email:      result.Email,
		Phone:      result.Phone,
		Name:       result.Name,
	}

	_, err = server.collection.Driver.InsertOne(c, resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)

}

func (server *Server) updateDriver(c *gin.Context) {
	var req models.UpdateDriverRequest
	var result models.CreateDriverResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	pByte, err := bson.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	var updateDoc bson.M

	err = bson.Unmarshal(pByte, &updateDoc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	filter := bson.M{"email": authPayload.Email}
	update := bson.M{"$set": updateDoc}

	if err := server.collection.Driver.FindOneAndUpdate(c, filter, update).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}
