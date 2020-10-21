package lib

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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
			if tc.shouldError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, got, tc.expected, "should be the same")
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
		Id:         genRandStr(12),
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
		if tc.shouldError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
		assert.Equal(t, got, tc.expected)
	}
}
