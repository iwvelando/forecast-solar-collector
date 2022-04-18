// Package config defines the data structures related to configuration and
// includes functions for loading and parsing the config.
package config

import (
	"fmt"
	"github.com/iwvelando/forecast-solar-collector/model"
	"github.com/spf13/viper"
)

// Configuration holds all configuration.
type Configuration struct {
	SolarPanels   SolarPanels
	ForecastSolar ForecastSolar
	InfluxDB      InfluxDB
}

// SolarPanels holds the solar panel hardware information
type SolarPanels struct {
	Location SolarPanelLocation
	Arrays   []SolarPanelArray
}

// SolarPanelLocation indicates the geographic coordinates where the arrays are installed
type SolarPanelLocation struct {
	Latitude  float64
	Longitude float64
}

// SolarPanelArray describes the physical properties of the solar panels
type SolarPanelArray struct {
	Name        string
	Declination float64
	Azimuth     float64
	Power       float64
}

// ForecastSolar holds the information for querying from forecast.solar
type ForecastSolar struct {
	URL             string // Base URL for forecast.solar, defaults to https://api.forecast.solar
	APIKey          string
	SkipVerifySsl   bool
	NoComputeTotals bool
}

// InfluxDB holds the connection parameters for InfluxDB
type InfluxDB struct {
	Address           string
	Username          string
	Password          string
	MeasurementPrefix string
	Database          string
	RetentionPolicy   string
	Token             string
	Organization      string
	Bucket            string
	SkipVerifySsl     bool
	FlushInterval     uint
}

func (r *Configuration) CheckDefaults() {
	if r.ForecastSolar.URL == "" {
		r.ForecastSolar.URL = "https://api.forecast.solar"
	}
}

func (r *Configuration) Validate() error {
	for _, array := range r.SolarPanels.Arrays {
		if array.Name == model.AggregateProductionArrayName {
			return fmt.Errorf("array name '%s' conflicts with reserved name for logical array that sums all arrays", model.AggregateProductionArrayName)
		}
	}
	return nil
}

// LoadConfiguration takes a file path as input and loads the YAML-formatted
// configuration there.
func LoadConfiguration(configPath string) (*Configuration, error) {
	viper.SetConfigFile(configPath)
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file, %s", err)
	}

	var configuration Configuration
	err := viper.Unmarshal(&configuration)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct, %s", err)
	}

	return &configuration, nil
}
