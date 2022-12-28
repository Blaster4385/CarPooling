package api

import (
	"errors"
	"net/http"

	"github.com/achintya-7/car_pooling_backend/mapsApi"
	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) createRide(c *gin.Context) {
	var req models.CreateRideReq
	var result models.CreatePassengerResponse
	var result2 models.CreateDriverResponse

	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter2 := bson.M{"email": authPayload.Email, "complete": false}
	err = server.collection.Ride.FindOne(c, filter2).Decode(&result2)
	if err == nil || result2.Email != "" {
		err := errors.New("ride already exists")
		c.JSON(http.StatusConflict, errorResponse(err))
		return
	}

	placeRoute, err := mapsApi.GetRoute(req.Origin, req.Destination, server.config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	response := models.CreateRideResp{
		Origin:      req.Origin,
		Destination: req.Destination,
		Seats:       req.Seats,
		Email:       result.Email,
		Phone:       result.Phone,
		OriginLat:   placeRoute.Points[0].Lat,
		OriginLng:   placeRoute.Points[0].Lng,
		Price:       req.Price,
		PlaceId:     req.PlaceId,
		Timestamp:   req.Timestamp,
		Complete:    false,
		Passengers: []models.Passenger{
			{
				RequestID: "0",
				Email:     result.Email,
				Origin:    req.Origin,
				Phone:     result.Phone,
				Name:      result.Name,
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

// func (server *Server) searchRides(c *gin.Context) {

// 	var req models.SearchRideReq
// 	var result []models.CreateRideResp

// 	err := c.ShouldBindJSON(&req); if err != nil {
// 		c.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	// Find all the points from the source to destination
// 	route, err := mapsApi.GetRoute(req.Origin, req.Destination, server.config)
// 	if err != nil {
// 		c.JSON(http.StatusExpectationFailed, errorResponse(err))
// 		return
// 	}

// 	var queryPoints []mapsApi.Point

// 	queryPoints = append(queryPoints, route.Points[0])
// 	queryPoints = append(queryPoints, route.Points[len(route.Points)/2])
// 	queryPoints = append(queryPoints, route.Points[len(route.Points)-1])

// 	// Query the request database for all the rides which have 3 points in their path
// 	// and the points are within 2km of the points in the queryPoints
// 	filter := bson.M{
// 		"complete": false,
// 		"origin": bson.M{
// 			"$near": bson.M{
// 				"$geometry": bson.M{
// 					"type": "Point",
// 					"coordinates": []float64{
// 						queryPoints[0].Lng,
// 						queryPoints[0].Lat,
// 					},
// 				},
// 				"$maxDistance": 2000,
// 			},
// 		},
// 		"destination": bson.M{
// 			"$near": bson.M{
// 				"$geometry": bson.M{
// 					"type": "Point",
// 					"coordinates": []float64{
// 						queryPoints[2].Lng,
// 						queryPoints[2].Lat,
// 					},
// 				},
// 				"$maxDistance": 2000,
// 			},
// 		},
// 	}



// 	// Send the list of rides to the user

// }