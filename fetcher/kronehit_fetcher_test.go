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

var mockKronehitAPIMidnightLoopCounter = 0

type MockKronehitAPIMidnightLoop struct{}

func (api MockKronehitAPIMidnightLoop) GetItems(date time.Time) (KronehitItems, error) {
	defer func() { mockKronehitAPIMidnightLoopCounter++ }()

	item0 := KronehitItems{
		[]KronehitItem{
			{"00:05:06", "JUSTIN TIMBERLAKE", "SAY SOMETHING", ""},
			{"00:09:23", "CALVIN HARRIS", "PROMISES", ""},
			{"00:13:12", "PASSENGER", "LET HER GO", ""},
			{"00:18:50", "KYGO & MIGUEL ", "REMIND ME TO FORGET", ""},
			{"00:22:23", "EL PROFESOR", "BELLA CIAO", ""},
			{"00:25:16", "MARSHMELLO", "HAPPIER", ""},
			{"00:28:59", "BEYONCE", "CRAZY IN LOVE", ""},
		}}

	item1 := KronehitItems{
		[]KronehitItem{
			{"23:45:44", "EMELI SANDE", "READ ALL ABOUT IT", ""},
			{"23:50:31", "ROBIN SCHULZ", "OH CHILD", ""},
			{"23:53:58", "SEAN PAUL", "NO LIE", ""},
			{"23:56:42", "P.DIDDY", "I'LL BE MISSING YOU", ""},
			{"00:02:09", "DYNORO & GIGI D'AGOSTINO", "IN MY MIND", ""},
			{"00:05:06", "JUSTIN TIMBERLAKE", "SAY SOMETHING", ""},
			{"00:09:23", "CALVIN HARRIS", "PROMISES", ""},
		}}

	item2 := KronehitItems{
		[]KronehitItem{
			{"23:28:05", "RIHANNA", "ONLY GIRL", ""},
			{"23:31:55", "LOUD LUXURY", "BODY", ""},
			{"23:34:52", "MAJOR LAZER", "COLD WATER", ""},
			{"23:37:58", "ED SHEERAN", "HAPPIER", ""},
			{"23:42:34", "FELIX JAEHN", "JENNIE", ""},
			{"23:45:44", "EMELI SANDE", "READ ALL ABOUT IT", ""},
			{"23:50:31", "ROBIN SCHULZ", "OH CHILD", ""},
		},
	}

	if mockKronehitAPIMidnightLoopCounter == 0 {
		return item0, nil
	}
	if mockKronehitAPIMidnightLoopCounter == 1 {
		return item1, nil
	}
	return item2, nil
}

var kronehitStationId = "kronehit"

var kronehitExpectedTrackRecords0 = []*model.TrackRecord{
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
	{kronehitStationId, 1538190689, "track", model.Track{"MAGIC!", "RUDE"}},
	{kronehitStationId, 1538190485, "track", model.Track{"PINK", "SECRETS"}},
	{kronehitStationId, 1538190271, "track", model.Track{"KYGO & SELENA GOMEZ", "IT AIN'T ME"}},
}

func TestKronehitFetcher_Next_Loop(t *testing.T) {
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 05:04:31", location)
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
	{kronehitStationId, 1538179705, "track", model.Track{"LAUV", "CHASING FIRE"}},
	{kronehitStationId, 1538179523, "track", model.Track{"THE FAIM", "SUMMER IS A CURSE"}},
	{kronehitStationId, 1538179345, "track", model.Track{"NAMIKA", "JE NE PARLE PAS FRANCAIS"}},
	{kronehitStationId, 1538160966, "track", model.Track{"MIKE POSNER", "I TOOK A PILL IN IBIZA"}},
	{kronehitStationId, 1538160818, "track", model.Track{"LOST FREQUENCIES & ZONDERLING", "CRAZY"}},
	{kronehitStationId, 1538160609, "track", model.Track{"BASTILLE", "POMPEII"}},
	{kronehitStationId, 1538160449, "track", model.Track{"ROBIN SCHULZ", "OH CHILD"}},
}

func TestKronehitFetcher_Next_Midnight_Weekend(t *testing.T) {
	nextFetchTimeAfterMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-09-29 02:06:00",
		location)
	nextFetchTimeBeforeMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-09-28 23:59:59",
		location)
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-09-28 20:47:29", location)
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
	}

	for _, test := range tests {
		runKronehitTest(test, t)
	}
}

var kronehitExpectedTrackRecords3 = []*model.TrackRecord{
	{kronehitStationId, 1538604383, "track", model.Track{"IMANY", "DON'T BE SO SHY"}},
	{kronehitStationId, 1538604185, "track", model.Track{"JONAS BLUE", "RISE"}},
	{kronehitStationId, 1538603928, "track", model.Track{"NICKY JAM", "EL PERDON"}},
	{kronehitStationId, 1538603712, "track", model.Track{"SIA", "CHEAP THRILLS"}},
	{kronehitStationId, 1538603494, "track", model.Track{"NAMIKA", "JE NE PARLE PAS FRANCAIS"}},
}

func TestKronehitFetcher_Next_Midnight_Weekday(t *testing.T) {
	nextFetchTimeAfterMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-10-04 00:05:00", location)
	nextFetchTimeBeforeMidnight, _ := time.ParseInLocation(timeFormatStr, "2018-10-03 23:55:59", location)
	expectedNextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-10-03 23:51:34", location)
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
	{kronehitStationId, 1539209930, "track", model.Track{"KYGO & MIGUEL ", "REMIND ME TO FORGET"}},
	{kronehitStationId, 1539209592, "track", model.Track{"PASSENGER", "LET HER GO"}},
	{kronehitStationId, 1539209363, "track", model.Track{"CALVIN HARRIS", "PROMISES"}},
	{kronehitStationId, 1539209106, "track", model.Track{"JUSTIN TIMBERLAKE", "SAY SOMETHING"}},

	{kronehitStationId, 1539208929, "track", model.Track{"DYNORO & GIGI D'AGOSTINO", "IN MY MIND"}},
	{kronehitStationId, 1539208602, "track", model.Track{"P.DIDDY", "I'LL BE MISSING YOU"}},
	{kronehitStationId, 1539208438, "track", model.Track{"SEAN PAUL", "NO LIE"}},
	{kronehitStationId, 1539208231, "track", model.Track{"ROBIN SCHULZ", "OH CHILD"}},
	{kronehitStationId, 1539207944, "track", model.Track{"EMELI SANDE", "READ ALL ABOUT IT"}},

	{kronehitStationId, 1539207754, "track", model.Track{"FELIX JAEHN", "JENNIE"}},
	{kronehitStationId, 1539207478, "track", model.Track{"ED SHEERAN", "HAPPIER"}},
	{kronehitStationId, 1539207292, "track", model.Track{"MAJOR LAZER", "COLD WATER"}},
	{kronehitStationId, 1539207115, "track", model.Track{"LOUD LUXURY", "BODY"}},
	{kronehitStationId, 1539206885, "track", model.Track{"RIHANNA", "ONLY GIRL"}},
}

func TestKronehitFetcher_Next_Midnight_Loop(t *testing.T) {
	nextFetchTime, _ := time.ParseInLocation(timeFormatStr, "2018-10-11 00:20:00", location)
	nextFetchTime = nextFetchTime.Add(-kronehitTimeCorrection)
	fetcher := KronehitFetcher{MockKronehitAPIMidnightLoop{}, nextFetchTime, 0}

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
