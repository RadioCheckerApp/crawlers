package main

import (
	"github.com/RadioCheckerApp/crawlers/crawler"
	"github.com/RadioCheckerApp/crawlers/fetcher"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"log"
	"os"
)

func Handler(event events.CloudWatchEvent) error {
	log.Println("INFO:    Crawler triggered.")
	defer log.Println("INFO:    Crawler finshed.")

	ooeFetcher, err := fetcher.NewORFFetcher("radio-oberösterreich")
	if err != nil {
		log.Printf("ERROR:   %s", err.Error())
		return err
	}

	stationId := os.Getenv("STATION_ID_RADIO_OBEROESTERREICH")
	rcAPIHost := os.Getenv("RC_API_HOST")
	rcAPIKey := os.Getenv("RC_API_KEY")
	rcAPIAuthorization := os.Getenv("RC_API_AUTHORIZATION")

	homebase := crawler.HomeBaseConnector{
		rcAPIHost,
		rcAPIKey,
		rcAPIAuthorization,
	}

	ooeCrawler, err := crawler.NewCrawler(stationId, ooeFetcher, homebase)
	if err != nil {
		log.Printf("ERROR:   Unable to create crawler for station `%s`. Message: `%s`.",
			stationId, err.Error())
		return err
	}

	ooeCrawler.Crawl()

	return nil
}

func main() {
	lambda.Start(Handler)
}
