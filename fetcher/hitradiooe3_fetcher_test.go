package fetcher

import (
	"errors"
	"github.com/ChimeraCoder/anaconda"
	"github.com/RadioCheckerApp/api/model"
	"net/url"
	"reflect"
	"testing"
)

type MockTwitterAPI struct{}

func (api MockTwitterAPI) GetUserTimeline(v url.Values) ([]anaconda.Tweet, error) {
	tweets0 := []anaconda.Tweet{
		{
			Text:      "16:39: \"River\" von Eminem feat. Ed Sheeran",
			CreatedAt: "Mon Aug 26 09:39:00 -0700 2018",
			IdStr:     "1",
		},
		{
			Text:      "16:35: \"Last Friday Night\" von Katy Perry",
			CreatedAt: "Mon Aug 26 09:35:00 -0700 2018",
			IdStr:     "2",
		},
		{
			Text:      "16:32: \"Hey Jessy\" von Simon Lewis",
			CreatedAt: "Mon Aug 26 09:32:00 -0700 2018",
			IdStr:     "3",
		},
	}

	tweets1 := []anaconda.Tweet{
		{
			Text:      "16:32: \"Hey Jessy\" von Simon Lewis",
			CreatedAt: "Mon Aug 26 09:32:00 -0700 2018",
			IdStr:     "3",
		},
		{
			Text:      "16:25: \"Sign of the Times\" von Harry Styles",
			CreatedAt: "Mon Aug 26 09:25:00 -0700 2018",
			IdStr:     "4",
		},
		{
			Text:      "16:22: \"Faded\" von Alan Walker",
			CreatedAt: "Mon Aug 26 09:22:00 -0700 2018",
			IdStr:     "5",
		},
	}

	if v.Get("error") == "ok" {
		return nil, errors.New("error")
	}
	if v.Get("max_id") == "3" {
		return tweets1, nil
	}
	return tweets0, nil
}

func TestNewHitradioOE3Fetcher(t *testing.T) {
	var tests = []struct {
		consumerKey       string
		consumerKeySecret string
		accessToken       string
		accessTokenSecret string
		expectedErr       bool
	}{
		{"abcdefg", "abcdefg", "abcdefg", "abcdefg", false},
		{"", "", "", "", true},
	}

	for _, test := range tests {
		_, err := NewHitradioOE3Fetcher(test.consumerKey, test.consumerKeySecret,
			test.accessToken, test.accessTokenSecret)
		if (err != nil) != test.expectedErr {
			t.Errorf("NewHitradioOE3Fetcher(%q, %q, %q, %q): got error: (%v), expected error: (%v)",
				test.consumerKey, test.consumerKeySecret, test.accessToken,
				test.accessTokenSecret, err != nil, test.expectedErr)
		}
	}
}

var stationId = "hitradio-oe3"

var expectedTrackRecords = []model.TrackRecord{
	{stationId, 1535301540, "track", model.Track{"Eminem feat. Ed Sheeran", "River"}},
	{stationId, 1535301300, "track", model.Track{"Katy Perry", "Last Friday Night"}},
	{stationId, 1535301120, "track", model.Track{"Simon Lewis", "Hey Jessy"}},
}

type FetcherTest struct {
	fetcher              HitradioOE3Fetcher
	expectedTrackRecords []model.TrackRecord
	expectedMaxID        string
	expectedErr          bool
}

func TestHitradioOE3Fetcher_Next_Basic(t *testing.T) {
	var tests = []FetcherTest{
		{
			HitradioOE3Fetcher{MockTwitterAPI{}, url.Values{}},
			expectedTrackRecords,
			"3",
			false,
		},
		{
			HitradioOE3Fetcher{MockTwitterAPI{}, url.Values{"error": []string{"ok"}}},
			[]model.TrackRecord{},
			"X",
			true,
		},
	}

	for _, test := range tests {
		runTest(test, t)
	}
}

func TestHitradioOE3Fetcher_Next_Loop(t *testing.T) {
	test := FetcherTest{
		HitradioOE3Fetcher{MockTwitterAPI{}, url.Values{}},
		[]model.TrackRecord{
			{stationId, 1535300700, "track", model.Track{"Harry Styles", "Sign of the Times"}},
			{stationId, 1535300520, "track", model.Track{"Alan Walker", "Faded"}},
		},
		"5",
		false,
	}

	test.fetcher.Next() // first iteration, skip it
	runTest(test, t)
}

func runTest(test FetcherTest, t *testing.T) {
	trackRecords, err := test.fetcher.Next()

	if (err != nil) != test.expectedErr {
		t.Errorf("(%q) Next(): got err (%v), expected err (%v)",
			test.fetcher, err, test.expectedErr)
	}

	if err != nil {
		return
	}

	if !reflect.DeepEqual(trackRecords, test.expectedTrackRecords) {
		t.Errorf("(%q) Next(): got\n(%q, %v), expected\n(%q, %v)",
			test.fetcher, trackRecords, err, test.expectedTrackRecords, test.expectedErr)
	}

	if test.fetcher.twitterAPIParams.Get("max_id") != test.expectedMaxID {
		t.Errorf("(%q) twitterAPIParams.Get(\"max_id\"): got (%s), expected (%s)",
			test.fetcher, test.fetcher.twitterAPIParams.Get("max_id"), test.expectedMaxID)
	}
}
