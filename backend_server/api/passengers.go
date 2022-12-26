package api

import (
	"net/http"
	"strings"

	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) createPassenger(c *gin.Context) {
	var req models.CreatePassengerRequest
	var resp models.CreatePassengerResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(req.Email, req.Phone, server.config.AccessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &resp,
		TagName: "json",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = decoder.Decode(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp.Token = accessToken

	_, err = server.collection.Passenger.InsertOne(c, resp)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key error") {
			c.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)

}

func (server *Server) updatePassenger(c *gin.Context) {
	var req models.UpdatePassengerRequest
	var result models.CreatePassengerResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	token, err := server.tokenMaker.CreateToken(req.Email, req.Phone, server.config.AccessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

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

	updateDoc["token"] = token

	filter := bson.M{"email": authPayload.Email}
	update := bson.M{
		"$set": updateDoc,
	}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	if err := server.collection.Passenger.FindOneAndUpdate(c, filter, update, options).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) getPassenger(c *gin.Context) {
	var result models.CreatePassengerResponse

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email}

	if err := server.collection.Passenger.FindOne(c, filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}
