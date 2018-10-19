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

	twitterConsumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	twitterConsumerKeySecret := os.Getenv("TWITTER_CONSUMER_KEY_SECRET")
	twitterOauthAccessToken := os.Getenv("TWITTER_OAUTH_ACCESS_TOKEN")
	twitterOauthAccessTokenSecret := os.Getenv("TWITTER_OAUTH_ACCESS_TOKEN_SECRET")

	stationId := os.Getenv("STATION_ID_HITRADIO_OE3")
	rcAPIHost := os.Getenv("RC_API_HOST")
	rcAPIKey := os.Getenv("RC_API_KEY")
	rcAPIAuthorization := os.Getenv("RC_API_AUTHORIZATION")

	oe3Fetcher, err := fetcher.NewHitradioOE3Fetcher(
		twitterConsumerKey,
		twitterConsumerKeySecret,
		twitterOauthAccessToken,
		twitterOauthAccessTokenSecret,
	)
	if err != nil {
		log.Printf("ERROR:  Unable to create HitradioOE3Fetcher. Message: `%s`.", err.Error())
		return err
	}

	homebase := crawler.HomeBaseConnector{
		rcAPIHost,
		rcAPIKey,
		rcAPIAuthorization,
	}

	oe3Crawler, err := crawler.NewCrawler(stationId, oe3Fetcher, homebase)
	if err != nil {
		log.Printf("ERROR:   Unable to create crawler for station `%s`. Message: `%s`.",
			stationId, err.Error())
		return err
	}

	return oe3Crawler.Crawl()
}

func main() {
	lambda.Start(Handler)
}
