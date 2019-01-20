package fetcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RadioCheckerApp/api/model"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const requestTimeout = 5
const kronehitId = "kronehit"
const kronehitAPI = "https://www.kronehit.at/alles-ueber-kronehit/hitsuche/?format=json&day=%s&channel=%d&hours=%02d&minutes=%02d"
const kronehitChannel = 1
const kronehitRequestLimit = 10
const kronehitTimeCorrection = 7 * time.Minute

type KronehitItem struct {
	PlayTime       string
	ArtistName     string
	TrackName      string
	ArtistPageLink string `json:"-"`
}

func (item *KronehitItem) Hour() int {
	playTimeStrings := strings.Split(item.PlayTime, ":")
	hour, _ := strconv.Atoi(playTimeStrings[0])
	return hour
}

func (item *KronehitItem) toTrackRecord(playDate *time.Time) (*model.TrackRecord, error) {
	dateTimeStr := fmt.Sprintf("%s %s", playDate.Format("2006-01-02"), item.PlayTime)
	dateTime, err := time.ParseInLocation("2006-01-02 15:04", dateTimeStr, getLocation())
	if err != nil {
		log.Printf("ERROR:   Unable to parse `%s` to time. Message: `%s`.",
			dateTimeStr, err.Error())
		return nil, err
	}

	return &model.TrackRecord{
		kronehitId,
		dateTime.Unix(),
		"track",
		model.Track{item.ArtistName, item.TrackName},
	}, nil
}

type KronehitItems struct {
	Items []KronehitItem
}

func (items *KronehitItems) toTrackRecords(fetchTime *time.Time,
	skip func(record *model.TrackRecord) bool) []*model.TrackRecord {
	spanningOverMidnight := items.spanOverMidnight()
	fetchedOverMidnight := items.fetchedOverMidnight(fetchTime)
	var trackRecords []*model.TrackRecord
	for _, item := range items.Items {
		playDate := *fetchTime
		if spanningOverMidnight || fetchedOverMidnight {
			playDate = sanitizedPlayDate(&playDate, item.Hour())
		}
		trackRecord, err := item.toTrackRecord(&playDate)
		if err != nil {
			log.Printf("ERROR:   Unable to extract TrackRecord from item: `%q`. Message: `%s`.",
				item, err.Error())
			continue
		}
		if skip(trackRecord) {
			log.Printf("INFO:    Skipping item `%s - %s` (Airtime: %s). "+
				"Newer than or equal to last fetched Track.",
				item.ArtistName, item.TrackName, item.PlayTime)
			continue
		}
		trackRecords = append(trackRecords, trackRecord)
	}
	return trackRecords
}

func (items *KronehitItems) spanOverMidnight() bool {
	if len(items.Items) <= 1 {
		return false
	}
	return items.Items[0].Hour() > items.Items[len(items.Items)-1].Hour()
}

func (items *KronehitItems) fetchedOverMidnight(fetchTime *time.Time) bool {
	return items.Items[0].Hour() > fetchTime.Hour()
}

func sanitizedPlayDate(fetchTime *time.Time, itemHour int) time.Time {
	if fetchTime.Hour() < 12 && itemHour > 12 {
		return fetchTime.AddDate(0, 0, -1)
	} else if fetchTime.Hour() >= 12 && itemHour <= 12 {
		return fetchTime.AddDate(0, 0, 1)
	}
	return *fetchTime
}

type KronehitAPI interface {
	GetItems(time.Time) (KronehitItems, error)
}

type KronehitAPIImplementation struct {
	client *http.Client
}

func NewKronehitAPIImplementation(timeout time.Duration) KronehitAPIImplementation {
	client := &http.Client{Timeout: timeout * time.Second}
	return KronehitAPIImplementation{client}
}

