package fetcher

import (
	"encoding/json"
	"errors"
	"github.com/RadioCheckerApp/api/model"
	"github.com/RadioCheckerApp/crawlers/util"
	"log"
	"strings"
	"time"
)

const orfRequestLimit = 1

type ORFItem struct {
	PlayedAt int64
	Title    string
}

func (item *ORFItem) toTrackRecord(config *stationConfig) (*model.TrackRecord, error) {
	splitted := strings.Split(item.Title, " - ")
	if len(splitted) != 2 {
		return nil, errors.New("could not extract title and artist")
	}

	title, artist := splitted[0], splitted[1]
	if !config.titleDashArtistFormat {
		title, artist = splitted[1], splitted[0]
	}

	return &model.TrackRecord{
		config.stationID,
		item.PlayedAt,
		"track",
		model.Track{artist, title},
	}, nil
}

type ORFDTO []*ORFItem

func (dto *ORFDTO) toTrackRecords(config *stationConfig) []*model.TrackRecord {
	var trackRecords []*model.TrackRecord
	for _, item := range *dto {
		trackRecord, err := item.toTrackRecord(config)
		if err != nil {
			log.Printf("ERROR:   Unable to extract TrackRecord from item: `%q`. Message: `%s`.",
				item.Title, err.Error())
			continue
		}
		if config.skipper(trackRecord) {
			log.Printf("INFO:    Skipping item `%s` (Airtime: %d). ", item.Title, item.PlayedAt)
			continue
		}
		trackRecords = append(trackRecords, trackRecord)
	}
	return trackRecords
}

type ORFAPI interface {
	GetItems() (ORFDTO, error)
}

type ORFAPIImplementation struct {
	client   util.HTTP
	endpoint string
}

func NewORFAPIImplementation(url string, timeout time.Duration) *ORFAPIImplementation {
	client := util.NewHTTPClient(timeout)
	return &ORFAPIImplementation{client, url}
}

func (api *ORFAPIImplementation) GetItems() (ORFDTO, error) {
	respPayload, err := api.client.Get(api.endpoint)
	if err != nil {
		return nil, err
	}

	var dto ORFDTO
	if err := json.Unmarshal(respPayload, &dto); err != nil {
		log.Printf("ERROR:   Unmarshalling JSON body failed. Message: `%s`.", err.Error())
		return nil, err
	}

	return dto, nil
}

type Skipper func(record *model.TrackRecord) bool

type stationConfig struct {
	stationID             string
	fetchURL              string
	titleDashArtistFormat bool
	skipper               Skipper
}

type ORFFetcher struct {
	orfAPI       ORFAPI
	fetchCounter int
	config       *stationConfig
}

func NewORFFetcher(stationID string) (*ORFFetcher, error) {
	config, err := findStationConfig(stationID)
	if err != nil {
		return nil, err
	}
	orfAPI := NewORFAPIImplementation(config.fetchURL, requestTimeout)
	return &ORFFetcher{orfAPI, 0, config}, nil
}

func (fetcher *ORFFetcher) Next() ([]*model.TrackRecord, error) {
	if fetcher.fetchCounter >= orfRequestLimit {
		return nil, errors.New("orf resource already fetched - exiting loop")
	}

	items, err := fetcher.orfAPI.GetItems()
	if err != nil {
		log.Printf("ERROR:   Unable to fetch items from ORF. Message: `%s`.", err.Error())
		return nil, err
	}

	log.Printf("INFO:    Fetched %d items from ORF.", len(items))

	trackRecords := items.toTrackRecords(fetcher.config)

	if len(trackRecords) == 0 {
		log.Printf("WARNING: Unable to extract any TrackRecords from %d items. SkipRate = %."+
			"2f%%", len(items), calculateSkipRate(len(trackRecords), len(items)))
		return nil, errors.New("unable to extract any trackRecords")
	}

	fetcher.fetchCounter++

	log.Printf("INFO:    Returned %d TrackRecords, extracted from %d items. SkipRate = %.2f%%",
		len(trackRecords), len(items), calculateSkipRate(len(trackRecords), len(items)))
	return trackRecords, nil
}

