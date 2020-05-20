package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"traffic-to-common-destination/traffic"

	"github.com/labstack/echo"
	"github.com/spf13/viper"
)

//Struct for the templates
type Template struct {
	templates *template.Template
}

//travel maps initialized with a default travel
var travels = map[string]*traffic.Travel{
	"Default": &traffic.Travel{
		Start: "Madrid", End: "Barcelona", By: traffic.GetTravelMode("driving")}}

//Api key for googleMaps API
var apiKey = ""

//Implementation used by Echo to render the templates
func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

//Returns the page with all the itinerary information included
func printDestinations(c echo.Context) error {
	travelsToPrint := traffic.QueryTravels(&travels, apiKey)

	if len(*travelsToPrint) < 1 {
		return c.Render(http.StatusOK, "nodestinationtemplate", "")
	}

	return c.Render(http.StatusOK, "traffictemplate", travelsToPrint)
}

//Add a destination
func addDestination(c echo.Context) error {
	name := c.QueryParam("Name")
	start := c.QueryParam("Start")
	end := c.QueryParam("End")
	by := c.QueryParam("By")

	if name == "" || start == "" || end == "" {
		return c.Render(http.StatusBadRequest, "operationtemplate", "Impossible to add the destination!")
	}

	travels[name] = &traffic.Travel{Start: start, End: end, By: traffic.GetTravelMode(by)}

	return c.Render(http.StatusCreated, "operationtemplate", "Destination added!")
}

//Func remove a destination
func removeDestination(c echo.Context) error {
	name := c.QueryParam("Name")

	if name != "" {
		delete(travels, name)

		return c.Render(http.StatusAccepted, "operationtemplate", "Destination removed!")
	}

	return c.Render(http.StatusBadRequest, "operationtemplate", "Impossible to remove destination!")
}

//Prints in console the usage for the program
func printUsage(programName string) {
	fmt.Println("Usage: ", programName, " [-c configfilename]")
}

//Creates a map with all the itineraries from configuration
func createTravelMap(input []interface{}) map[string]*traffic.Travel {
	result := map[string]*traffic.Travel{}

	for _, travel := range input {
		for k, v := range travel.(map[string]interface{}) {
			//initialized with default values to avoid errors when loading config
			individualTravel := &(traffic.Travel{
				Start: "Madrid",
				End:   "Barcelona",
				By:    traffic.GetTravelMode("")})
			for k1, v1 := range v.(map[string]interface{}) {
				switch k1 {
				case "Start":
					individualTravel.Start = v1.(string)
				case "End":
					individualTravel.End = v1.(string)
				case "By":
					individualTravel.By = traffic.GetTravelMode(v1.(string))
				default:
					panic(fmt.Sprintf("Wrong configuration value: %s[%s]->%s\n", k, k1, v1))
				}
				fmt.Printf("%s[%s]: %s\n", k, k1, v1)
			}
			result[k] = individualTravel
		}
	}

	return result
}

//Load the configuration from a file
func loadConfig(configFile string) bool {
	viper.SetConfigName(configFile)
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		fmt.Printf("loadConfig: Error loading config file: %s.json\n", err)
		return false
	}

	apiKey = viper.GetString("APIKey")
	travels = createTravelMap(viper.Get("Travels").([]interface{}))

	return true
}

//main function
func main() {
	file := ""
	flag.StringVar(&file, "c", "default.json", "Config file for the server.")
	flag.Parse()

	if file == "" {
		return
	}

	if !loadConfig(file) {
		return
	}

	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("*.tpl")),
	}
	e.Renderer = t
	e.GET("/", printDestinations)
	e.POST("/addTravel", addDestination)
	e.DELETE("/deleteTravel", removeDestination)
	e.Static("/resources", "resources")
	e.File("/favicon.ico", "resources/favicon.ico")
	e.Logger.Fatal(e.Start("0.0.0.0:8025"))
}
