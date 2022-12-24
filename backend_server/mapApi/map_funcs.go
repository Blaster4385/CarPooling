package mapApi

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/achintya-7/car_pooling_backend/util"
)

type PredictionList struct {
	Predictions []struct {
		Description string `json:"description"`
		PlaceID     string `json:"place_id"`
	} `json:"predictions"`
}

type Predictions struct {
	Description string `json:"description"`
	PlaceId     string `json:"place_id"`
}

type Response struct {
	Result struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
	} `json:"result"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Place struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type PolyPoints struct {
	Routes []struct {
		Bounds struct {
			Northeast struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"northeast"`
			Southwest struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"southwest"`
		} `json:"bounds"`
		Legs []struct {
			Steps []struct {
				StartLocation struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"start_location"`
			} `json:"steps"`
		} `json:"legs"`
	} `json:"routes"`
}

type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type Bounds struct {
	LatNE float64 `json:"latNE"`
	LngNE float64 `json:"lngNE"`
	LatSW float64 `json:"latSW"`
	LngSW float64 `json:"lngSW"`
}

type Route struct {
	Points []Point `json:"points"`
	Bounds Bounds   `json:"bounds"`
}

func GetPlacePredictions(input string, config util.Config) ([]Predictions, error) {
	location := "Northern India"

	apiUrl := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/autocomplete/json?input=%s&location=%s&maxresults=6&key=%s", url.QueryEscape(input), url.QueryEscape(location), config.MapsKey)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot get place recommendations : %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read response body : %v", err)
	}

	var predictions []Predictions

	var predictionList PredictionList
	err = json.Unmarshal(body, &predictionList)
	if err != nil {
		print("cannot unmarshal response body : %v", err)
	}

	for _, prediction := range predictionList.Predictions {
		predictions = append(predictions, Predictions{
			Description: prediction.Description,
			PlaceId:     prediction.PlaceID,
		})
	}

	return predictions, nil

}

func GetPlaceDetails(placeId string, config util.Config) (Place, error) {
	apiUrl := fmt.Sprintf("https://maps.googleapis.com/maps/api/place/details/json?placeid=%s&key=%s", url.QueryEscape(placeId), config.MapsKey)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return Place{}, fmt.Errorf("cannot get place details : %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Place{}, fmt.Errorf("cannot read response body : %v", err)
	}

	var response Response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return Place{}, fmt.Errorf("cannot unmarshal response body : %v", err)
	} else {
		fmt.Println(response)
	}

	place := Place{
		Lat: response.Result.Geometry.Location.Lat,
		Lng: response.Result.Geometry.Location.Lng,
	}

	return place, nil
}

func GetRoute(origin, destination string, config util.Config) (Route, error) {
	var route Route

	apiUrl := fmt.Sprintf("https://maps.googleapis.com/maps/api/directions/json?origin=%s&destination=%s&key=%s", url.QueryEscape(origin), url.QueryEscape(destination), config.MapsKey)

	resp, err := http.Get(apiUrl)
	if err != nil {
		return route, fmt.Errorf("cannot get route : %v", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return route, fmt.Errorf("cannot read response body : %v", err)
	}

	var polyPoints PolyPoints
	err = json.Unmarshal(body, &polyPoints)
	if err != nil {
		return route, fmt.Errorf("cannot unmarshal response body : %v", err)
	} 

	route.Bounds = Bounds{
		LatNE: polyPoints.Routes[0].Bounds.Northeast.Lat,
		LngNE: polyPoints.Routes[0].Bounds.Northeast.Lng,
		LatSW: polyPoints.Routes[0].Bounds.Southwest.Lat,
		LngSW: polyPoints.Routes[0].Bounds.Southwest.Lng,
	}

	for _, step := range polyPoints.Routes[0].Legs[0].Steps {
		route.Points = append(route.Points, Point{
			Lat: step.StartLocation.Lat,
			Lng: step.StartLocation.Lng,
		})
	}

	return route, nil
}
