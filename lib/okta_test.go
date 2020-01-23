package lib

import (
	"testing"
)

type tgodCase struct {
	region      string
	expected    string
	shouldError bool
}

func TestGetOktaDomain(t *testing.T) {
	testCases := []tgodCase{
		tgodCase{"us", OktaServerUs, false},
		tgodCase{"emea", OktaServerEmea, false},
		tgodCase{"preview", OktaServerPreview, false},
		tgodCase{"invalid", "", true},
	}
	for _, testCase := range testCases {
		t.Run("get okta domain", func(t *testing.T) {
			got, err := GetOktaDomain(testCase.region)
			if err != nil && !testCase.shouldError {
				t.Error(err)
			}
			if err == nil && testCase.shouldError {
				t.Errorf("was expecting an error for %s", testCase.region)
			}
			if got != testCase.expected {
				t.Errorf("unexpected failure wanted %s got %s", testCase.expected, got)
			}
		})
	}
}
