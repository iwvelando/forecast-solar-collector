package model

import (
	"time"
)

// DateTimeForecast provides a format for timestamps provided by forecast.solar
const DateTimeForecast = "2006-01-02 15:04:05"

// AggregateProductionArrayName is the reserved name given to the logical array that sums up all others
const AggregateProductionArrayName = "all"

// SolarForecast contains a response for /estimate/
type SolarForecast struct {
	Name    string
	Result  SolarForecastResult  `json:"result"`
	Message SolarForecastMessage `json:"message"`
}

// SolarForecastResult contains the production data
type SolarForecastResult struct {
	WattsRaw        map[string]int `json:"watts"`
	Watts           map[time.Time]int
	WattHoursRaw    map[string]int `json:"watt_hours"`
	WattHours       map[time.Time]int
	WattHoursDayRaw map[string]int `json:"watt_hours_day"`
	WattHoursDay    map[time.Time]int
}

// SolarForecastMessage contains metadata about the production data
type SolarForecastMessage struct {
	Code      int                           `json:"code"`
	Type      string                        `json:"type"`
	Text      string                        `json:"text"`
	Info      SolarForecastMessageInfo      `json:"info"`
	Ratelimit SolarForecastMessageRatelimit `json:"ratelimit"`
}

// SolarForecastMessageInfo contains location-related metadata about the production data
type SolarForecastMessageInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Place     string  `json:"place"`
	Timezone  string  `json:"timezone"`
}

// SolarForecastMessageRatelimit contains API rate limiting information for forecast.solar
type SolarForecastMessageRatelimit struct {
	Period    int `json:"period"`
	Limit     int `json:"limit"`
	Remaining int `json:"remaining"`
}

// ParseTime converts the string times provided by forecast.solar into time.Time
func (r *SolarForecast) ParseTime() error {

	r.Result.Watts = make(map[time.Time]int)
	r.Result.WattHours = make(map[time.Time]int)
	r.Result.WattHoursDay = make(map[time.Time]int)

	location, err := time.LoadLocation(r.Message.Info.Timezone)
	if err != nil {
		return err
	}

	for ts, value := range r.Result.WattsRaw {
		if ts != "" {
			t, err := time.ParseInLocation(DateTimeForecast, ts, location)
			if err == nil {
				r.Result.Watts[t] = value
			}
		}
	}

	for ts, value := range r.Result.WattHoursRaw {
		if ts != "" {
			t, err := time.ParseInLocation(DateTimeForecast, ts, location)
			if err == nil {
				r.Result.WattHours[t] = value
			}
		}
	}

	for ts, value := range r.Result.WattHoursDayRaw {
		if ts != "" {
			t, err := time.ParseInLocation(DateTimeForecast, ts+" 23:59:59", location)
			if err == nil {
				r.Result.WattHoursDay[t] = value
			}
		}
	}

	return nil
}
