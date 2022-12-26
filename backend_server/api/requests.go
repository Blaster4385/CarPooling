package api

import (
	"net/http"

	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (server *Server) createRequest(c *gin.Context) {
	var req models.ToDriverReq
	var result models.CreateRideResp
	var resp models.ToDriverRes

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{
		"_id":       req.RideId,
		"completed": false,
		"seats": bson.M{
			"$gte": bson.M{"$size": "$passengers"},
		},
	}

	err := server.collection.Ride.FindOne(c, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp = models.ToDriverRes{
		RideId:   req.RideId,
		Email:    authPayload.Email,
		Phone:    authPayload.Phone,
		Name:     req.Name,
		Origin:  result.Origin,
		OriginLat: req.OriginLat,
		OriginLng: req.OriginLng,
	}

	_, err = server.collection.Request.InsertOne(c, resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)

}

// func (server *Server) getRideRequests(c *gin.Context) {
// 	var result models.ToDriverReq

// 	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

// 	pipeline := []bson.M{
// 		bson.M{
// 			"$lookup": bson.M{
// 				"from":         "rides",
// 				"localField":   "ride_id",
// 				"foreignField": "_id",
// 		}

	

	
// }