func (api KronehitAPIImplementation) GetItems(date time.Time) (KronehitItems, error) {
	url := fmt.Sprintf(
		kronehitAPI,
		date.Format("2006-01-02"),
		kronehitChannel,
		date.Hour(),
		date.Minute(),
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("ERROR:   Unable to create HTTP request. Message: `%s`.", err.Error())
		return KronehitItems{}, err
	}
	req.Header.Add("User-Agent", randomizedUserAgent())
	resp, err := api.client.Do(req)
	if err != nil {
		log.Printf("ERROR:   HTTP request to URL `%s` failed. Message: `%s`.", url, err.Error())
		return KronehitItems{}, err
	}
	defer resp.Body.Close()

	log.Printf("INFO:    HTTP call executed: `%s`.", url)

	var items KronehitItems
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		log.Printf("ERROR:   Unmarshalling JSON body failed. Message: `%s`.", err.Error())
		return KronehitItems{}, err
	}

	return items, nil
}

type KronehitFetcher struct {
	kronehitAPI   KronehitAPI
	nextFetchTime time.Time
	fetchCounter  int
}

func NewKronehitFetcher() KronehitFetcher {
	kronehitAPI := NewKronehitAPIImplementation(requestTimeout)
	// ALWAYS crawl in the past to avoid inconsistent data
	nextFetchTime := time.Now().Add(-kronehitTimeCorrection).In(getLocation())
	log.Printf("INFO:    Set nextFetchTime to %s.", nextFetchTime.Format("2006-01-02 15:04:05"))
	return KronehitFetcher{kronehitAPI, nextFetchTime, 0}
}

func (fetcher *KronehitFetcher) Next() ([]*model.TrackRecord, error) {
	if fetcher.fetchCounter >= kronehitRequestLimit {
		log.Printf("ERROR:   Request limit exceeded.")
		return nil, errors.New("request limit exceeded")
	}

	items, err := fetcher.kronehitAPI.GetItems(fetcher.nextFetchTime)
	if err != nil {
		log.Printf("ERROR:   Unable to fetch items from kronehit. Message: `%s`.", err.Error())
		return nil, err
	}

	log.Printf("INFO:    Fetched %d items from Kronehit.", len(items.Items))

	trackRecords := items.toTrackRecords(&fetcher.nextFetchTime, func(record *model.TrackRecord) bool {
		// always skip records that are younger than the last fetch time
		return record.Timestamp >= fetcher.nextFetchTime.Add(kronehitTimeCorrection).Unix()
	})

	if len(trackRecords) == 0 {
		log.Printf("WARNING: Unable to extract any TrackRecords from %d items. SkipRate = %."+
			"2f%%", len(items.Items), calculateSkipRate(len(trackRecords), len(items.Items)))
		return nil, errors.New("unable to extract any trackRecords")
	}

	sort.Slice(trackRecords, func(i, j int) bool {
		return trackRecords[i].Timestamp > trackRecords[j].Timestamp
	})

	lastFetchedTrackTimestamp := trackRecords[len(trackRecords)-1].Timestamp
	fetcher.nextFetchTime = time.Unix(lastFetchedTrackTimestamp, 0).
		In(getLocation()).Add(-kronehitTimeCorrection)
	fetcher.fetchCounter++

	log.Printf("INFO:    Returned %d TrackRecords, extracted from %d items. SkipRate = %.2f%%",
		len(trackRecords), len(items.Items), calculateSkipRate(len(trackRecords), len(items.Items)))
	return trackRecords, nil
}

func (fetcher *KronehitFetcher) isFirstFetch() bool {
	return fetcher.fetchCounter == 0
}

// randomizedInitialFetchTime generates a random user agent that is sent by common browsers.
// This behaviour should help to disguise the fetcher as real user/browser.
func randomizedUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/60.0.3112.113 Safari/537.36"
}

func getLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		log.Fatal("getLocation: " + err.Error())
	}
	return loc
}

func calculateSkipRate(extractedTrackRecords, fetchedItems int) float32 {
	if fetchedItems <= 0 {
		return 0
	}
	skippedItems := fetchedItems - extractedTrackRecords
	return float32(skippedItems) / float32(fetchedItems) * 100
}
