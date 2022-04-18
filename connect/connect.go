package connect

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/iwvelando/forecast-solar-collector/config"
	"github.com/iwvelando/forecast-solar-collector/model"
	"io/ioutil"
	"net/http"
)

const expectedHTTPStatus = 200
const expectedMessageType = "success"

// Client provides and HTTP client connected to forecast.solar
func Client(config *config.Configuration) (*http.Client, string) {

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: config.ForecastSolar.SkipVerifySsl}
	client := &http.Client{}

	var baseURL string
	if config.ForecastSolar.APIKey == "" {
		baseURL = config.ForecastSolar.URL
	} else {
		baseURL = config.ForecastSolar.URL + "/" + config.ForecastSolar.APIKey
	}

	return client, baseURL
}

// GetEndpoint retrieves JSON formatted data from a specific endpoint of forecast.solar
func GetEndpoint(config *config.Configuration, client *http.Client, baseURL string, endpoint string, data *model.SolarForecast) error {
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	status := resp.StatusCode
	if status != expectedHTTPStatus {
		err = fmt.Errorf("expected %d HTTP status code but got %d; raw body %s", expectedHTTPStatus, resp.StatusCode, body)
		return err
	}

	err = json.Unmarshal(body, data)

	if err != nil {
		err = fmt.Errorf("%w; raw body %s", err, body)
		return err
	}

	return nil
}

// GetAll queries all required endpoints from forecast.solar for the production data
func GetAll(config *config.Configuration, client *http.Client, baseURL string) ([]model.SolarForecast, error) {

	forecasts := make([]model.SolarForecast, len(config.SolarPanels.Arrays))

	for i, array := range config.SolarPanels.Arrays {
		forecasts[i].Name = array.Name
		endpoint := fmt.Sprintf("/estimate/%f/%f/%f/%f/%f", config.SolarPanels.Location.Latitude,
			config.SolarPanels.Location.Longitude,
			array.Declination,
			array.Azimuth,
			array.Power)
		err := GetEndpoint(config, client, baseURL, endpoint, &forecasts[i])
		if err != nil {
			return forecasts, err
		}
		if forecasts[i].Message.Type != expectedMessageType {
			return forecasts, fmt.Errorf("expected message type %s but got %s", expectedMessageType, forecasts[i].Message.Type)
		}
		err = forecasts[i].ParseTime()
		if err != nil {
			return forecasts, err
		}
	}

	return forecasts, nil
}
