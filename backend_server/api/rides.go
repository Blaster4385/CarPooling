package api

import (
	"errors"
	"net/http"

	"github.com/achintya-7/car_pooling_backend/mapsApi"
	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) createRide(c *gin.Context) {
	var req models.CreateRideReq
	var result models.CreateDriverResponse

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email, "complete": false}
	err = server.collection.Ride.FindOne(c, filter).Decode(&result)
	if err == nil || result.Email != "" {
		err := errors.New("ride already exists")
		c.JSON(http.StatusConflict, errorResponse(err))
		return
	}

	placeRoute, err := mapsApi.GetRoute(req.Origin, req.Destination, server.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	originPoint := placeRoute.Points[0]
	middlePoint := placeRoute.Points[len(placeRoute.Points)/2]
	destinationPoint := placeRoute.Points[len(placeRoute.Points)-1]

	response := models.CreateRideResp{
		Origin:      req.Origin,
		Destination: req.Destination,
		Seats:       req.Seats,
		Email:       authPayload.Email,
		Phone:       authPayload.Phone,
		Price:       req.Price,
		PlaceId:     req.PlaceId,
		Timestamp:   req.Timestamp,
		Complete:    false,
		GeoJSON: primitive.M{
			"type": "MultiPoint",
			"coordinates": [][]float64{
				{originPoint.Lng, originPoint.Lat},
				{middlePoint.Lng, middlePoint.Lat},
				{destinationPoint.Lng, destinationPoint.Lat},
			},
		},
		Passengers: []models.Passenger{
			{
				RequestID: "0",
				Email:     authPayload.Email,
				Origin:    req.Origin,
				Phone:     authPayload.Phone,
				Name:      req.Name,
				OriginLat: placeRoute.Points[0].Lat,
				OriginLng: placeRoute.Points[0].Lng,
			},
		},
	}

	_, err = server.collection.Ride.InsertOne(c, response)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, response)

}

func (server *Server) deleteRide(c *gin.Context) {
	var result models.CreateRideResp

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email, "complete": false}
	err := server.collection.Ride.FindOneAndDelete(c, filter).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// TODO : send notification to all passengers that ride has been cancelled

	c.JSON(http.StatusOK, gin.H{"message": "ride deleted successfully"})
}

func (server *Server) getAllRidesDriver(c *gin.Context) {
	var result []models.CreateRideResp

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email}
	opts := options.Find().SetSort(bson.M{"timestamp": -1})

	cursor, err := server.collection.Ride.Find(c, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = cursor.All(c, &result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, []models.CreateDriverResponse{})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) getAllRidesPassenger(c *gin.Context) {
	var result []models.CreateRideResp

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"passengers.email": authPayload.Email}

	opts := options.Find().SetSort(bson.M{"timestamp": -1})

	cursor, err := server.collection.Ride.Find(c, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = cursor.All(c, &result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, []models.CreateDriverResponse{})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TODO : Check all the conditions before updating the ride, otherwise skip it
func (server *Server) updateRide(c *gin.Context) {
	var req models.UpdateRideReq
	var result models.CreateRideResp

	err := c.ShouldBindJSON(&req)
	if err != nil {
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

	update := bson.M{"$set": updateDoc}
	filter := bson.M{"email": authPayload.Email, "complete": false}
	err = server.collection.Ride.FindOneAndUpdate(c, filter, update).Decode(&result)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) completeRide(c *gin.Context) {

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email, "complete": false}
	update := bson.M{"$set": bson.M{"complete": true}}

	err := server.collection.Ride.FindOneAndUpdate(c, filter, update)
	if err.Err() != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err.Err()))
		return
	}

	// TODO : send notification to all passengers that ride has been completed

	c.JSON(http.StatusOK, gin.H{"message": "ride completed successfully"})
}

func (server *Server) getCurrentRideDriver(c *gin.Context) {
	var result models.CreateRideResp

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"email": authPayload.Email, "complete": false}
	err := server.collection.Ride.FindOne(c, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) getCurrentRidePassengers(c *gin.Context) {
	var result models.CreateRideResp

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{
		"complete": false,
		"passengers": bson.M{
			"$elemMatch": bson.M{
				"email": authPayload.Email,
			},
		},
	}

	err := server.collection.Ride.FindOne(c, filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) searchRide(c *gin.Context) {
	var result []models.CreateRideResp

	var req models.SearchRideReq

	err := c.ShouldBindUri(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	point, err := mapsApi.GetCords(req.Origin, server.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	pipeline := []bson.M{
		{
			"$geoNear": bson.M{
				"near": bson.M{
					"type":        "MultiPoint",
					"coordinates": []float64{point.Lng, point.Lat},
				},
				"maxDistance": 5000,
				"spherical":   true,
				"distanceField": "distance",
			},
		},
		{
			"$match": bson.M{
				"complete": false,
			},
		},
	}

	cursor, err := server.collection.Ride.Aggregate(c, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	defer cursor.Close(c)

	cursor.All(c, &result)

	c.JSON(http.StatusOK, result)
}
