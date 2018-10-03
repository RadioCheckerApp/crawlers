package crawler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RadioCheckerApp/api/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type HomeBase interface {
	getLatestTrackRecord(stationId string) (*model.TrackRecord, error)
	persistTrackRecord(trackRecord *model.TrackRecord) error
}

type HomeBaseConnector struct {
	APIHost          string
	APIKey           string
	APIAuthorization string
}

func (api HomeBaseConnector) getLatestTrackRecord(stationId string) (*model.TrackRecord, error) {
	url := fmt.Sprintf("https://%s/stations/%s/tracks?filter=latest",
		api.APIHost, stationId)

	responseData, err := api.callEndpoint(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("WARNING: Unable to get latest TrackRecord from endpoint `%s %s`. "+
			"Message: `%s`.", http.MethodGet, url, err.Error())
		return nil, err
	}

	responseDataMap := responseData.(map[string]interface{})
	trackRecordJSON, _ := json.Marshal(responseDataMap)

	var latestTrackRecord model.TrackRecord
	err = json.Unmarshal(trackRecordJSON, &latestTrackRecord)
	if err != nil {
		log.Printf("ERROR:   Unable to unmarshal JSON object into TrackRecord. Message: `%s`.",
			err.Error())
		return nil, err
	}

	return &latestTrackRecord, nil
}

func (api HomeBaseConnector) persistTrackRecord(trackRecord *model.TrackRecord) error {
	url := fmt.Sprintf("https://%s/stations/%s/tracks/%d",
		api.APIHost, trackRecord.StationId, trackRecord.Timestamp)

	payload, err := json.Marshal(trackRecord.Track)
	if err != nil {
		log.Printf("ERROR:   Unable to marshal Track to JSON. Message: `%s`.", err.Error())
		return err
	}

	responseData, err := api.callEndpoint(http.MethodPut, url, bytes.NewBuffer(payload))
	if err != nil {
		log.Printf("ERROR:   Unable call endpoint `%s %s`. Message: `%s`.",
			http.MethodPut, url, err.Error())
		return err
	}

	log.Printf("INFO:    TrackRecord persisted. Message: `%s`.", responseData)
	return nil
}

func (api HomeBaseConnector) callEndpoint(method, url string, payload io.Reader) (interface{},
	error) {
	responseBody, err := api.sendHTTPRequest(method, url, payload)
	if err != nil {
		log.Printf("ERROR:   Unable to read body of response to `%s %s`. Message: `%s`.",
			method, url, err.Error())
		return nil, err
	}

	responseData, err := getDataFromResponseBody(responseBody)
	if err != nil {
		log.Printf("WARNING: Unable to read data from response. Message: `%s`.", err.Error())
		return nil, err
	}

	return responseData, nil
}

func (api HomeBaseConnector) sendHTTPRequest(method, url string, payload io.Reader) ([]byte,
	error) {
	client := http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		log.Printf("ERROR:   Unable to create new %s request. Message: `%s`.", method, err)
		return nil, err
	}

	if method == http.MethodPut {
		req.Header.Set("Authorization", "Bearer "+api.APIAuthorization)
	} else {
		req.Header.Set("X-API-KEY", api.APIKey)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ERROR:   Unable to call endpoint `%s %s`. Message: `%s`.",
			url, method, err.Error())
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getDataFromResponseBody(body []byte) (interface{}, error) {
	var responseDataFields map[string]interface{}
	err := json.Unmarshal(body, &responseDataFields)
	if err != nil {
		log.Printf("ERROR:   Unable to unmarshal JSON response into a key map. Message: `%s`.",
			err.Error())
		return nil, err
	}

	if _, ok := responseDataFields["success"]; !ok {
		log.Printf("ERROR:   Illegal JSON response format: `%s`.", body)
		return nil, errors.New("illegal JSON response format")
	}

	success := responseDataFields["success"].(bool)
	if !success {
		log.Printf("WARNING: Request did not return any data. Message: `%s`.",
			responseDataFields["message"])
		return nil, errors.New("request did not return any data")
	}

	return responseDataFields["data"], nil
}
