package models

type ToDriverReq struct {
	Name      string  `json:"name" binding:"required"`
	RideId    string  `json:"ride_id" bson:"_id,omitempty"`
	Origin    string  `json:"origin"`
	OriginLat float64 `json:"origin_lat"`
	OriginLng float64 `json:"origin_lng"`
}

type ToDriverRes struct {
	Name      string  `json:"name" binding:"required"`
	RideId    string  `json:"ride_id" bson:"_id,omitempty"`
	Origin    string  `json:"origin"`
	OriginLat float64 `json:"origin_lat"`
	OriginLng float64 `json:"origin_lng"`
	Email     string  `json:"email"`
	Phone     string  `json:"phone"`
}
