package models

type CreateRideReq struct {
	Origin      string `json:"origin" binding:"required"`
	Destination string `json:"destination" binding:"required"`
	Seats       int    `json:"seats" binding:"required"`
	Price       int    `json:"price" binding:"required"`
	PlaceId     string `json:"place_id" binding:"required"`
	Timestamp   int64  `json:"timestamp" binding:"required"`
}

type Rider struct {
	Email     string  `json:"email" binding:"required"`
	Phone     string  `json:"phone" binding:"required"`
	Name      string  `json:"name" binding:"required"`
	Origin    string  `json:"origin" binding:"required"`
	OriginLat float64 `json:"origin_lat" binding:"required"`
	OriginLng float64 `json:"origin_lng" binding:"required"`
}

type CreateRideResp struct {
	Id          string  `json:"id" bson:"_id,omitempty"`
	Origin      string  `json:"origin" binding:"required"`
	Destination string  `json:"destination" binding:"required"`
	Seats       int     `json:"seats" binding:"required"`
	Price       int     `json:"price" binding:"required"`
	PlaceId     string  `json:"place_id" binding:"required"`
	OriginLat   float64 `json:"origin_lat" binding:"required"`
	OriginLng   float64 `json:"origin_lng" binding:"required"`
	Email       string  `json:"email" binding:"required"`
	Phone       string  `json:"phone" binding:"required"`
	Timestamp   int64   `json:"timestamp" binding:"required"`
	Riders      []Rider `json:"riders"`
	Complete    bool    `json:"complete"`
}

type UpdateRideReq struct {
	Price     int   `json:"price" binding:"required"`
	Timestamp int64 `json:"timestamp" binding:"required"`
}

type CompleteRideReq struct {
	Complete bool `json:"complete" binding:"required"`
}
