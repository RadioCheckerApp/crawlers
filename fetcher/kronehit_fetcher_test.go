package fetcher

import (
	"errors"
	"github.com/RadioCheckerApp/api/model"
	"reflect"
	"testing"
	"time"
)

var timeFormatStr = "2006-01-02 15:04"
var location, _ = time.LoadLocation("Europe/Vienna")
var timeCorrection = 7 * time.Minute

type MockKronehitAPI struct{}

func (api MockKronehitAPI) GetItems(date time.Time) (KronehitItems, error) {
	items0 := KronehitItems{
		[]KronehitItem{
			{"05:16", "DENNIS LLOYD", "NEVERMIND", ""},
			{"05:19", "AXWELL & INGROSSO", "DREAMER", ""},
			{"05:23", "JASON MRAZ", "HAVE IT ALL", ""},
			{"05:26", "STROMAE", "ALORS ON DANSE", ""},
			{"05:29", "GEORGE EZRA", "SHOTGUN", ""},
			{"05:32", "ENRIQUE IGLESIAS", "SÚBEME LA RADIO", ""},
			{"05:35", "KYGO & MIGUEL", "REMIND ME TO FORGET", ""},
			{"05:40", "SHAWN MENDES", "NERVOUS", ""},
		},
	}

	items1 := KronehitItems{
		[]KronehitItem{
			{"05:04", "KYGO & SELENA GOMEZ", "IT AIN'T ME", ""},
			{"05:08", "PINK", "SECRETS", ""},
			{"05:11", "MAGIC!", "RUDE", ""},
			{"05:16", "DENNIS LLOYD", "NEVERMIND", ""},
			{"05:19", "AXWELL & INGROSSO", "DREAMER", ""},
			{"05:23", "JASON MRAZ", "HAVE IT ALL", ""},
			{"05:26", "STROMAE", "ALORS ON DANSE", ""},
		},
	}

	items2 := KronehitItems{
		[]KronehitItem{
			{"20:47", "ROBIN SCHULZ", "OH CHILD", ""},
			{"20:50", "BASTILLE", "POMPEII", ""},
			{"20:53", "LOST FREQUENCIES & ZONDERLING", "CRAZY", ""},
			{"20:56", "MIKE POSNER", "I TOOK A PILL IN IBIZA", ""},
			{"02:02", "NAMIKA", "JE NE PARLE PAS FRANCAIS", ""},
			{"02:05", "THE FAIM", "SUMMER IS A CURSE", ""},
			{"02:08", "LAUV", "CHASING FIRE", ""},
		},
	}

	items3 := KronehitItems{
		[]KronehitItem{
			{"23:51", "NAMIKA", "JE NE PARLE PAS FRANCAIS", ""},
			{"23:54", "SIA", "CHEAP THRILLS", ""},
			{"23:58", "NICKY JAM", "EL PERDON", ""},
			{"00:03", "JONAS BLUE", "RISE", ""},
			{"00:06", "IMANY", "DON'T BE SO SHY", ""},
		},
	}

	if date.Equal(time.Time{}) {
		return KronehitItems{}, errors.New("error triggered for testing purposes")
	}

	loopTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:16", location)
	if date.Equal(loopTime.Add(-timeCorrection)) {
		return items1, nil
	}

	if date.Hour() == 23 && date.Minute() == 58 ||
		date.Hour() == 23 && date.Minute() == 48 {
		// midnight2
		return items3, nil
	}

	if date.Hour() == 1 || date.Hour() == 23 {
		// midnight1
		return items2, nil
	}

	return items0, nil
}

type MockKronehitAPIMidnightLoop struct {
	loopCounter int
}

func (api *MockKronehitAPIMidnightLoop) GetItems(date time.Time) (KronehitItems, error) {
	defer func() { api.loopCounter++ }()

	item0 := KronehitItems{
		[]KronehitItem{
			{"00:05", "JUSTIN TIMBERLAKE", "SAY SOMETHING", ""},
			{"00:09", "CALVIN HARRIS", "PROMISES", ""},
			{"00:13", "PASSENGER", "LET HER GO", ""},
			{"00:18", "KYGO & MIGUEL ", "REMIND ME TO FORGET", ""},
			{"00:22", "EL PROFESOR", "BELLA CIAO", ""},
			{"00:25", "MARSHMELLO", "HAPPIER", ""},
			{"00:28", "BEYONCE", "CRAZY IN LOVE", ""},
		}}

	item1 := KronehitItems{
		[]KronehitItem{
			{"23:45:", "EMELI SANDE", "READ ALL ABOUT IT", ""},
			{"23:50", "ROBIN SCHULZ", "OH CHILD", ""},
			{"23:53", "SEAN PAUL", "NO LIE", ""},
			{"23:56", "P.DIDDY", "I'LL BE MISSING YOU", ""},
			{"00:02", "DYNORO & GIGI D'AGOSTINO", "IN MY MIND", ""},
			{"00:05", "JUSTIN TIMBERLAKE", "SAY SOMETHING", ""},
			{"00:09", "CALVIN HARRIS", "PROMISES", ""},
		}}

	item2 := KronehitItems{
		[]KronehitItem{
			{"23:28", "RIHANNA", "ONLY GIRL", ""},
			{"23:31", "LOUD LUXURY", "BODY", ""},
			{"23:34", "MAJOR LAZER", "COLD WATER", ""},
			{"23:37", "ED SHEERAN", "HAPPIER", ""},
			{"23:42", "FELIX JAEHN", "JENNIE", ""},
			{"23:45", "EMELI SANDE", "READ ALL ABOUT IT", ""},
			{"23:50", "ROBIN SCHULZ", "OH CHILD", ""},
		},
	}

	if api.loopCounter == 0 {
		return item0, nil
	}
	if api.loopCounter == 1 {
		return item1, nil
	}
	return item2, nil
}

