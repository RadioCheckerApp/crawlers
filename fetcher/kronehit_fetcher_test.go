package fetcher

import (
	"errors"
	"github.com/RadioCheckerApp/api/model"
	"reflect"
	"testing"
	"time"
)

var timeFormatStr = "2006-01-02 15:04:05"
var location, _ = time.LoadLocation("Europe/Vienna")
var timeCorrection = 7 * time.Minute

type MockKronehitAPI struct{}

func (api MockKronehitAPI) GetItems(date time.Time) (KronehitItems, error) {
	items0 := KronehitItems{
		[]KronehitItem{
			{"05:16:46", "DENNIS LLOYD", "NEVERMIND", ""},
			{"05:19:23", "AXWELL & INGROSSO", "DREAMER", ""},
			{"05:23:22", "JASON MRAZ", "HAVE IT ALL", ""},
			{"05:26:03", "STROMAE", "ALORS ON DANSE", ""},
			{"05:29:19", "GEORGE EZRA", "SHOTGUN", ""},
			{"05:32:28", "ENRIQUE IGLESIAS", "SÚBEME LA RADIO", ""},
			{"05:35:50", "KYGO & MIGUEL", "REMIND ME TO FORGET", ""},
			{"05:40:29", "SHAWN MENDES", "NERVOUS", ""},
		},
	}

	items1 := KronehitItems{
		[]KronehitItem{
			{"05:04:31", "KYGO & SELENA GOMEZ", "IT AIN'T ME", ""},
			{"05:08:05", "PINK", "SECRETS", ""},
			{"05:11:29", "MAGIC!", "RUDE", ""},
			{"05:16:46", "DENNIS LLOYD", "NEVERMIND", ""},
			{"05:19:23", "AXWELL & INGROSSO", "DREAMER", ""},
			{"05:23:22", "JASON MRAZ", "HAVE IT ALL", ""},
			{"05:26:03", "STROMAE", "ALORS ON DANSE", ""},
		},
	}

	items2 := KronehitItems{
		[]KronehitItem{
			{"20:47:29", "ROBIN SCHULZ", "OH CHILD", ""},
			{"20:50:09", "BASTILLE", "POMPEII", ""},
			{"20:53:38", "LOST FREQUENCIES & ZONDERLING", "CRAZY", ""},
			{"20:56:06", "MIKE POSNER", "I TOOK A PILL IN IBIZA", ""},
			{"02:02:25", "NAMIKA", "JE NE PARLE PAS FRANCAIS", ""},
			{"02:05:23", "THE FAIM", "SUMMER IS A CURSE", ""},
			{"02:08:25", "LAUV", "CHASING FIRE", ""},
		},
	}

	items3 := KronehitItems{
		[]KronehitItem{
			{"23:51:34", "NAMIKA", "JE NE PARLE PAS FRANCAIS", ""},
			{"23:55:12", "SIA", "CHEAP THRILLS", ""},
			{"23:58:48", "NICKY JAM", "EL PERDON", ""},
			{"00:03:05", "JONAS BLUE", "RISE", ""},
			{"00:06:23", "IMANY", "DON'T BE SO SHY", ""},
		},
	}

	if date.Equal(time.Time{}) {
		return KronehitItems{}, errors.New("error triggered for testing purposes")
	}

	loopTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:16:46", location)
	if date.Equal(loopTime.Add(-timeCorrection)) {
		return items1, nil
	}

	if date.Hour() == 00 && date.Minute() == 5 ||
		date.Hour() == 23 && date.Minute() == 55 {
		// midnight2
		return items3, nil
	}

	if date.Hour() == 0 || date.Hour() == 23 {
		// midnight1
		return items2, nil
	}

	return items0, nil
}

var kronehitStationId = "kronehit"

var kronehitExpectedTrackRecords0 = []*model.TrackRecord{
	{kronehitStationId, 1538192429, "track", model.Track{"SHAWN MENDES", "NERVOUS"}},
	{kronehitStationId, 1538192150, "track", model.Track{"KYGO & MIGUEL", "REMIND ME TO FORGET"}},
	{kronehitStationId, 1538191948, "track", model.Track{"ENRIQUE IGLESIAS", "SÚBEME LA RADIO"}},
	{kronehitStationId, 1538191759, "track", model.Track{"GEORGE EZRA", "SHOTGUN"}},
	{kronehitStationId, 1538191563, "track", model.Track{"STROMAE", "ALORS ON DANSE"}},
	{kronehitStationId, 1538191402, "track", model.Track{"JASON MRAZ", "HAVE IT ALL"}},
	{kronehitStationId, 1538191163, "track", model.Track{"AXWELL & INGROSSO", "DREAMER"}},
	{kronehitStationId, 1538191006, "track", model.Track{"DENNIS LLOYD", "NEVERMIND"}},
}

type KronehitFetcherTest struct {
	fetcher               KronehitFetcher
	expectedTrackRecords  []*model.TrackRecord
	expectedNextFetchTime time.Time
	expectedErr           bool
}

var nextFetchTime, _ = time.ParseInLocation(timeFormatStr, "2018-09-29 05:30:00", location)

func TestKronehitFetcher_Next_Basic(t *testing.T) {
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:16:46", location)
	var tests = []KronehitFetcherTest{
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTime, 0},
			kronehitExpectedTrackRecords0,
			expectedNextFetchTime,
			false,
		},
		{
			KronehitFetcher{MockKronehitAPI{}, time.Time{}, 0},
			nil,
			time.Time{},
			true,
		},
	}

	for _, test := range tests {
		runKronehitTest(test, t)
	}
}

