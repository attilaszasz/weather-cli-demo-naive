package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type apiResponse struct {
	Latitude             float64           `json:"latitude"`
	Longitude            float64           `json:"longitude"`
	GenerationTimeMS     float64           `json:"generationtime_ms"`
	UTCOffsetSeconds     int               `json:"utc_offset_seconds"`
	Timezone             string            `json:"timezone"`
	TimezoneAbbreviation string            `json:"timezone_abbreviation"`
	Elevation            float64           `json:"elevation"`
	CurrentUnits         map[string]string `json:"current_units"`
	Current              currentWeather    `json:"current"`
}

type currentWeather struct {
	Time               string  `json:"time"`
	Interval           int     `json:"interval"`
	Temperature2M      float64 `json:"temperature_2m"`
	RelativeHumidity2M int     `json:"relative_humidity_2m"`
	ApparentTemperature float64 `json:"apparent_temperature"`
	IsDay              int     `json:"is_day"`
	Precipitation      float64 `json:"precipitation"`
	Rain               float64 `json:"rain"`
	Showers            float64 `json:"showers"`
	Snowfall           float64 `json:"snowfall"`
	WeatherCode        int     `json:"weather_code"`
	CloudCover         int     `json:"cloud_cover"`
	PressureMSL        float64 `json:"pressure_msl"`
	SurfacePressure    float64 `json:"surface_pressure"`
	WindSpeed10M       float64 `json:"wind_speed_10m"`
	WindDirection10M   int     `json:"wind_direction_10m"`
	WindGusts10M       float64 `json:"wind_gusts_10m"`
}

type output struct {
	Latitude  float64      `json:"latitude"`
	Longitude float64      `json:"longitude"`
	Timezone  string       `json:"timezone"`
	Current   outputCurrent `json:"current"`
}

type outputCurrent struct {
	Time                string  `json:"time"`
	Temperature         float64 `json:"temperature_c"`
	RelativeHumidity    int     `json:"relative_humidity_percent"`
	ApparentTemperature float64 `json:"apparent_temperature_c"`
	IsDay               bool    `json:"is_day"`
	Precipitation       float64 `json:"precipitation_mm"`
	Rain                float64 `json:"rain_mm"`
	Showers             float64 `json:"showers_mm"`
	Snowfall            float64 `json:"snowfall_cm"`
	CloudCover          int     `json:"cloud_cover_percent"`
	PressureMSL         float64 `json:"pressure_msl_hpa"`
	SurfacePressure     float64 `json:"surface_pressure_hpa"`
	WindSpeed           float64 `json:"wind_speed_kmh"`
	WindDirection       int     `json:"wind_direction_degrees"`
	WindGusts           float64 `json:"wind_gusts_kmh"`
	WeatherCode         int     `json:"weather_code"`
	WeatherDescription  string  `json:"weather_description"`
}

func main() {
	latFlag := flag.String("lat", "", "latitude")
	lonFlag := flag.String("lon", "", "longitude")
	flag.Parse()

	if *latFlag == "" || *lonFlag == "" {
		fail(errors.New("both -lat and -lon are required"))
	}

	lat, err := strconv.ParseFloat(*latFlag, 64)
	if err != nil {
		fail(fmt.Errorf("invalid -lat value: %w", err))
	}

	lon, err := strconv.ParseFloat(*lonFlag, 64)
	if err != nil {
		fail(fmt.Errorf("invalid -lon value: %w", err))
	}

	result, err := fetchWeather(lat, lon)
	if err != nil {
		fail(err)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		fail(fmt.Errorf("failed to encode output: %w", err))
	}
}

func fetchWeather(lat, lon float64) (*output, error) {
	endpoint, err := url.Parse("https://api.open-meteo.com/v1/forecast")
	if err != nil {
		return nil, fmt.Errorf("failed to parse API URL: %w", err)
	}

	query := endpoint.Query()
	query.Set("latitude", strconv.FormatFloat(lat, 'f', -1, 64))
	query.Set("longitude", strconv.FormatFloat(lon, 'f', -1, 64))
	query.Set("current", "temperature_2m,relative_humidity_2m,apparent_temperature,is_day,precipitation,rain,showers,snowfall,weather_code,cloud_cover,pressure_msl,surface_pressure,wind_speed_10m,wind_direction_10m,wind_gusts_10m")
	query.Set("timezone", "auto")
	endpoint.RawQuery = query.Encode()

	client := &http.Client{Timeout: 15 * time.Second}
	response, err := client.Get(endpoint.String())
	if err != nil {
		return nil, fmt.Errorf("weather request failed: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %s", response.Status)
	}

	var payload apiResponse
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode weather response: %w", err)
	}

	return &output{
		Latitude:  payload.Latitude,
		Longitude: payload.Longitude,
		Timezone:  payload.Timezone,
		Current: outputCurrent{
			Time:                payload.Current.Time,
			Temperature:         payload.Current.Temperature2M,
			RelativeHumidity:    payload.Current.RelativeHumidity2M,
			ApparentTemperature: payload.Current.ApparentTemperature,
			IsDay:               payload.Current.IsDay == 1,
			Precipitation:       payload.Current.Precipitation,
			Rain:                payload.Current.Rain,
			Showers:             payload.Current.Showers,
			Snowfall:            payload.Current.Snowfall,
			CloudCover:          payload.Current.CloudCover,
			PressureMSL:         payload.Current.PressureMSL,
			SurfacePressure:     payload.Current.SurfacePressure,
			WindSpeed:           payload.Current.WindSpeed10M,
			WindDirection:       payload.Current.WindDirection10M,
			WindGusts:           payload.Current.WindGusts10M,
			WeatherCode:         payload.Current.WeatherCode,
			WeatherDescription:  describeWeatherCode(payload.Current.WeatherCode),
		},
	}, nil
}

func fail(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func describeWeatherCode(code int) string {
	switch code {
	case 0:
		return "clear sky"
	case 1:
		return "mainly clear"
	case 2:
		return "partly cloudy"
	case 3:
		return "overcast"
	case 45:
		return "fog"
	case 48:
		return "depositing rime fog"
	case 51:
		return "light drizzle"
	case 53:
		return "moderate drizzle"
	case 55:
		return "dense drizzle"
	case 56:
		return "light freezing drizzle"
	case 57:
		return "dense freezing drizzle"
	case 61:
		return "slight rain"
	case 63:
		return "moderate rain"
	case 65:
		return "heavy rain"
	case 66:
		return "light freezing rain"
	case 67:
		return "heavy freezing rain"
	case 71:
		return "slight snow fall"
	case 73:
		return "moderate snow fall"
	case 75:
		return "heavy snow fall"
	case 77:
		return "snow grains"
	case 80:
		return "slight rain showers"
	case 81:
		return "moderate rain showers"
	case 82:
		return "violent rain showers"
	case 85:
		return "slight snow showers"
	case 86:
		return "heavy snow showers"
	case 95:
		return "thunderstorm"
	case 96:
		return "thunderstorm with slight hail"
	case 99:
		return "thunderstorm with heavy hail"
	default:
		return "unknown"
	}
}
