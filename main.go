package main

import (
	"flag"
	"fmt"
	"github.com/iwvelando/forecast-solar-collector/config"
	"github.com/iwvelando/forecast-solar-collector/connect"
	"github.com/iwvelando/forecast-solar-collector/influxdb"
	log "github.com/sirupsen/logrus"
	"os"
)

// BuildVersion is the software build version
var BuildVersion = "UNKNOWN"

// CliInputs holds the data passed in via CLI parameters
type CliInputs struct {
	BuildVersion string
	Config       string
	ShowVersion  bool
}

func main() {

	cliInputs := CliInputs{
		BuildVersion: BuildVersion,
	}
	flags := flag.NewFlagSet("forecast-solar-collector", 0)
	flags.StringVar(&cliInputs.Config, "config", "config.yaml", "Set the location for the YAML config file")
	flags.BoolVar(&cliInputs.ShowVersion, "version", false, "Print the version of forecast-solar-collector")
	flags.Parse(os.Args[1:])

	if cliInputs.ShowVersion {
		fmt.Println(cliInputs.BuildVersion)
		os.Exit(0)
	}

	configuration, err := config.LoadConfiguration(cliInputs.Config)
	if err != nil {
		log.WithFields(log.Fields{
			"op":    "config.LoadConfiguration",
			"error": err,
		}).Fatal("failed to parse configuration")
	}
	configuration.CheckDefaults()
	err = configuration.Validate()
	if err != nil {
		log.WithFields(log.Fields{
			"op":    "config.Validate",
			"error": err,
		}).Fatal("encountered configuration validation error")
	}

	client, baseURL := connect.Client(configuration)
	defer client.CloseIdleConnections()

	influxClient, writeAPI, err := influxdb.Connect(configuration)
	if err != nil {
		log.WithFields(log.Fields{
			"op":    "influxdb.Connect",
			"error": err,
		}).Fatal("failed to authenticate to InfluxDB")
	}
	defer influxClient.Close()

	forecasts, err := connect.GetAll(configuration, client, baseURL)
	if err != nil {
		log.WithFields(log.Fields{
			"op":    "connect.GetAll",
			"error": err,
		}).Error("failed to query all metrics, exiting")
		os.Exit(1)
	} else {
		err = influxdb.WriteAll(configuration, writeAPI, forecasts)
		if err != nil {
			log.WithFields(log.Fields{
				"op":    "influxdb.WriteAll",
				"error": err,
			}).Error("failed to write data to InfluxDB, exiting")
			os.Exit(1)
		}
	}

	log.WithFields(log.Fields{
		"op":        "main",
		"remaining": forecasts[len(forecasts)-1].Message.Ratelimit.Remaining,
	}).Infof("there are %d remaining API queries out of %d for the next %d seconds",
		forecasts[len(forecasts)-1].Message.Ratelimit.Remaining,
		forecasts[len(forecasts)-1].Message.Ratelimit.Limit,
		forecasts[len(forecasts)-1].Message.Ratelimit.Period)

}
