package crawler

import (
	"reflect"
	"testing"
)

func TestGetDataFromResponseBody(t *testing.T) {
	var tests = []struct {
		json           []byte
		expectedResult interface{}
		expectedErr    bool
	}{
		{
			[]byte("{\"success\": true,\"data\": \"track created: /stations/hitradio-oe3/tracks/1534966200\"}"),
			"track created: /stations/hitradio-oe3/tracks/1534966200",
			false,
		},
		{
			[]byte("invalid json"),
			nil,
			true,
		},
		{
			[]byte("{\"status\": true,\"data\": \"track created: /stations/hitradio-oe3/tracks/1534966200\"}"),
			nil,
			true,
		},
		{
			[]byte("{\"success\": false,\"message\": \"API Error\"}"),
			nil,
			true,
		},
	}

	for _, test := range tests {
		data, err := getDataFromResponseBody(test.json)
		if (err != nil) != test.expectedErr {
			t.Errorf("GetDataFromResponseBody(%v): got error: (%q), expected error: (%v)",
				test.json, err, test.expectedErr)
		}

		if err != nil {
			continue
		}

		if !reflect.DeepEqual(data, test.expectedResult) {
			t.Errorf("GetDataFromResponseBody(%v): got data: (%v), expected data: (%v)",
				test.json, data, test.expectedResult)
		}
	}
}
