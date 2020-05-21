//Package traffic includes the functionality to request google maps API for the
//itineraries information
package traffic

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/javierlgroba/cache"
	"googlemaps.github.io/maps"
)

//Local cache
var localCache = cache.New(5, 10)

//This struct defines a travel itinerary
type Travel struct {
	Start, End string
	By         maps.Mode
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
func calculateTraffic(steps []*maps.Step) string {
	return "green"
}

//Generates a duration in a human readable way
func calculateDuration(d time.Duration) string {
	duration := ""

	minutos := math.Floor(d.Minutes())
	if horas := math.Floor(d.Hours()); horas > 0 {
		duration += fmt.Sprintf("%d hours ", int(horas))
		minutos -= horas * 60
		if minutos > 0 {
			duration += "and "
		}
	}

	if minutos > 0 {
		duration += fmt.Sprintf("%d minutes", int(minutos))
	}

	return duration
}

//Convert the GoogleMaps response to our format
func mapsToTravelInfo(travelName string, travel *Travel, route *maps.Route) *TravelInfoToPrint {
	travelInfo := &TravelInfoToPrint{
		Name:     travelName,
		Start:    travel.Start,
		End:      travel.End,
		By:       fmt.Sprint(travel.By),
		Duration: calculateDuration(route.Legs[0].Duration),
		Color:    calculateTraffic(route.Legs[0].Steps),
		Summary:  route.Summary,
		Distance: fmt.Sprint(route.Legs[0].Distance.HumanReadable)}
	return travelInfo
}

//Interface for the queries to googleMaps
type googleMapsQuerier interface {
	googleMapsQuery(travelName string, travel *Travel, apiKey string, travels *[]*TravelInfoToPrint, sem chan int)
	QueryTravels(travels *map[string]*Travel, apiKey string) *[]*TravelInfoToPrint
}

//custom implementation of the interface
type GoogleMapsQuery struct {
	//Mutex for the response resource
	resourceMutex sync.Mutex
}

//Queries the googleMaps API and returns the information for an itinerary
func (custom *GoogleMapsQuery) googleMapsQuery(travelName string, travel *Travel, apiKey string, travels *[]*TravelInfoToPrint, sem chan int) {
	c, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		fmt.Println("googleMapsQuery error: ", err)
		<-sem
		return
	}

	//Prepare the request
	r := &maps.DirectionsRequest{
		Origin:      travel.Start,
		Destination: travel.End,
		Mode:        travel.By}
	route, _, err := c.Directions(context.Background(), r)

	//check everything before anything else
	if err != nil {
		fmt.Println("googleMapsQuery error: ", err)
		<-sem
		return
	}
	if len(route) < 1 || len(route[0].Legs) < 1 {
		fmt.Println("googleMapsQuery: Unable to calculate the route.")
		<-sem
		return
	}

	//Get the travelInfo and add it to the local cache
	travelInfo := mapsToTravelInfo(travelName, travel, &route[0])
	localCache.Add(travelName, travelInfo)

	//Append it to the answer
	custom.resourceMutex.Lock()
	*travels = append(*travels, travelInfo)
	custom.resourceMutex.Unlock()

	//Notify that the go routine si over
	<-sem
}

//Queries all the travels and returns a pointer to a slice with all the itinerary information
func (custom *GoogleMapsQuery) QueryTravels(travels *map[string]*Travel, apiKey string) *[]*TravelInfoToPrint {
	result := make([]*TravelInfoToPrint, 0, len(*travels))
	sem := make(chan int, len(*travels))

	for key, travel := range *travels {
		sem <- 1
		//let's check the cache
		var travelInfo *TravelInfoToPrint
		err, cachedTravel := localCache.Get(key)
		if err != nil {
			fmt.Println("Query for data")
			go custom.googleMapsQuery(key, travel, apiKey, &result, sem)
		} else {
			fmt.Println("Cached data")
			travelInfo = cachedTravel.(*TravelInfoToPrint)
			if travelInfo != nil {
				custom.resourceMutex.Lock()
				result = append(result, travelInfo)
				custom.resourceMutex.Unlock()
			}
			<-sem
		}
	}

	for i := 0; i < len(*travels); i++ {
		sem <- 1
	}

	return &result
}
