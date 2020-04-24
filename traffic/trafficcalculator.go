//Package traffic includes the functionality to request google maps API for the
//itineraries information
package traffic

import (
	"fmt"
	"context"
	"time"
	"math"

	"googlemaps.github.io/maps"
)

//This struct defines a travel itinerary
type Travel struct {
	Start, End string
	By maps.Mode
}

//This struct is used to store all the information from a itinerary response
type TravelInfoToPrint struct {
	Name, Start, End, By, Duration, Color, Summary, Distance string
}

//Using a string returns the travel mode in the googlemaps format.
//Always returns driving by default
func GetTravelMode(by string) maps.Mode {
	switch by {
	case "bicycling":
		return maps.TravelModeBicycling
	case "transit":
		return maps.TravelModeTransit
	case "walking":
		return maps.TravelModeWalking
	default:
		return maps.TravelModeDriving
	}
}

//This function returns a colour depending on the traffic
//Can't test now because of COVID19, returns always green
func calculateTraffic(steps[] *maps.Step) string {
	return "green"
}

//Generates a duration in a human readable way
func calculateDuration(d time.Duration) string {
	duration := ""

	minutos := math.Floor(d.Minutes())
	if horas := math.Floor(d.Hours()); horas>0 {
		duration += fmt.Sprintf("%d hours ", int(horas))
		minutos -= horas * 60
		if minutos >0 {
			duration += "and "
		}
	}

	if minutos>0 {
		duration += fmt.Sprintf("%d minutes", int(minutos))
	}

	return duration
}

//Queries the googleMaps API and returns the information for an itinerary
func googleMapsQuery(travel *Travel, apiKey string) *TravelInfoToPrint {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		fmt.Println("googleMapsQuery error: ", err)
		return nil
	}
	r := &maps.DirectionsRequest{
		Origin:      travel.Start,
		Destination: travel.End,
		Mode:	 travel.By	}
	route, _, err := c.Directions(context.Background(), r)
	if err != nil {
		fmt.Println("googleMapsQuery error: ", err)
		return nil
	}

	if len(route)<1 || len(route[0].Legs)<1{
		fmt.Println("googleMapsQuery: Unable to calculate the route.")
		return nil
	}

	return &TravelInfoToPrint{
		Start: travel.Start,
		End: travel.End,
		By: fmt.Sprint(travel.By),
		Duration: calculateDuration(route[0].Legs[0].Duration),
		Color: calculateTraffic(route[0].Legs[0].Steps),
		Summary: route[0].Summary,
		Distance: fmt.Sprint(route[0].Legs[0].Distance.HumanReadable)}
}

//Queries all the travels and returns a pointer to a slice with all the itinerary information
func QueryTravels(travels *map[string]*Travel, apiKey string) *[]*TravelInfoToPrint {
	result := make([]*TravelInfoToPrint, 0, len(*travels))
	for key, travel := range *travels {
    travelInfo := googleMapsQuery(travel, apiKey)
		if travelInfo!=nil{
			travelInfo.Name = key
			result = append(result, travelInfo)
		}
	}
	return &result
}
