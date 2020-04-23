package traffic

import (
	"fmt"
	"context"
	"time"
	"math"

	"googlemaps.github.io/maps"
	"github.com/kr/pretty"
)


//DRIVING (Default) indicates standard driving directions using the road network.
//BICYCLING requests bicycling directions via bicycle paths & preferred streets.
//TRANSIT requests directions via public transit routes.
//WALKING requests walking directions via pedestrian paths & sidewalks.
type Travel struct {
	Start, End string
	By maps.Mode
}

type TravelInfoToPrint struct {
	Name, Start, End, By, Duration, Color, Summary, Distance string
}

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

func calculateTraffic(steps[] *maps.Step) string {
	return "green"
}

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

func googleMapsQuery(travel *Travel, apiKey string) *TravelInfoToPrint {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		fmt.Println("Fatal error: ", err)
		return &TravelInfoToPrint{}
	}
	r := &maps.DirectionsRequest{
		Origin:      travel.Start,
		Destination: travel.End,
		Mode:	 travel.By	}
	route, _, err := c.Directions(context.Background(), r)
	if err != nil {
		fmt.Println("Fatal error: ", err)
		return &TravelInfoToPrint{}
	}

	if len(route)<1 || len(route[0].Legs)<1{
		fmt.Println("Unable to calculate the route.")
		return &TravelInfoToPrint{}
	}

	travelInfo := &TravelInfoToPrint{}
	travelInfo.Start = travel.Start
	travelInfo.End = travel.End
	travelInfo.By = pretty.Sprint(travel.By)
	travelInfo.Duration = calculateDuration(route[0].Legs[0].Duration)
	travelInfo.Color = calculateTraffic(route[0].Legs[0].Steps)
	travelInfo.Summary = route[0].Summary
	travelInfo.Distance = pretty.Sprint(route[0].Legs[0].Distance.HumanReadable)

	return travelInfo
}

func QueryTravels(travels *map[string]*Travel, apiKey string) *map[string]*TravelInfoToPrint {
	result := &(map[string]*TravelInfoToPrint{})
	for key, travel := range *travels {
    (*result)[key] = googleMapsQuery(travel, apiKey)
		(*result)[key].Name = key
	}
	return result
}