var kronehitExpectedTrackRecords1 = []*model.TrackRecord{
	{kronehitStationId, 1538190689, "track", model.Track{"MAGIC!", "RUDE"}},
	{kronehitStationId, 1538190485, "track", model.Track{"PINK", "SECRETS"}},
	{kronehitStationId, 1538190271, "track", model.Track{"KYGO & SELENA GOMEZ", "IT AIN'T ME"}},
}

func TestKronehitFetcher_Next_Loop(t *testing.T) {
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:04:31", location)
	test := KronehitFetcherTest{
		KronehitFetcher{MockKronehitAPI{}, nextFetchTime, 0},
		kronehitExpectedTrackRecords1,
		expectedNextFetchTime,
		false,
	}

	test.fetcher.Next() // first iteration, skip it
	runKronehitTest(test, t)
}

var kronehitExpectedTrackRecords2 = []*model.TrackRecord{
	{kronehitStationId, 1538179705, "track", model.Track{"LAUV", "CHASING FIRE"}},
	{kronehitStationId, 1538179523, "track", model.Track{"THE FAIM", "SUMMER IS A CURSE"}},
	{kronehitStationId, 1538179345, "track", model.Track{"NAMIKA", "JE NE PARLE PAS FRANCAIS"}},
	{kronehitStationId, 1538160966, "track", model.Track{"MIKE POSNER", "I TOOK A PILL IN IBIZA"}},
	{kronehitStationId, 1538160818, "track", model.Track{"LOST FREQUENCIES & ZONDERLING", "CRAZY"}},
	{kronehitStationId, 1538160609, "track", model.Track{"BASTILLE", "POMPEII"}},
	{kronehitStationId, 1538160449, "track", model.Track{"ROBIN SCHULZ", "OH CHILD"}},
}

func TestKronehitFetcher_Next_Midnight(t *testing.T) {
	nextFetchTimeAfterMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 00:00:00", location)
	nextFetchTimeBeforeMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-09-28 23:59:59", location)
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-28 20:47:29", location)
	kronehitTestNextMidnight(nextFetchTimeAfterMidnight, nextFetchTimeBeforeMidnight,
		expectedNextFetchTime, kronehitExpectedTrackRecords2, t)
}

var kronehitExpectedTrackRecords3 = []*model.TrackRecord{
	{kronehitStationId, 1538604383, "track", model.Track{"IMANY", "DON'T BE SO SHY"}},
	{kronehitStationId, 1538604185, "track", model.Track{"JONAS BLUE", "RISE"}},
	{kronehitStationId, 1538603928, "track", model.Track{"NICKY JAM", "EL PERDON"}},
	{kronehitStationId, 1538603712, "track", model.Track{"SIA", "CHEAP THRILLS"}},
	{kronehitStationId, 1538603494, "track", model.Track{"NAMIKA", "JE NE PARLE PAS FRANCAIS"}},
}

func TestKronehitFetcher_Next_Midnight2(t *testing.T) {
	nextFetchTimeAfterMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-10-04 00:05:00", location)
	nextFetchTimeBeforeMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-10-03 23:55:59", location)
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-10-03 23:51:34", location)
	kronehitTestNextMidnight(nextFetchTimeAfterMidnight, nextFetchTimeBeforeMidnight,
		expectedNextFetchTime, kronehitExpectedTrackRecords3, t)
}

func TestKronehitFetcher_Next_RequestLimit(t *testing.T) {
	fetcher := KronehitFetcher{MockKronehitAPI{}, nextFetchTime, 0}
	_, err := fetcher.Next()
	for i := 0; i < 10; i++ {
		fetcher.nextFetchTime = nextFetchTime
		_, err = fetcher.Next()
	}
	if err == nil {
		t.Errorf("Next(): Request limit is not obeyed. Fetcher exceeds 10 fetches.")
	}
}

func kronehitTestNextMidnight(nextFetchTimeAfterMidnight, nextFetchTimeBeforeMidnight,
	expectedNextFetchTime time.Time, expectedTrackRecords []*model.TrackRecord, t *testing.T) {
	tests := []KronehitFetcherTest{
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTimeAfterMidnight, 0},
			expectedTrackRecords,
			expectedNextFetchTime,
			false,
		},
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTimeBeforeMidnight, 0},
			expectedTrackRecords,
			expectedNextFetchTime,
			false,
		},
	}

	for _, test := range tests {
		runKronehitTest(test, t)
	}
}

func runKronehitTest(test KronehitFetcherTest, t *testing.T) {
	trackRecords, err := test.fetcher.Next()

	if (err != nil) != test.expectedErr {
		t.Errorf("(%v) Next(): got err (%v), expected err (%v)",
			test.fetcher, err, test.expectedErr)
	}

	if err != nil {
		return
	}

	if !reflect.DeepEqual(trackRecords, test.expectedTrackRecords) {
		t.Errorf("(%v) Next(): got\n(%q, %v), expected\n(%q, %v)",
			test.fetcher, trackRecords, err, test.expectedTrackRecords, test.expectedErr)
	}

	if !test.fetcher.nextFetchTime.Add(timeCorrection).Equal(test.expectedNextFetchTime) {
		t.Errorf("(%v) nextFetchTime: got (%v), expected (%v)",
			test.fetcher, test.fetcher.nextFetchTime, test.expectedNextFetchTime)
	}
}

func TestRandomizedInitialFetchTime(t *testing.T) {
	for i := 0; i < 20; i++ {
		randomizedTime := randomizedInitialFetchTime()
		lowerBound := time.Now().Add(-1 * time.Minute)
		upperBound := time.Now().Add(10 * time.Minute)
		if randomizedTime.Before(lowerBound) {
			t.Errorf("RandomizedInitialFetchTime: randomizedTime before lowerBound")
		}
		if randomizedTime.After(upperBound) {
			t.Errorf("RandomizedInitialFetchTime: randomizedTime after upperBound")
		}
	}
}
