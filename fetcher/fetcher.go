package fetcher

import "github.com/RadioCheckerApp/api/model"

type Fetcher interface {
	Next() ([]*model.TrackRecord, error)
}
