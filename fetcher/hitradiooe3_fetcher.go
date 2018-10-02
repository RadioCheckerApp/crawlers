package fetcher

import (
	"errors"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"github.com/RadioCheckerApp/api/model"
	"log"
	"net/url"
	"strings"
)

const twitterUserID = "7901732"
const twitterTweetCount = 200
const radioStationId = "hitradio-oe3"
const trackType = "track"

type TwitterAPI interface {
	GetUserTimeline(v url.Values) ([]anaconda.Tweet, error)
}

type HitradioOE3Fetcher struct {
	twitterAPI       TwitterAPI
	twitterAPIParams url.Values
}

func NewHitradioOE3Fetcher(consumerKey, consumerKeySecret, oauthAccessToken,
	oauthAccessTokenSecret string) (HitradioOE3Fetcher, error) {
	if consumerKey == "" || consumerKeySecret == "" || oauthAccessToken == "" ||
		oauthAccessTokenSecret == "" {
		return HitradioOE3Fetcher{}, errors.New("keys and tokens must not be empty")
	}

	twitterAPI := anaconda.NewTwitterApiWithCredentials(oauthAccessToken, oauthAccessTokenSecret,
		consumerKey, consumerKeySecret)
	if twitterAPI == nil {
		return HitradioOE3Fetcher{}, errors.New("could not create Twitter API handler")
	}

	return HitradioOE3Fetcher{twitterAPI, buildInitialParams()}, nil
}

func buildInitialParams() url.Values {
	values := url.Values{}
	values.Set("user_id", twitterUserID)
	values.Set("count", fmt.Sprintf("%d", twitterTweetCount))
	values.Set("trim_user", "true")
	values.Set("exclude_replies", "true")
	return values
}

func (fetcher HitradioOE3Fetcher) Next() ([]*model.TrackRecord, error) {
	tweets, err := fetcher.twitterAPI.GetUserTimeline(fetcher.twitterAPIParams)
	if err != nil {
		return nil, err
	}

	log.Printf("INFO:    Fetched %d tweets from Twitter account `%s`.", len(tweets), twitterUserID)

	var trackRecords []*model.TrackRecord

	for _, tweet := range tweets {
		if tweet.IdStr == fetcher.twitterAPIParams.Get("max_id") {
			// Requests to the Twitter API that contain the `max_id` param are inclusive,
			// meaning that the tweet with the respective ID is (again) included in the response.
			// To avoid duplicates, the first (matching) tweet of the response has to be skipped.
			log.Printf("INFO:    Skipped tweet with ID `%s` (created_at: %s).",
				tweet.IdStr, tweet.CreatedAt)
			continue
		}
		trackRecord, err := extractTrackRecordFromTweet(tweet)
		if err != nil {
			log.Printf("ERROR:   Unable to extract TrackRecord from tweet: `%s`. Message: `%s`.",
				tweet.Text, err.Error())
			continue
		}
		trackRecords = append(trackRecords, trackRecord)
		fetcher.twitterAPIParams.Set("max_id", tweet.IdStr)
	}

	log.Printf("INFO:    Returned %d TrackRecords, extracted from %d tweets.",
		len(trackRecords), len(tweets))
	return trackRecords, nil
}

func extractTrackRecordFromTweet(tweet anaconda.Tweet) (*model.TrackRecord, error) {
	// Tweet format: `<airtime>: "<title>" von <artist>`
	// For convenience (and also error resistance),
	// the time when the tweet was created is used as airtime in the track record. In most cases,
	// this is a pretty accurate measurement and only diverges about a minute or less from the
	// real time included in the Tweet text. However, this approach could lead to errors if the
	// tweets are not created directly after the track has been aired.
	// TODO: find a viable fallback for the case that the tweet's creation date and the actual
	// time in the text diverge too much (edge cases, i.e. midnight leaps, should be considered).
	airtime, err := tweet.CreatedAtTime()
	if err != nil {
		return nil, err
	}

	title, err := getTitleFromTweetText(tweet.FullText)
	if err != nil {
		return nil, err
	}

	artist, err := getArtistFromTweetText(tweet.FullText)
	if err != nil {
		return nil, err
	}

	return &model.TrackRecord{
		radioStationId,
		airtime.Unix(),
		trackType,
		model.Track{Title: title, Artist: artist},
	}, nil
}

func getTitleFromTweetText(text string) (string, error) {
	splitted := strings.Split(text, "\"")
	if len(splitted) < 3 {
		return "", errors.New("unable to extract title from tweet")
	}
	return splitted[1], nil
}

func getArtistFromTweetText(text string) (string, error) {
	splitted := strings.Split(text, " von ")
	if len(splitted) < 2 {
		return "", errors.New("unable to extract artist from tweet")
	}
	return splitted[1], nil
}
