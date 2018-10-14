package util

import (
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type HTTP interface {
	Get(url string) ([]byte, error)
}

type HTTPClient struct {
	c *http.Client
}

func NewHTTPClient(timeout time.Duration) HTTP {
	client := &http.Client{Timeout: timeout * time.Second}
	return &HTTPClient{client}
}

func (client *HTTPClient) Get(url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("ERROR:   Unable to create HTTP request. Message: `%s`.", err.Error())
		return nil, err
	}
	req.Header.Add("User-Agent", client.randomizedUserAgent())
	resp, err := client.c.Do(req)
	if err != nil {
		log.Printf("ERROR:   HTTP request to URL `%s` failed. Message: `%s`.", url, err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	log.Printf("INFO:    HTTP call executed: `%s`.", url)

	return ioutil.ReadAll(resp.Body)
}

func (client *HTTPClient) randomizedUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36"
}
