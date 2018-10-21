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

var trackRecordBatch3 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301000, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	// a track with critical proximity to the latestTrackRecord, therefore ignore it and finish successfully
	{"station-a", 1234567900, "track", model.Track{"Skipped", "Track"}},
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
		{trackRecordBatch3, 2, true},
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

var duplicateTrackRecords0 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

var duplicateTrackRecords1 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301640, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

var duplicateTrackRecords2 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301001, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

var duplicateTrackRecords3 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
	{"station-a", 1535301320, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

var duplicateTrackRecords4 = []*model.TrackRecord{
	{"station-a", 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301444, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{"station-a", 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{"station-a", 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
	{"station-a", 1535301000, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
	// a track that is older than the latestTrackRecord and therefore should cause the loop to exit
	{"station-a", 1, "track", model.Track{"Old Test", "Track"}},
	{"station-a", 12345, "track", model.Track{"Ignored", "Track"}},
}

func TestCrawler_filterDuplicates(t *testing.T) {
	var tests = [][]*model.TrackRecord{
		duplicateTrackRecords0,
		duplicateTrackRecords1,
		duplicateTrackRecords2,
		duplicateTrackRecords3,
		duplicateTrackRecords4,
	}

	crawler := Crawler{latestTrackRecordTimestamp: 1234}
	for _, test := range tests {
		result := crawler.filterDuplicates(test)
		if !reflect.DeepEqual(result, duplicateTrackRecords0) {
			t.Errorf("filterDuplicates(%q): got\n%q, expected\n%q", test, result, duplicateTrackRecords0)
		}
	}
}

func TestAbs(t *testing.T) {
	var tests = map[int64]int64{
		0:     0,
		-1:    1,
		1:     1,
		-100:  100,
		100:   100,
		-1000: 1000,
		1000:  1000,
		-123:  123,
		123:   123,
	}

	for testValue, expected := range tests {
		result := abs(testValue)
		if result != expected {
			t.Errorf("abs(%d): expected `%d`, got `%d`", testValue, expected, result)
		}
	}
}