var kronehitStationId = "kronehit"

var kronehitExpectedTrackRecords0 = []*model.TrackRecord{
	{kronehitStationId, 1538191740, "track", model.Track{"GEORGE EZRA", "SHOTGUN"}},
	{kronehitStationId, 1538191560, "track", model.Track{"STROMAE", "ALORS ON DANSE"}},
	{kronehitStationId, 1538191380, "track", model.Track{"JASON MRAZ", "HAVE IT ALL"}},
	{kronehitStationId, 1538191140, "track", model.Track{"AXWELL & INGROSSO", "DREAMER"}},
	{kronehitStationId, 1538190960, "track", model.Track{"DENNIS LLOYD", "NEVERMIND"}},
}

type KronehitFetcherTest struct {
	fetcher               KronehitFetcher
	expectedTrackRecords  []*model.TrackRecord
	expectedNextFetchTime time.Time
	expectedErr           bool
}

var nextFetchTime, _ = time.ParseInLocation(timeFormatStr, "2018-09-29 05:30", location)

func TestKronehitFetcher_Next_Basic(t *testing.T) {
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:16", location)
	var tests = []KronehitFetcherTest{
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTime.Add(-kronehitTimeCorrection), 0},
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
	{kronehitStationId, 1538190660, "track", model.Track{"MAGIC!", "RUDE"}},
	{kronehitStationId, 1538190480, "track", model.Track{"PINK", "SECRETS"}},
	{kronehitStationId, 1538190240, "track", model.Track{"KYGO & SELENA GOMEZ", "IT AIN'T ME"}},
}

func TestKronehitFetcher_Next_Loop(t *testing.T) {
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:04", location)
	test := KronehitFetcherTest{
		KronehitFetcher{MockKronehitAPI{}, nextFetchTime.Add(-kronehitTimeCorrection), 0},
		kronehitExpectedTrackRecords1,
		expectedNextFetchTime,
		false,
	}

	test.fetcher.Next() // first iteration, skip it
	runKronehitTest(test, t)
}

var kronehitExpectedTrackRecords2 = []*model.TrackRecord{
	{kronehitStationId, 1538179680, "track", model.Track{"LAUV", "CHASING FIRE"}},
	{kronehitStationId, 1538179500, "track", model.Track{"THE FAIM", "SUMMER IS A CURSE"}},
	{kronehitStationId, 1538179320, "track", model.Track{"NAMIKA", "JE NE PARLE PAS FRANCAIS"}},
	{kronehitStationId, 1538160960, "track", model.Track{"MIKE POSNER", "I TOOK A PILL IN IBIZA"}},
	{kronehitStationId, 1538160780, "track", model.Track{"LOST FREQUENCIES & ZONDERLING", "CRAZY"}},
	{kronehitStationId, 1538160600, "track", model.Track{"BASTILLE", "POMPEII"}},
	{kronehitStationId, 1538160420, "track", model.Track{"ROBIN SCHULZ", "OH CHILD"}},
}

func TestKronehitFetcher_Next_Midnight_Weekend(t *testing.T) {
	nextFetchTimeAfterMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 02:06",
		location)
	nextFetchTimeBeforeMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-09-28 23:59",
		location)
	nextFetchTimeNoMidnightSpan, _ := time.ParseInLocation(timeFormatStr, "2018-10-11 02:00",
		location)
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-28 20:47", location)
	expectedNextFetchTimeNoMidnightSpan, _ := time.ParseInLocation(timeFormatStr, "2018-10-10 23:28", location)
	tests := []KronehitFetcherTest{
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTimeAfterMidnight.Add(-kronehitTimeCorrection), 0},
			kronehitExpectedTrackRecords2[1:],
			expectedNextFetchTime,
			false,
		},
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTimeBeforeMidnight.Add(-kronehitTimeCorrection), 0},
			kronehitExpectedTrackRecords2[3:],
			expectedNextFetchTime,
			false,
		},
		{
			KronehitFetcher{&MockKronehitAPIMidnightLoop{2}, nextFetchTimeNoMidnightSpan.Add(-kronehitTimeCorrection), 0},
			kronehitExpectedTrackRecordsNextMidnightLoop[7:],
			expectedNextFetchTimeNoMidnightSpan,
			false,
		},
	}

	for _, test := range tests {
		runKronehitTest(test, t)
	}
}

