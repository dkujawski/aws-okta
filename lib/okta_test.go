package lib

import (
	"fmt"
	"math/rand"
	"testing"
)

type testCase struct {
	given       string
	expected    string
	shouldError bool
}

func TestGetOktaDomain(t *testing.T) {
	testCases := []testCase{
		testCase{"us", OktaServerUs, false},
		testCase{"emea", OktaServerEmea, false},
		testCase{"preview", OktaServerPreview, false},
		testCase{"invalid", "", true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("get okta domain %s", tc.given), func(t *testing.T) {
			got, err := GetOktaDomain(tc.given)
			if err != nil && !tc.shouldError {
				t.Error(err)
			}
			if err == nil && tc.shouldError {
				t.Errorf("was expecting an error for %s", tc.given)
			}
			if got != tc.expected {
				t.Errorf("unexpected failure wanted %s got %s", tc.expected, got)
			}
		})
	}
}

type testCaseOUAF struct {
	givenType     string
	givenProvider string
	expected      string
	shouldError   bool
}

func TestGetFactorId(t *testing.T) {
	ouaf := OktaUserAuthnFactor{
		Id:         fmt.Sprintf("%d", rand.Intn(999)),
		FactorType: "",
		Provider:   "",
		Embedded:   OktaUserAuthnFactorEmbedded{},
		Profile:    OktaUserAuthnFactorProfile{},
	}
	testCases := []testCaseOUAF{
		testCaseOUAF{"web", "", ouaf.Id, false},
		testCaseOUAF{"token", "SYMANTEC", ouaf.Id, false},
		testCaseOUAF{"token", "unknown", "", true},
		testCaseOUAF{"token:software:totp", "", ouaf.Id, false},
		testCaseOUAF{"token:hardware", "", ouaf.Id, false},
		testCaseOUAF{"sms", "", ouaf.Id, false},
		testCaseOUAF{"u2f", "", ouaf.Id, false},
		testCaseOUAF{"webauthn", "", ouaf.Id, false},
		testCaseOUAF{"push", "OKTA", ouaf.Id, false},
		testCaseOUAF{"push", "DUO", ouaf.Id, false},
		testCaseOUAF{"push", "unknown", "", true},
		testCaseOUAF{"default", "", "", true},
	}
	for _, tc := range testCases {
		ouaf.FactorType = tc.givenType
		ouaf.Provider = tc.givenProvider
		got, err := GetFactorId(&ouaf)
		if err != nil && !tc.shouldError {
			t.Error(err)
		} else if err == nil && tc.shouldError {
			t.Errorf("expecting error for test %s %s", tc.givenType, tc.givenProvider)
		}
		if got != tc.expected {
			t.Errorf("unexpected error wanted %s got %s", tc.expected, got)
		}
	}
}
