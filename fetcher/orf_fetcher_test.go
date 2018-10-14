package fetcher

import (
	"github.com/RadioCheckerApp/api/model"
	"reflect"
	"testing"
)

type MockORFAPI struct{}

func (api *MockORFAPI) GetItems() (ORFDTO, error) {
	return ORFDTO{
		{PlayedAt: 1539532794, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539532528, Title: "NOTHING'S GONNA STOP US NOW - STARSHIP"},
		{PlayedAt: 1539532508, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539532242, Title: "LOVE CHANGES EVERYTHING - CLIMIE FISHER"},
		{PlayedAt: 1539531997, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539531786, Title: "DIE WELT (live) - AUSTRIA 3"},
		{PlayedAt: 1539531595, Title: "DREAMIN' - CLIFF RICHARD"},
		{PlayedAt: 1539531375, Title: "DO THE LIMBO DANCE - DAVID HASSELHOFF"},
		{PlayedAt: 1539530999, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539530784, Title: "TU'S DOCH TU'S - FRANCINE JORDI"},
		{PlayedAt: 1539530769, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539530543, Title: "L' ITALIANO - TOTO CUTUGNO"},
		{PlayedAt: 1539530292, Title: "KANSAS CITY - LES HUMPHRIES SINGERS"},
		{PlayedAt: 1539530187, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539530006, Title: "BLUE NIGHT SHADOW - TWO OF US"},
		{PlayedAt: 1539530001, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539529791, Title: "MANCHMAL DENK I NO AN DI - RAINHARD FENDRICH"},
		{PlayedAt: 1539529565, Title: "LOVE TOUCH - ROD STEWART"},
		{PlayedAt: 1539529189, Title: "RADIO OÖ - MEIN LAND. MEIN RADIO"},
		{PlayedAt: 1539529034, Title: "A SUMMER SONG - CHAD & JEREMY"},
	}, nil
}

type MockHTTPClient struct{}

func (client *MockHTTPClient) Get(url string) ([]byte, error) {
	return []byte(`[{"playedat": 1539532794,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539532528,"title": "NOTHING'S GONNA STOP US NOW - STARSHIP","metadata": {"tit2": "NOTHING'S GONNA STOP US NOW - STARSHIP"}},{"playedat": 1539532508,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539532242,"title": "LOVE CHANGES EVERYTHING - CLIMIE FISHER","metadata": {"tit2": "LOVE CHANGES EVERYTHING - CLIMIE FISHER"}},{"playedat": 1539531997,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539531786,"title": "DIE WELT (live) - AUSTRIA 3","metadata": {"tit2": "DIE WELT (live) - AUSTRIA 3"}},{"playedat": 1539531595,"title": "DREAMIN' - CLIFF RICHARD","metadata": {"tit2": "DREAMIN' - CLIFF RICHARD"}},{"playedat": 1539531375,"title": "DO THE LIMBO DANCE - DAVID HASSELHOFF","metadata": {"tit2": "DO THE LIMBO DANCE - DAVID HASSELHOFF"}},{"playedat": 1539530999,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539530784,"title": "TU'S DOCH TU'S - FRANCINE JORDI","metadata": {"tit2": "TU'S DOCH TU'S - FRANCINE JORDI"}},{"playedat": 1539530769,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539530543,"title": "L' ITALIANO - TOTO CUTUGNO","metadata": {"tit2": "L' ITALIANO - TOTO CUTUGNO"}},{"playedat": 1539530292,"title": "KANSAS CITY - LES HUMPHRIES SINGERS","metadata": {"tit2": "KANSAS CITY - LES HUMPHRIES SINGERS"}},{"playedat": 1539530187,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539530006,"title": "BLUE NIGHT SHADOW - TWO OF US","metadata": {"tit2": "BLUE NIGHT SHADOW - TWO OF US"}},{"playedat": 1539530001,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539529791,"title": "MANCHMAL DENK I NO AN DI - RAINHARD FENDRICH","metadata": {"tit2": "MANCHMAL DENK I NO AN DI - RAINHARD FENDRICH"}},{"playedat": 1539529565,"title": "LOVE TOUCH - ROD STEWART","metadata": {"tit2": "LOVE TOUCH - ROD STEWART"}},{"playedat": 1539529189,"title": "RADIO OÖ - MEIN LAND. MEIN RADIO","metadata": {"tit2": "RADIO OÖ - MEIN LAND. MEIN RADIO"}},{"playedat": 1539529034,"title": "A SUMMER SONG - CHAD & JEREMY","metadata": {"tit2": "A SUMMER SONG - CHAD & JEREMY"}}]`), nil
}

