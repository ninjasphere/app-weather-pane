package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"time"

	"github.com/bugsnag/bugsnag-go"
	"github.com/ninjasphere/forecast/v2"
	"github.com/ninjasphere/gestic-tools/go-gestic-sdk"
	"github.com/ninjasphere/go-ninja/api"
	"github.com/ninjasphere/go-ninja/config"
	"github.com/ninjasphere/go-ninja/model"
	"github.com/ninjasphere/sphere-go-led-controller/fonts/O4b03b"
	"github.com/ninjasphere/sphere-go-led-controller/util"
)

var enableWeatherPane = config.MustBool("weather.enabled")
var weatherUpdateInterval = config.MustDuration("weather.updateInterval")
var temperatureDisplayTime = config.Duration(time.Second*5, "weather.temperatureDisplayTime")
var apiKey = config.MustString("weather.forecast.io.apiKey")

var globalSite *model.Site
var timezone *time.Location

type WeatherPane struct {
	siteModel   *ninja.ServiceClient
	site        *model.Site
	getWeather  *time.Timer
	tempTimeout *time.Timer
	temperature bool
	forecast    *forecast.Forecast
	image       util.Image
}

func NewWeatherPane(conn *ninja.Connection) *WeatherPane {

	pane := &WeatherPane{
		siteModel: conn.GetServiceClient("$home/services/SiteModel"),
		image:     util.LoadImage(util.ResolveImagePath("weather/loading.gif")),
	}

	pane.tempTimeout = time.AfterFunc(0, func() {
		pane.temperature = false
	})

	if !enableWeatherPane {
		return pane
	}

	go pane.GetWeather()

	return pane
}

func (p *WeatherPane) KeepAwake() bool {
	return false
}

func (p *WeatherPane) GetWeather() {

	var latitude, longitude string

	enableWeatherPane = false

	for {
		site := &model.Site{}
		err := p.siteModel.Call("fetch", config.MustString("siteId"), site, time.Second*5)
		if err == nil && (site.Longitude != nil || site.Latitude != nil) {
			latitude, longitude = fmt.Sprintf("%f", *site.Latitude), fmt.Sprintf("%f", *site.Longitude)
			break
		}

		log.Infof("Failed to get site, or site has no location.")

		time.Sleep(time.Second * 2)
	}

	for {

		f, err := forecast.Get(apiKey, latitude, longitude, "now", forecast.AUTO)

		if err != nil {
			log.Fatalf("Failed to get weather", err)
		} else {

			p.forecast = f

			filename := util.ResolveImagePath("weather-skycons/clear-day.gif")

			if _, err := os.Stat(filename); os.IsNotExist(err) {
				enableWeatherPane = false
				fmt.Printf("Couldn't load image for weather: %s", filename)
				bugsnag.Notify(fmt.Errorf("Unknown weather icon: %s", filename), p.forecast)
			} else {
				p.image = util.LoadImage(filename)
				enableWeatherPane = true
			}
		}

		time.Sleep(weatherUpdateInterval)
	}

}

func (p *WeatherPane) IsEnabled() bool {
	return enableWeatherPane && p.forecast.Timezone != ""
}

func (p *WeatherPane) Gesture(gesture *gestic.GestureMessage) {
	if gesture.Tap.Active() {
		log.Infof("Weather tap!")

		p.temperature = true
		p.tempTimeout.Reset(temperatureDisplayTime)
	}
}

func (p *WeatherPane) Render() (*image.RGBA, error) {
	if p.temperature {
		img := image.NewRGBA(image.Rect(0, 0, 16, 16))

		drawText := func(text string, col color.RGBA, top int) {
			width := O4b03b.Font.DrawString(img, 0, 8, text, color.Black)
			start := int(16 - width - 1)

			//spew.Dump("text", text, "width", width, "start", start)

			O4b03b.Font.DrawString(img, start, top, text, col)
		}

		today := p.forecast.Daily.Data[0]

		var min, max string
		if p.forecast.Flags.Units == "us" {
			min = fmt.Sprintf("%dF", int(today.TemperatureMin))
			max = fmt.Sprintf("%dF", int(today.TemperatureMax))
		} else {
			min = fmt.Sprintf("%dC", int(today.TemperatureMin))
			max = fmt.Sprintf("%dC", int(today.TemperatureMax))
		}

		drawText(max, color.RGBA{253, 151, 32, 255}, 3)
		drawText(min, color.RGBA{69, 175, 249, 255}, 10)

		return img, nil
	} else {
		return p.image.GetNextFrame(), nil
	}
}

func (p *WeatherPane) IsDirty() bool {
	return true
}
