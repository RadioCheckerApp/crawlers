package crawler

import (
	"errors"
	"github.com/RadioCheckerApp/api/model"
	"github.com/RadioCheckerApp/crawlers/fetcher"
	"log"
	"time"
)

type Crawler struct {
	stationId                  string
	fetcher                    fetcher.Fetcher
	homeBase                   HomeBase
	latestTrackRecordTimestamp int64
}

func NewCrawler(stationId string, fetcher fetcher.Fetcher, homeBase HomeBase) (Crawler, error) {
	if stationId == "" || fetcher == nil || homeBase == nil {
		return Crawler{}, errors.New("invalid parameter(s) provided")
	}

	latestTrackRecordTimestamp := currentDayBeginTimestamp()
	mostRecentTrackRecord, err := homeBase.getLatestTrackRecord(stationId)
	if err != nil && err.Error() != "request did not return any data" {
		return Crawler{}, errors.New("unable to fetch latest TrackRecord: " + err.Error())
	} else if err == nil {
		latestTrackRecordTimestamp = mostRecentTrackRecord.Timestamp
	}

	return Crawler{stationId, fetcher, homeBase, latestTrackRecordTimestamp}, nil
}

func currentDayBeginTimestamp() int64 {
	loc, err := time.LoadLocation("Europe/Vienna")
	if err != nil {
		log.Fatal("FATAL:   Unable to load timezone `Europe/Vienna`.")
	}
	timeInVienna := time.Now().In(loc)
	todayMidnight := time.Date(
		timeInVienna.Year(),
		timeInVienna.Month(),
		timeInVienna.Day(),
		23, 59, 59, 0,
		timeInVienna.Location())
	return todayMidnight.AddDate(0, 0, -1).Unix()
}

func (crawler Crawler) Crawl() {
	overallPersistedCounter := 0
	upToDate := false
	var fetchErr error = nil
	for !upToDate {
		trackRecords, err := crawler.fetcher.Next()
		if err != nil {
			fetchErr = err
			break
		}
		persistedCounter, status := crawler.batchPersistTrackRecords(trackRecords)
		upToDate = status
		overallPersistedCounter += persistedCounter
	}

	if fetchErr != nil {
		log.Printf("WARNING: Crawler finished with error. Message: `%s`.", fetchErr.Error())
	}

	if upToDate {
		log.Printf("INFO:    Crawler successfully updated records.")
	}

	log.Printf("INFO:    %d TrackRecords persisted.", overallPersistedCounter)
}

func (crawler Crawler) batchPersistTrackRecords(trackRecords []model.TrackRecord) (int, bool) {
	insertedTracksCounter := 0
	for _, trackRecord := range trackRecords {
		if trackRecord.Timestamp <= crawler.latestTrackRecordTimestamp {
			return insertedTracksCounter, true
		}
		err := crawler.homeBase.persistTrackRecord(trackRecord)
		if err != nil {
			log.Printf("ERROR:   Unable to persist TrackRecord: `%q`. Message: `%s`.",
				trackRecord, err.Error())
			continue
		}
		insertedTracksCounter++
	}
	return insertedTracksCounter, false
}
