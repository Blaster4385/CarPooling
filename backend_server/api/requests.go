package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/achintya-7/car_pooling_backend/token"
	"github.com/achintya-7/car_pooling_backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (server *Server) createRequest(c *gin.Context) {
	var req models.RequestToDriverReq

	var result models.CreateRideResp
	var rideResult models.CreateRideResp
	var resp models.RequestToDriverRes
	var notification models.NotificationModel

	errChan1 := make(chan error)
	errChan2 := make(chan error)

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	objectId, err := utils.StringToObjectId(req.RideId)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// * check if user has already requested a ride
	go func() {
		filter := bson.M{
			"complete":         false,
			"passengers.email": authPayload.Email,
		}

		err = server.collection.Ride.FindOne(c, filter).Decode(&rideResult)

		errChan1 <- err
	}()

	// * check if ride exists and has seats available
	go func() {
		filter2 := bson.M{
			"_id":      objectId,
			"complete": false,
			"$expr": bson.M{
				"$gt": bson.A{"$seats", bson.M{"$size": "$passengers"}},
			},
		}

		err = server.collection.Ride.FindOne(c, filter2).Decode(&result)

		errChan2 <- err
	}()

	err1 := <-errChan1
	err2 := <-errChan2

	if err1 == nil || rideResult.Email != "" {
		err := errors.New("ride already exists")
		c.JSON(http.StatusConflict, errorResponse(err))
		return
	}

	if err2 != nil {
		if err2 == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(err2))
			return
		}
		c.JSON(http.StatusConflict, errorResponse(err2))
		return
	}

	// * create response now
	ride_id, err := primitive.ObjectIDFromHex(req.RideId)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	resp = models.RequestToDriverRes{
		ID:        primitive.NewObjectID(),
		RideId:    ride_id,
		Email:     authPayload.Email,
		Phone:     authPayload.Phone,
		Name:      req.Name,
		Origin:    result.Origin,
		OriginLat: req.OriginLat,
		OriginLng: req.OriginLng,
		Timestamp: time.Now().Unix(),
		Status:    0,
	}

	_, err = server.collection.Request.InsertOne(c, resp)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// TODO : send notification to driver that a request has been made
	msg := fmt.Sprintf("New request from %s has been made for the Ride %s", authPayload.Name, req.RideId)

	notification = models.NotificationModel{
		Email:       result.Email,
		SenderPhone: authPayload.Phone,
		SenderName:  authPayload.Name,
		Type:        0,
		Content:     msg,
		Timestamp:   time.Now().Unix(),
	}

	_, err = server.collection.Notification.InsertOne(c, notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, resp)

}

func (server *Server) getRideRequestsDriver(c *gin.Context) {
	var req models.DriverRidesReq
	var result []models.RequestToDriverRes

	if err := c.ShouldBindUri(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// * Get all requests for a ride using ride_id
	ride_id, err := primitive.ObjectIDFromHex(req.RideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	filter := bson.M{"ride_id": ride_id}

	cursor, err := server.collection.Request.Find(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if err = cursor.All(c, &result); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if len(result) == 0 {
		c.JSON(http.StatusNotFound, []models.RequestToDriverRes{})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) getRideRequestsPassenger(c *gin.Context) {

	var result []models.RequestToDriverRes

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	// * check all rides created by user which are less than 12 hours old
	now_time_stamp := time.Now().Unix() - 3600*12

	filter := bson.M{"email": authPayload.Email, "timestamp": bson.M{"$gt": now_time_stamp}}

	cursor, err := server.collection.Request.Find(c, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if err = cursor.All(c, &result); err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, result)
}

func (server *Server) acceptRideRequest(c *gin.Context) {
	var req models.Passenger

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	session, err := server.client.StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	defer session.EndSession(c)

	// * check if ride exists and has seats available and then push in the passenger data
	filter := bson.M{
		"email":    authPayload.Email,
		"complete": false,
		"$expr": bson.M{
			"$gt": bson.A{"$seats", bson.M{"$size": "$passengers"}},
		}}
	update := bson.M{"$push": bson.M{"passengers": req}}
	options := options.FindOneAndUpdate().SetReturnDocument(options.After)

	result := server.collection.Ride.FindOneAndUpdate(c, filter, update, options)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(result.Err()))
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse(result.Err()))
		return
	}

	// * update the status of the request to accepted
	request_id, err := utils.StringToObjectId(req.RequestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	filter2 := bson.M{"_id": request_id}
	update2 := bson.M{"$set": bson.M{"status": 1}}

	updateResult := server.collection.Request.FindOneAndUpdate(c, filter2, update2)
	if updateResult.Err() != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// * Delete all other requests for the passenger
	filter3 := bson.M{"email": req.Email}
	_, err = server.collection.Request.DeleteMany(c, filter3)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// TODO : Send notification to the passenger that the request has been accepted
	msg := fmt.Sprintf("Your request for ride has been ACCEPTED by the driver %s", authPayload.Name)

	notification := models.NotificationModel{
		Email:       req.Email,
		Content:     msg,
		Timestamp:   time.Now().Unix(),
		SenderPhone: authPayload.Phone,
		SenderName:  authPayload.Name,
		Type:        1,
	}
	_, err = server.collection.Notification.InsertOne(c, notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request accepted successfully"})

}

func (server *Server) rejectRideRequest(c *gin.Context) {
	var req models.ModifyRideRequestReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	request_id, err := utils.StringToObjectId(req.RequestID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ride_id, err := utils.StringToObjectId(req.RideID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// * check if ride exists of the ID and the user is the owner of the ride
	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)
	checkFilter := bson.M{"_id": ride_id, "email": authPayload.Email}
	result := server.collection.Ride.FindOne(c, checkFilter)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(result.Err()))
			return
		}
		c.JSON(http.StatusInternalServerError, errorResponse(result.Err()))
		return
	}

	// * update the status of the request to rejected
	filter := bson.M{"_id": request_id}
	update := bson.M{"$set": bson.M{"status": 2}}

	result = server.collection.Request.FindOneAndUpdate(c, filter, update)
	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(result.Err()))
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse(result.Err()))
		return
	}

	// TODO : Send notification to the passenger that the request has been rejected
	msg := fmt.Sprintf("Your request for ride has been DECLINED by the driver %s", authPayload.Name)

	notification := models.NotificationModel{
		Email:       req.Email,
		Content:     msg,
		Timestamp:   time.Now().Unix(),
		SenderPhone: authPayload.Phone,
		SenderName:  authPayload.Name,
		Type:        2,
	}

	_, err = server.collection.Notification.InsertOne(c, notification)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request Declined Successfully"})
}

func (server *Server) deleteRideRequest(c *gin.Context) {
	var req models.DeleteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// * Delete the request
	request_id, err := utils.StringToObjectId(req.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := c.MustGet(authorizationPayloadKey).(*token.Payload)

	filter := bson.M{"_id": request_id, "email": authPayload.Email}
	result := server.collection.Request.FindOneAndDelete(c, filter)
	if err != nil {
		if result.Err() == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, errorResponse(result.Err()))
			return
		}

		c.JSON(http.StatusInternalServerError, errorResponse(result.Err()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Request Deleted Successfully"})
}