func findStationConfig(stationID string) (*stationConfig, error) {
	switch stationID {
	case "radio-oe1":
		return &stationConfig{
			stationID: "radio-oe1",
			fetchURL:  "http://mp3ooe1.apasf.sf.apa.at/played.html?type=json",
			skipper: func(record *model.TrackRecord) bool {
				// skip everything at the moment since Ö1 doesn't really play tracks
				// [FORMAT: "Jetzt in Ö1: Songs mit Wah Wah, Brumm und Boing!"]
				return true
			},
		}, nil
	case "hitradio-oe3":
		return &stationConfig{
			stationID:             "hitradio-oe3",
			fetchURL:              "http://mp3oe3.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: false,
			skipper: func(record *model.TrackRecord) bool {
				// skip all tracks that have "Hitradio Ö3" as artist
				// [FORMAT: "Hitradio Ö3 - LiveStream"]
				return record.Artist == "Hitradio Ö3"
			},
		}, nil
	case "radio-burgenland":
		return &stationConfig{
			stationID:             "radio-burgenland",
			fetchURL:              "http://mp3burgenland.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// skip all tracks that have "Radio Burgenland" as title
				// [FORMAT: "Radio Burgenland - Da bin ich daheim"]
				return record.Title == "Radio Burgenland"
			},
		}, nil
	case "radio-kärnten":
		return &stationConfig{
			stationID:             "radio-kaernten",
			fetchURL:              "http://mp3kaernten.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// nothing to skip since meta tracks are causing toTrackRecord()
				// to throw an error anyways [FORMAT: "Radio Kärnten - Mein Daheim -"]
				return record.Title == "Radio Kärnten"
			},
		}, nil
	case "radio-niederösterreich":
		return &stationConfig{
			stationID:             "radio-niederoesterreich",
			fetchURL:              "http://mp3noe.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// nothing to skip since meta tracks are causing toTrackRecord()
				// to throw an error anyways [FORMAT: "DIE GROESSTEN HITS UND SCHOENSTEN OLDIES"]
				return false
			},
		}, nil
	case "radio-oberösterreich":
		return &stationConfig{
			stationID:             "radio-oberoesterreich",
			fetchURL:              "http://mp3ooe.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// skip all tracks that have "RADIO OÖ" as title
				// [FORMAT: "RADIO OÖ - MEIN LAND. MEIN RADIO"]
				return record.Title == "RADIO OÖ"
			},
		}, nil
	case "radio-salzburg":
		return &stationConfig{
			stationID:             "radio-salzburg",
			fetchURL:              "http://mp3salzburg.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// nothing to skip since meta tracks are causing toTrackRecord()
				// to throw an error anyways [FORMAT: "Radio Salzburg"]
				return true
			},
		}, nil
	case "radio-tirol":
		return &stationConfig{
			stationID:             "radio-tirol",
			fetchURL:              "http://mp3tirol.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// nothing to skip since meta tracks are causing toTrackRecord()
				// to throw an error anyways
				// [FORMAT: "ORF Radio Tirol | meine Musik mein Land mein Radio!"]
				return true
			},
		}, nil
	case "radio-vorarlberg":
		return &stationConfig{
			stationID:             "radio-vorarlberg",
			fetchURL:              "http://mp3vlbg.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// nothing to skip since meta tracks are causing toTrackRecord()
				// to throw an error anyways
				// [FORMAT: "ORF Radio Vorarlberg | Die 70er, die 80er und die schönsten Songs bis heute!"]
				return true
			},
		}, nil
	case "radio-wien":
		return &stationConfig{
			stationID:             "radio-wien",
			fetchURL:              "http://mp3wien2.apasf.sf.apa.at/played.html?type=json",
			titleDashArtistFormat: true,
			skipper: func(record *model.TrackRecord) bool {
				// nothing to skip since meta tracks are causing toTrackRecord()
				// to throw an error anyways [FORMAT: "Radio Wien   Einfach gute Musik"]
				return true
			},
		}, nil
	}
	return nil, errors.New("unsupported station")
}
