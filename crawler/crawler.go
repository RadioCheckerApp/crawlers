package crawler

import (
	"errors"
	"github.com/RadioCheckerApp/api/model"
	"github.com/RadioCheckerApp/crawlers/fetcher"
	"log"
	"time"
)

const (
	timeDeltaLatestTrackRecord = 10
	timeDeltaEqualTrackRecords = 300
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

func (crawler *Crawler) Crawl() error {
	if time.Now().Unix() <= crawler.latestTrackRecordTimestamp {
		log.Println("INFO:    Crawler quit since latest TrackRecord is newer than current time.")
		return nil
	}

	overallPersistedCounter := 0
	upToDate := false
	var fetchErr error = nil
	for !upToDate {
		trackRecords, err := crawler.fetcher.Next()
		if err != nil {
			fetchErr = err
			break
		}
		trackRecords = crawler.filterDuplicates(trackRecords)
		persistedCounter, status := crawler.batchPersistTrackRecords(trackRecords)
		upToDate = status
		overallPersistedCounter += persistedCounter
	}

	log.Printf("INFO:    %d TrackRecords persisted.", overallPersistedCounter)

	if fetchErr != nil {
		log.Printf("WARNING: Crawler finished with error. Message: `%s`.", fetchErr.Error())
		return fetchErr
	}

	if !upToDate {
		log.Println("WARNING: Crawler failed to update records.")
		return errors.New("crawler failed to update records")
	}

	log.Println("INFO:    Crawler successfully updated records.")
	return nil
}

func (crawler *Crawler) batchPersistTrackRecords(trackRecords []*model.TrackRecord) (int, bool) {
	insertedTracksCounter := 0
	for _, trackRecord := range trackRecords {
		if trackRecord.Timestamp <= crawler.latestTrackRecordTimestamp {
			return insertedTracksCounter, true
		}
		if trackRecord.Timestamp-crawler.latestTrackRecordTimestamp <= timeDeltaLatestTrackRecord {
			log.Printf("WARNING: TrackRecord of batch has critical proximity to "+
				"lastestTreckRecordTimestamp (%d), but is not equal. Finishing with success. "+
				"TrackRecord: `%q`", crawler.latestTrackRecordTimestamp, trackRecord)
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

func (crawler *Crawler) filterDuplicates(records []*model.TrackRecord) []*model.TrackRecord {
	if len(records) < 2 {
		return records
	}

	filteredRecords := []*model.TrackRecord{records[0]}
	for i := 1; i < len(records) && records[i].Timestamp >= crawler.latestTrackRecordTimestamp; i++ {
		prevRecord := filteredRecords[len(filteredRecords)-1]
		currRecord := records[i]
		if areTrackRecordsEqual(prevRecord, currRecord) {
			log.Printf("WARNING: Crawler skipping duplicated track record: `%q`.", currRecord)
			continue
		}
		filteredRecords = append(filteredRecords, currRecord)
	}
	return filteredRecords
}

func areTrackRecordsEqual(a, b *model.TrackRecord) bool {
	// TrackRecords are equal if they have the same artist & title
	// and have a matching timestamp (±5 minutes)
	return a.Title == b.Title &&
		a.Artist == b.Artist &&
		abs(a.Timestamp-b.Timestamp) < timeDeltaEqualTrackRecords
}

func abs(n int64) int64 {
	// a fast absolute value function for int64
	// without the casting overhead required when by the stdlib
	// http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
	y := n >> 63
	return (n ^ y) - y
}
