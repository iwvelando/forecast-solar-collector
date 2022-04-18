package influxdb

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/huandu/go-clone"
	influx "github.com/influxdata/influxdb-client-go/v2"
	influxAPI "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/iwvelando/forecast-solar-collector/config"
	"github.com/iwvelando/forecast-solar-collector/model"
)

// Connect establishes an InfluxDB client
func Connect(config *config.Configuration) (influx.Client, influxAPI.WriteAPIBlocking, error) {
	var auth string
	if config.InfluxDB.Token != "" {
		auth = config.InfluxDB.Token
	} else if config.InfluxDB.Username != "" && config.InfluxDB.Password != "" {
		auth = fmt.Sprintf("%s:%s", config.InfluxDB.Username, config.InfluxDB.Password)
	} else {
		auth = ""
	}

	var writeDest string
	if config.InfluxDB.Bucket != "" {
		writeDest = config.InfluxDB.Bucket
	} else if config.InfluxDB.Database != "" && config.InfluxDB.RetentionPolicy != "" {
		writeDest = fmt.Sprintf("%s/%s", config.InfluxDB.Database, config.InfluxDB.RetentionPolicy)
	} else {
		return nil, nil, fmt.Errorf("must configure at least one of bucket or database/retention policy")
	}

	options := influx.DefaultOptions().
		SetTLSConfig(&tls.Config{
			InsecureSkipVerify: config.InfluxDB.SkipVerifySsl,
		})
	client := influx.NewClientWithOptions(config.InfluxDB.Address, auth, options)

	writeAPI := client.WriteAPIBlocking(config.InfluxDB.Organization, writeDest)

	return client, writeAPI, nil
}

// WriteAll performs final processing and formatting of raw data before submitting to InfluxDB
func WriteAll(config *config.Configuration, writeAPI influxAPI.WriteAPIBlocking, forecasts []model.SolarForecast) error {

	if !config.ForecastSolar.NoComputeTotals {
		// Compute aggregate data
		var forecastAggregate model.SolarForecast
		for i, forecast := range forecasts {
			if i == 0 {
				forecastAggregate = clone.Clone(forecast).(model.SolarForecast)
				continue
			}

			for ts0 := range forecastAggregate.Result.Watts {
				for ts1, value := range forecast.Result.Watts {
					if ts0.Equal(ts1) {
						forecastAggregate.Result.Watts[ts0] += value
					}
				}
			}

			for ts0 := range forecastAggregate.Result.WattHours {
				for ts1, value := range forecast.Result.WattHours {
					if ts0.Equal(ts1) {
						forecastAggregate.Result.WattHours[ts0] += value
					}
				}
			}

			for ts0 := range forecastAggregate.Result.WattHoursDay {
				for ts1, value := range forecast.Result.WattHoursDay {
					if ts0.Equal(ts1) {
						forecastAggregate.Result.WattHoursDay[ts0] += value
					}
				}
			}
		}

		// Aggregate forecast data

		// Points for instantaneous production data
		for ts, value := range forecastAggregate.Result.Watts {
			p := influx.NewPoint(
				config.InfluxDB.MeasurementPrefix+"energy_forecast",
				map[string]string{
					"array_name":     model.AggregateProductionArrayName,
					"site":           forecastAggregate.Message.Info.Place,
					"site_latitude":  fmt.Sprintf("%f", forecastAggregate.Message.Info.Latitude),
					"site_longitude": fmt.Sprintf("%f", forecastAggregate.Message.Info.Longitude),
				},
				map[string]interface{}{
					"instantaneous_production_watts": value,
				},
				ts)
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}

		// Points for cumulative hourly production data
		for ts, value := range forecastAggregate.Result.WattHours {
			p := influx.NewPoint(
				config.InfluxDB.MeasurementPrefix+"energy_forecast",
				map[string]string{
					"array_name":     model.AggregateProductionArrayName,
					"site":           forecastAggregate.Message.Info.Place,
					"site_latitude":  fmt.Sprintf("%f", forecastAggregate.Message.Info.Latitude),
					"site_longitude": fmt.Sprintf("%f", forecastAggregate.Message.Info.Longitude),
				},
				map[string]interface{}{
					"cumulative_production_watt_hours_hourly": value,
				},
				ts)
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}

		// Points for cumulative daily production data
		for ts, value := range forecastAggregate.Result.WattHoursDay {
			p := influx.NewPoint(
				config.InfluxDB.MeasurementPrefix+"energy_forecast",
				map[string]string{
					"array_name":     model.AggregateProductionArrayName,
					"site":           forecastAggregate.Message.Info.Place,
					"site_latitude":  fmt.Sprintf("%f", forecastAggregate.Message.Info.Latitude),
					"site_longitude": fmt.Sprintf("%f", forecastAggregate.Message.Info.Longitude),
				},
				map[string]interface{}{
					"cumulative_production_watt_hours_daily": value,
				},
				ts)
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}
	}

	// Per-array forecast data

	for _, forecast := range forecasts {
		// Points for instantaneous production data
		for ts, value := range forecast.Result.Watts {
			p := influx.NewPoint(
				config.InfluxDB.MeasurementPrefix+"energy_forecast",
				map[string]string{
					"array_name":     forecast.Name,
					"site":           forecast.Message.Info.Place,
					"site_latitude":  fmt.Sprintf("%f", forecast.Message.Info.Latitude),
					"site_longitude": fmt.Sprintf("%f", forecast.Message.Info.Longitude),
				},
				map[string]interface{}{
					"instantaneous_production_watts": value,
				},
				ts)
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}

		// Points for cumulative hourly production data
		for ts, value := range forecast.Result.WattHours {
			p := influx.NewPoint(
				config.InfluxDB.MeasurementPrefix+"energy_forecast",
				map[string]string{
					"array_name":     forecast.Name,
					"site":           forecast.Message.Info.Place,
					"site_latitude":  fmt.Sprintf("%f", forecast.Message.Info.Latitude),
					"site_longitude": fmt.Sprintf("%f", forecast.Message.Info.Longitude),
				},
				map[string]interface{}{
					"cumulative_production_watt_hours_hourly": value,
				},
				ts)
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}

		// Points for cumulative daily production data
		for ts, value := range forecast.Result.WattHoursDay {
			p := influx.NewPoint(
				config.InfluxDB.MeasurementPrefix+"energy_forecast",
				map[string]string{
					"array_name":     forecast.Name,
					"site":           forecast.Message.Info.Place,
					"site_latitude":  fmt.Sprintf("%f", forecast.Message.Info.Latitude),
					"site_longitude": fmt.Sprintf("%f", forecast.Message.Info.Longitude),
				},
				map[string]interface{}{
					"cumulative_production_watt_hours_daily": value,
				},
				ts)
			err := writeAPI.WritePoint(context.Background(), p)
			if err != nil {
				return err
			}
		}
	}

	return nil

}