func TestORFAPIImplementation_GetItems(t *testing.T) {
	apiMock := MockORFAPI{}
	expectedItems, _ := apiMock.GetItems()

	orfAPI := ORFAPIImplementation{&MockHTTPClient{}, "localhost"}
	resultItems, err := orfAPI.GetItems()

	if err != nil {
		t.Errorf("(%q) GetItems(): unexpected error `%v`", orfAPI, err)
	}

	if !reflect.DeepEqual(expectedItems, resultItems) {
		t.Errorf("(%q) GetItems(): expected\n%q, got\n%q", orfAPI, expectedItems, resultItems)
	}
}

var orfExpectedTrackRecords = []*model.TrackRecord{
	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539532528,
		Track: model.Track{Title: "NOTHING'S GONNA STOP US NOW", Artist: "STARSHIP"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539532242,
		Track: model.Track{Title: "LOVE CHANGES EVERYTHING", Artist: "CLIMIE FISHER"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539531786,
		Track: model.Track{Title: "DIE WELT (live)", Artist: "AUSTRIA 3"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539531595,
		Track: model.Track{Title: "DREAMIN'", Artist: "CLIFF RICHARD"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539531375,
		Track: model.Track{Title: "DO THE LIMBO DANCE", Artist: "DAVID HASSELHOFF"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539530784,
		Track: model.Track{Title: "TU'S DOCH TU'S", Artist: "FRANCINE JORDI"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539530543,
		Track: model.Track{Title: "L' ITALIANO", Artist: "TOTO CUTUGNO"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539530292,
		Track: model.Track{Title: "KANSAS CITY", Artist: "LES HUMPHRIES SINGERS"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539530006,
		Track: model.Track{Title: "BLUE NIGHT SHADOW", Artist: "TWO OF US"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539529791,
		Track: model.Track{Title: "MANCHMAL DENK I NO AN DI", Artist: "RAINHARD FENDRICH"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539529565,
		Track: model.Track{Title: "LOVE TOUCH", Artist: "ROD STEWART"}},

	{StationId: "radio-oberoesterreich", Type: "track", Timestamp: 1539529034,
		Track: model.Track{Title: "A SUMMER SONG", Artist: "CHAD & JEREMY"}},
}

func TestORFFetcher_Next(t *testing.T) {
	orfFetcher := ORFFetcher{
		&MockORFAPI{},
		0,
		&stationConfig{
			stationID:             "radio-oberoesterreich",
			fetchURL:              "http://mp3ooe.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				return record.Title == "RADIO OÖ"
			},
		},
	}

	result, err := orfFetcher.Next()
	if err != nil {
		t.Errorf("(%v) GetItems(): unexpected error `%v`", orfFetcher, err)
	}

	if !reflect.DeepEqual(orfExpectedTrackRecords, result) {
		t.Errorf("(%v) Next(): expected\n%q, got\n%q", orfFetcher, orfExpectedTrackRecords, result)
	}
}

func TestORFFetcher_Next_Loop(t *testing.T) {
	orfFetcher := ORFFetcher{
		&MockORFAPI{},
		0,
		&stationConfig{
			stationID:             "radio-oberoesterreich",
			fetchURL:              "http://mp3ooe.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				return record.Title == "RADIO OÖ"
			},
		},
	}

	orfFetcher.Next() // skip results
	if _, err := orfFetcher.Next(); err == nil {
		t.Errorf("(%v) GetItems(): expected error, got none", orfFetcher)
	}
}

func TestFindStationConfig(t *testing.T) {
	urlsByID := map[string]string{
		"radio-oe1":              "http://mp3ooe1.apasf.sf.apa.at/played.html?type=json",
		"hitradio-oe3":           "http://mp3oe3.apasf.sf.apa.at/played.html?type=json",
		"radio-burgenland":       "http://mp3burgenland.apasf.sf.apa.at/played.html?type=json",
		"radio-kärnten":          "http://mp3kaernten.apasf.sf.apa.at/played.html?type=json",
		"radio-niederösterreich": "http://mp3noe.apasf.sf.apa.at/played.html?type=json",
		"radio-oberösterreich":   "http://mp3ooe.apasf.sf.apa.at/played.html?type=json",
		"radio-salzburg":         "http://mp3salzburg.apasf.sf.apa.at/played.html?type=json",
		"radio-tirol":            "http://mp3tirol.apasf.sf.apa.at/played.html?type=json",
		"radio-vorarlberg":       "http://mp3vlbg.apasf.sf.apa.at/played.html?type=json",
		"radio-wien":             "http://mp3wien2.apasf.sf.apa.at/played.html?type=json",
	}

	for k, v := range urlsByID {
		result, err := findStationConfig(k)
		if err != nil {
			t.Errorf("findStationConfig(%s): unexpected error `%v`", k, err)
			continue
		}
		if result.fetchURL != v {
			t.Errorf("findStationConfig(%s): expected url `%s`, got `%s`", k, v, result.fetchURL)
		}
	}

	if _, err := findStationConfig("not-found"); err == nil {
		t.Errorf("findStationConfig(not-found): expected error, got none")
	}
}
