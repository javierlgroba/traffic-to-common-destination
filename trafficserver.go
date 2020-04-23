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

type Template struct {
    templates *template.Template
}

var travels = &map[string]*traffic.Travel{
	"Default": &traffic.Travel{
		Start: "Madrid", End: "Barcelona", By: traffic.GetTravelMode("driving")}}
var apiKey = ""

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func printDestinations(c echo.Context) error {
	travelsToPrint := traffic.QueryTravels(travels, apiKey)

	if len(*travelsToPrint)<1 {
	  return c.Render(http.StatusOK, "nodestinationtemplate", "")
	}

	travelsSlices := make([]*traffic.TravelInfoToPrint, 0, len(*travelsToPrint))
	for _, travelToPrint := range *travelsToPrint {
		travelsSlices = append(travelsSlices, travelToPrint)
	}

  return c.Render(http.StatusOK, "traffictemplate", travelsSlices)
}

func printUsage(programName string) {
  fmt.Println("Usage: ",programName," [-c configfilename]")
}

func createTravelMap(input []interface{}) *map[string]*traffic.Travel {
	result := &(map[string]*traffic.Travel{})

	for _, travel := range input {
		for k, v := range travel.(map[string]interface {}) {
			(*result)[k] = &(traffic.Travel{})
			for k1, v1 := range v.(map[string]interface {}) {
				switch k1 {
				case "Start":
					(*result)[k].Start = v1.(string)
				case "End":
					(*result)[k].End = v1.(string)
				case "By":
					(*result)[k].By = traffic.GetTravelMode(v1.(string))
				default:
					panic(fmt.Sprintf("Wrong configuration value: %s[%s]->%s\n", k, k1, v1))
				}
				fmt.Printf("%s[%s]: %s\n", k, k1, v1)
			}
		}
	}

	return result
}

func loadConfig(configFile string) bool {
  //viper.SetDefault("Travels", map[string]traffic.Travel{"Default": defaultTravel})
  //viper.SetDefault("APIKey", "")

	viper.SetConfigName(configFile)
	viper.SetConfigType("json")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		fmt.Printf("Fatal error config file: %s\n", err)
		return false
	}

	apiKey = viper.GetString("APIKey")
	travels = createTravelMap(viper.Get("Travels").([]interface{}))

  return true
}

func checkArgs() (bool, string) {
  argsWithoutProg := os.Args[1:]
  if len(argsWithoutProg)!=2 {
    fmt.Println("Using default config...")
    return true, ""
  } else {
    if argsWithoutProg[0]=="-c" {
      fmt.Printf("Reading config from %s.json\n", argsWithoutProg[1])
      return true, argsWithoutProg[1]
    } else {
      printUsage(os.Args[0])
      return false, ""
    }
  }
  return true, ""
}

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
	e.Logger.Fatal(e.Start("0.0.0.0:8080"))
}
