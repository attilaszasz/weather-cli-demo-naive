# weather-cli-demo-naive

A standalone Go command-line executable that accepts geocoordinates and returns the current weather conditions in JSON format.

## Usage

```bash
go run . -lat 47.4979 -lon 19.0402
```

Example output:

```json
{
  "latitude": 47.5,
  "longitude": 19.0,
  "timezone": "Europe/Budapest",
  "current": {
    "time": "2026-04-06T13:00",
    "temperature_c": 18.4,
    "relative_humidity_percent": 42,
    "apparent_temperature_c": 18.0,
    "is_day": true,
    "precipitation_mm": 0,
    "rain_mm": 0,
    "showers_mm": 0,
    "snowfall_cm": 0,
    "cloud_cover_percent": 12,
    "pressure_msl_hpa": 1017.2,
    "surface_pressure_hpa": 995.1,
    "wind_speed_kmh": 11.3,
    "wind_direction_degrees": 224,
    "wind_gusts_kmh": 21.8,
    "weather_code": 1,
    "weather_description": "mainly clear"
  }
}
```

## Build

```bash
go build -o weather-cli .
```

## Release

Pushing a tag like `v1.0.0` triggers GitHub Actions to build and upload binaries for:

- Linux amd64
- macOS amd64
- macOS arm64
- Windows amd64

The CLI uses the Open-Meteo API and does not require an API key.
