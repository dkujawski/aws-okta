package lib

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/segmentio/aws-okta/sessioncache"
)

type testCasePOV struct {
	name                    string
	givenSessionDuration    time.Duration
	givenAssumeRoleDuration time.Duration
	shouldError             bool
}

func TestProviderOptionsValidate(t *testing.T) {
	po := ProviderOptions{
		SessionDuration:        DefaultSessionDuration,
		AssumeRoleDuration:     DefaultAssumeRoleDuration,
		ExpiryWindow:           time.Minute * 5,
		Profiles:               Profiles{},
		MFAConfig:              MFAConfig{},
		AssumeRoleArn:          "",
		SessionCacheSingleItem: false,
	}

	dsd := DefaultSessionDuration
	mxsd := MaxSessionDuration
	mnsd := MinSessionDuration
	dard := DefaultAssumeRoleDuration
	mxard := MaxAssumeRoleDuration
	mnard := MinAssumeRoleDuration

	testCases := []testCasePOV{
		testCasePOV{"defaults", dsd, dard, false},
		testCasePOV{"max values", mxsd, mxard, false},
		testCasePOV{"min values", mnsd, mnard, false},
		testCasePOV{"over max session, max assume role", mxsd + (time.Minute * 5), mxard, true},
		testCasePOV{"over max values", mxsd + (time.Minute * 5), mxard + (time.Minute * 5), true},
		testCasePOV{"max session, over max assume role", mxsd, mxard + (time.Minute * 5), true},
		testCasePOV{"less min session, min assume role", mnsd - (time.Minute * 5), mnard, true},
		testCasePOV{"less min values", mnsd - (time.Minute * 5), mnard - (time.Minute * 5), true},
		testCasePOV{"min session, less min assume role", mnsd, mnard - (time.Minute * 5), true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			po.SessionDuration = tc.givenSessionDuration
			po.AssumeRoleDuration = tc.givenAssumeRoleDuration
			err := po.Validate()
			gotError := err != nil
			if gotError && !tc.shouldError {
				t.Errorf("unexpected (%s) gotError %t shouldError %t, session: %s assumeRole: %s",
					err, gotError, tc.shouldError,
					tc.givenSessionDuration.String(),
					tc.givenAssumeRoleDuration.String())
			}
			if !gotError && tc.shouldError {
				t.Errorf("should error but did not, gotError %t shouldError %t, session: %s assumeRole: %s",
					gotError, tc.shouldError,
					tc.givenSessionDuration.String(),
					tc.givenAssumeRoleDuration.String())
			}
		})
	}

	t.Run("Provider ApplyDefaults", func(t *testing.T) {
		po.SessionDuration = 0
		po.AssumeRoleDuration = 0
		po2 := po.ApplyDefaults()
		err := po2.Validate()
		if err != nil {
			t.Errorf("unexpected (%s) session: %s assumeRole: %s",
				err,
				po2.SessionDuration.String(),
				po2.AssumeRoleDuration.String())
		}
		if po2.SessionDuration != DefaultSessionDuration {
			t.Errorf("unexpected non-default value for SessionDuration %s", po2.SessionDuration)
		}
		if po2.AssumeRoleDuration != DefaultAssumeRoleDuration {
			t.Errorf("unexpected non-default value for AssumeRoleDurations %s", po2.AssumeRoleDuration)
		}
	})
}

func TestNewProvider(t *testing.T) {
	kr, err := keyring.Open(keyring.Config{
		AllowedBackends:          nil,
		KeychainTrustApplication: true,
		// this keychain name is for backwards compatibility
		ServiceName:             "aws-okta-login",
		LibSecretCollectionName: "awsvault",
		FileDir:                 "~/.aws-okta/",
		FilePasswordFunc:        keyringPrompt,
	})
	if err != nil {
		t.Error(err)
	}
	skis := reflect.TypeOf(&sessioncache.SingleKrItemStore{Keyring: kr})
	kipss := reflect.TypeOf(&sessioncache.KrItemPerSessionStore{Keyring: kr})

	profile := "test-profile"

	po := ProviderOptions{
		SessionDuration:        DefaultSessionDuration,
		AssumeRoleDuration:     DefaultAssumeRoleDuration,
		ExpiryWindow:           time.Minute * 5,
		Profiles:               Profiles{},
		MFAConfig:              MFAConfig{},
		AssumeRoleArn:          "",
		SessionCacheSingleItem: false,
	}
	for _, si := range []bool{true, false} {
		t.Run(fmt.Sprintf("NewProvider with SingleKrItemStore: %t", si), func(t *testing.T) {
			po.SessionCacheSingleItem = si
			p, err := NewProvider(kr, profile, po)
			if err != nil {
				t.Error(err)
			}
			got := reflect.TypeOf(p.sessions)
			if si {
				if got != skis {
					t.Errorf("unexpected sessioncache results wanted %s got %s", skis, got)
				}
			} else {
				if got != kipss {
					t.Errorf("unexpected sessioncache results wanted %s got %s", skis, got)
				}
			}
		})
	}
}
