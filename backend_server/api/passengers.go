package api

import (
	"net/http"
	"strings"

	"github.com/achintya-7/car_pooling_backend/models"
	"github.com/gin-gonic/gin"
)

func (server *Server) createPassenger(c *gin.Context) {
	var req models.CreatePassengerRequest
	var resp models.CreatePassengerResponse

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(req.Email, server.config.AccessTokenDuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	resp.User = req
	resp.Token = accessToken

	_, err = server.collection.Driver.InsertOne(c, resp)
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
