package main

import (
	"io"
	"net/http"
        "os"
        "fmt"
	"html/template"

	"github.com/labstack/echo"
        "github.com/spf13/viper"

        "./traffic"
)

//Struct for the templates
type Template struct {
    templates *template.Template
}

//travel maps initialized with a default travel
var travels = &map[string]*traffic.Travel{
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
	travelsToPrint := traffic.QueryTravels(travels, apiKey)

	if len(*travelsToPrint)<1 {
	  return c.Render(http.StatusOK, "nodestinationtemplate", "")
	}

  return c.Render(http.StatusOK, "traffictemplate", travelsToPrint)
}

//Prints in console the usage for the program
func printUsage(programName string) {
  fmt.Println("Usage: ",programName," [-c configfilename]")
}

//Creates a map with all the itineraries from configuration
func createTravelMap(input []interface{}) *map[string]*traffic.Travel {
	result := &(map[string]*traffic.Travel{})

	for _, travel := range input {
		for k, v := range travel.(map[string]interface {}) {
			//initialized with default values to avoid errors when loading config
			individualTravel := &(traffic.Travel{
				Start: "Madrid",
				End: "Barcelona",
				By: traffic.GetTravelMode("")})
			for k1, v1 := range v.(map[string]interface {}) {
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
			(*result)[k] = individualTravel
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

//Check program args at the beginning
func checkArgs() (bool, string) {
  argsWithoutProg := os.Args[1:]
  if len(argsWithoutProg)!=2 {
    fmt.Println("checkArgs: Using default config...")
    return true, ""
  } else {
    if argsWithoutProg[0]=="-c" {
      fmt.Printf("checkArgs: Config will be load from %s.json\n", argsWithoutProg[1])
      return true, argsWithoutProg[1]
    } else {
      printUsage(os.Args[0])
      return false, ""
    }
  }
  return true, ""
}

//main function
func main() {
  everythingOk, file := checkArgs()
  if !everythingOk {
    return
  }

  if !loadConfig(file){
    return
  }

	e := echo.New()
	t := &Template{
	    templates: template.Must(template.ParseGlob("*.tpl")),
	}
	e.Renderer = t
	e.GET("/", printDestinations)
	e.Static("/resources", "resources")
	e.File("/favicon.ico", "resources/favicon.ico")
	e.Logger.Fatal(e.Start("0.0.0.0:8025"))
}
