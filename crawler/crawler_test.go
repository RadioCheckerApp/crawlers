package crawler

import (
	"errors"
	"github.com/RadioCheckerApp/api/model"
	"github.com/RadioCheckerApp/crawlers/fetcher"
	"reflect"
	"testing"
	"time"
)

type MockHomeBaseSuccess struct{}

func (api MockHomeBaseSuccess) getLatestTrackRecord(stationId string) (*model.TrackRecord, error) {
	return &model.TrackRecord{
		"station-a",
		1234567890,
		"track",
		model.Track{"rhcp", "californication"},
	}, nil
}

func (api MockHomeBaseSuccess) persistTrackRecord(trackRecord *model.TrackRecord) error {
	if trackRecord.StationId == "fail" {
		return errors.New("just a test")
	}
	return nil
}

type MockHomeBaseFail struct{}

func (api MockHomeBaseFail) getLatestTrackRecord(stationId string) (*model.TrackRecord, error) {
	if stationId == "fail gracefully" {
		return nil, errors.New("request did not return any data")
	}
	return nil, errors.New("")
}

func (api MockHomeBaseFail) persistTrackRecord(trackRecord *model.TrackRecord) error {
	return errors.New("")
}

type newCrawlerTests struct {
	stationId string
	fetcher   fetcher.Fetcher
	homeBase  HomeBase
}

var loc, _ = time.LoadLocation("Europe/Vienna")
var yesterday = time.Now().In(loc).AddDate(0, 0, -1)
var todayMidnight = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 23, 59, 59, 0, loc)

func TestNewCrawler_Success(t *testing.T) {
	var tests = []newCrawlerTests{
		{"station-a", fetcher.HitradioOE3Fetcher{}, MockHomeBaseSuccess{}},
		{"fail gracefully", fetcher.HitradioOE3Fetcher{}, MockHomeBaseFail{}},
	}

	for _, test := range tests {
		crawler, err := NewCrawler(test.stationId, test.fetcher, test.homeBase)
		if err != nil {
			t.Errorf("NewCrawler: got err: `%s`, expected error: false", err.Error())
		}
		if test.stationId == "fail gracefully" {
			if crawler.latestTrackRecordTimestamp != todayMidnight.Unix() {
				t.Errorf("NewCrawler: expected latestTrackRecordTimestamp to be `%d`, got `%d`",
					todayMidnight.Unix(), crawler.latestTrackRecordTimestamp)
			}
			continue
		}
		expectedCrawler := Crawler{test.stationId, test.fetcher, test.homeBase, 1234567890}
		if !reflect.DeepEqual(crawler, expectedCrawler) {
			t.Errorf("NewCrawler: got\n(%q, %v), expected\n(%q, nil)", crawler, err, expectedCrawler)
		}
	}
}

func TestNewCrawler_Fail(t *testing.T) {
	var tests = []newCrawlerTests{
		{"", fetcher.HitradioOE3Fetcher{}, MockHomeBaseSuccess{}},
		{"station-a", nil, MockHomeBaseSuccess{}},
		{"station-a", fetcher.HitradioOE3Fetcher{}, nil},
	}

	for _, test := range tests {
		_, err := NewCrawler(test.stationId, test.fetcher, test.homeBase)
		if err == nil {
			t.Errorf("NewCrawler: got err: `%s`, expected error: false", err.Error())
		}
	}
}

func TestCrawler_Crawl_QuitIfUpToDate(t *testing.T) {
	crawler := Crawler{
		homeBase:                   MockHomeBaseSuccess{},
		latestTrackRecordTimestamp: time.Now().AddDate(0, 0, 1).Unix(),
	}
	crawler.Crawl()
}

var trackRecordBatch0 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

var trackRecordBatch1 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1234567890, "track", model.Track{"Katy Perry", "Last Friday Night"}},
}

var trackRecordBatch2 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"fail", 1535301300, "fail", model.Track{"fail", "fail"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

func TestCrawler_batchPersistTrackRecords(t *testing.T) {
	var tests = []struct {
		trackRecords                []*model.TrackRecord
		expectedInsertedTracksCount int
		expectedUpToDate            bool
	}{
		{trackRecordBatch0, 3, false},
		{trackRecordBatch1, 1, true},
		{trackRecordBatch2, 2, false},
	}
	crawler := Crawler{latestTrackRecordTimestamp: 1234567890, homeBase: MockHomeBaseSuccess{}}

	for _, test := range tests {
		insertedTracksCount, upToDate := crawler.batchPersistTrackRecords(test.trackRecords)
		if insertedTracksCount != test.expectedInsertedTracksCount {
			t.Errorf("Crawler batchPersistTrackRecords(%q): got insertedTracksCount: `%d`, "+
				"expected: `%d`", test.trackRecords, insertedTracksCount, test.expectedInsertedTracksCount)
		}

		if upToDate != test.expectedUpToDate {
			t.Errorf("Crawler batchPersistTrackRecords(%q): got upToDate: `%v`, "+
				"expected: `%v`", test.trackRecords, upToDate, test.expectedUpToDate)
		}
	}
}