var kronehitExpectedTrackRecords3 = []*model.TrackRecord{
	{kronehitStationId, 1538604360, "track", model.Track{"IMANY", "DON'T BE SO SHY"}},
	{kronehitStationId, 1538604180, "track", model.Track{"JONAS BLUE", "RISE"}},
	{kronehitStationId, 1538603880, "track", model.Track{"NICKY JAM", "EL PERDON"}},
	{kronehitStationId, 1538603640, "track", model.Track{"SIA", "CHEAP THRILLS"}},
	{kronehitStationId, 1538603460, "track", model.Track{"NAMIKA", "JE NE PARLE PAS FRANCAIS"}},
}

func TestKronehitFetcher_Next_Midnight_Weekday(t *testing.T) {
	nextFetchTimeAfterMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-10-04 00:05", location)
	nextFetchTimeBeforeMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-10-03 23:55", location)
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-10-03 23:51", location)
	tests := []KronehitFetcherTest{
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTimeAfterMidnight.Add(-kronehitTimeCorrection), 0},
			kronehitExpectedTrackRecords3[1:],
			expectedNextFetchTime,
			false,
		},
		{
			KronehitFetcher{MockKronehitAPI{}, nextFetchTimeBeforeMidnight.Add(-kronehitTimeCorrection), 0},
			kronehitExpectedTrackRecords3[3:],
			expectedNextFetchTime,
			false,
		},
	}

	for _, test := range tests {
		runKronehitTest(test, t)
	}
}

var kronehitExpectedTrackRecordsNextMidnightLoop = []*model.TrackRecord{
	{kronehitStationId, 1539209880, "track", model.Track{"KYGO & MIGUEL ", "REMIND ME TO FORGET"}},
	{kronehitStationId, 1539209580, "track", model.Track{"PASSENGER", "LET HER GO"}},
	{kronehitStationId, 1539209340, "track", model.Track{"CALVIN HARRIS", "PROMISES"}},
	{kronehitStationId, 1539209100, "track", model.Track{"JUSTIN TIMBERLAKE", "SAY SOMETHING"}},

	{kronehitStationId, 1539208920, "track", model.Track{"DYNORO & GIGI D'AGOSTINO", "IN MY MIND"}},
	{kronehitStationId, 1539208560, "track", model.Track{"P.DIDDY", "I'LL BE MISSING YOU"}},
	{kronehitStationId, 1539208380, "track", model.Track{"SEAN PAUL", "NO LIE"}},
	{kronehitStationId, 1539208200, "track", model.Track{"ROBIN SCHULZ", "OH CHILD"}},
	{kronehitStationId, 1539207900, "track", model.Track{"EMELI SANDE", "READ ALL ABOUT IT"}},

	{kronehitStationId, 1539207720, "track", model.Track{"FELIX JAEHN", "JENNIE"}},
	{kronehitStationId, 1539207420, "track", model.Track{"ED SHEERAN", "HAPPIER"}},
	{kronehitStationId, 1539207240, "track", model.Track{"MAJOR LAZER", "COLD WATER"}},
	{kronehitStationId, 1539207060, "track", model.Track{"LOUD LUXURY", "BODY"}},
	{kronehitStationId, 1539206880, "track", model.Track{"RIHANNA", "ONLY GIRL"}},
}

func TestKronehitFetcher_Next_Midnight_Loop(t *testing.T) {
	nextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-10-11 00:20", location)
	nextFetchTime = nextFetchTime.Add(-kronehitTimeCorrection)
	fetcher := KronehitFetcher{&MockKronehitAPIMidnightLoop{0}, nextFetchTime, 0}

	var results []*model.TrackRecord
	for i := 0; i < 3; i++ {
		result, err := fetcher.Next()
		if err != nil {
			t.Errorf("(%v) Next(): got err (%v)", fetcher, err)
		}
		results = append(results, result...)
	}

	if !reflect.DeepEqual(results, kronehitExpectedTrackRecordsNextMidnightLoop) {
		t.Errorf("(%v) Next(): got\n(%q), expected\n(%q)",
			fetcher, results, kronehitExpectedTrackRecordsNextMidnightLoop)
	}
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
