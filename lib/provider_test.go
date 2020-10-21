package lib

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"

	"github.com/99designs/keyring"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/segmentio/aws-okta/sessioncache"
	"github.com/stretchr/testify/assert"
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
	kr := keyring.NewArrayKeyring([]keyring.Item{})
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
		SessionCacheSingleItem: true,
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

var theDistantFuture = time.Date(3000, 0, 0, 0, 0, 0, 0, time.UTC)

func TestProviderRetrieve(t *testing.T) {
	kr := keyring.NewArrayKeyring([]keyring.Item{})
	skis := sessioncache.SingleKrItemStore{Keyring: kr}
	//kipss := sessioncache.KrItemPerSessionStore{Keyring: kr}

	profile := "test-provider-retrieve"
	aki := fmt.Sprintf("%d", rand.Intn(100000))
	sak := fmt.Sprintf("%d", rand.Intn(100000))
	sst := fmt.Sprintf("%d", rand.Intn(100000))
	sess := sessioncache.Session{
		Name: profile,
		Credentials: sts.Credentials{
			AccessKeyId:     &aki,
			SecretAccessKey: &sak,
			SessionToken:    &sst,
			Expiration:      &theDistantFuture,
		},
	}
	po := ProviderOptions{
		SessionDuration:    DefaultSessionDuration,
		AssumeRoleDuration: DefaultAssumeRoleDuration,
		ExpiryWindow:       time.Minute * 5,
		Profiles: Profiles{
			profile: map[string]string{
				"aws_saml_url":    "home/amazon_aws/SAML/272",
				"role_arn":        "arn:aws:iam::<account-id>:role/<okta-role-name>",
				"session_ttl":     "12h",
				"assume_role_ttl": "12h",
			},
		},
		MFAConfig:              MFAConfig{},
		AssumeRoleArn:          "",
		SessionCacheSingleItem: true,
	}

	p, err := NewProvider(kr, profile, po)
	if err != nil {
		t.Fatal(err)
	}

	key := sessioncache.KeyWithProfileARN{
		ProfileName: profile,
		ProfileConf: po.Profiles[profile],
		Duration:    p.SessionDuration,
		ProfileARN:  p.AssumeRoleArn,
	}

	err = skis.Put(&key, &sess)
	if err != nil {
		t.Fatalf("error initializing store on Put %s", err)
	}

	cv, err := p.Retrieve()
	assert.NoError(t, err, "unexpected error on retrieve")
	assert.Equal(t, aki, cv.AccessKeyID)
	assert.Equal(t, sak, cv.SecretAccessKey)
	assert.Equal(t, sst, cv.SessionToken)
	assert.Equal(t, "okta", cv.ProviderName)
}

func TestProviderGetOps(t *testing.T) {
	kr := keyring.NewArrayKeyring([]keyring.Item{})
	skis := sessioncache.SingleKrItemStore{Keyring: kr}
	//kipss := sessioncache.KrItemPerSessionStore{Keyring: kr}

	profile := "test-provider-retrieve"
	aki := genRandNumStr(12)
	sak := genRandStr(20)
	sst := genRandStr(32)
	sess := sessioncache.Session{
		Name: profile,
		Credentials: sts.Credentials{
			AccessKeyId:     &aki,
			SecretAccessKey: &sak,
			SessionToken:    &sst,
			Expiration:      &theDistantFuture,
		},
	}
	po := ProviderOptions{
		SessionDuration:    DefaultSessionDuration,
		AssumeRoleDuration: DefaultAssumeRoleDuration,
		ExpiryWindow:       time.Minute * 5,
		Profiles: Profiles{
			profile: map[string]string{
				"aws_saml_url":    "home/amazon_aws/SAML/272",
				"role_arn":        "arn:aws:iam::<account-id>:role/<okta-role-name>",
				"session_ttl":     "12h",
				"assume_role_ttl": "12h",
			},
		},
		MFAConfig:              MFAConfig{},
		AssumeRoleArn:          "",
		SessionCacheSingleItem: true,
	}
	p, err := NewProvider(kr, profile, po)
	assert.NoError(t, err)

	key := sessioncache.KeyWithProfileARN{
		ProfileName: profile,
		ProfileConf: po.Profiles[profile],
		Duration:    p.SessionDuration,
		ProfileARN:  p.AssumeRoleArn,
	}

	err = skis.Put(&key, &sess)
	assert.NoError(t, err, "error init store on Put")

	_, err = p.Retrieve()
	assert.NoError(t, err, "unexpected error on retrieve")

	t.Run("GetExpiration", func(t *testing.T) {
		assert.Equal(t, p.GetExpiration(), theDistantFuture)
	})

	t.Run("getSamlURL", func(t *testing.T) {
		su, err := p.getSamlURL()
		assert.NoError(t, err)
		assert.Equal(t, su, "home/amazon_aws/SAML/272")
	})

	t.Run("getOktaSessionCookieKey", func(t *testing.T) {
		assert.Equal(t, p.getOktaSessionCookieKey(), "okta-session-cookie")
	})

	t.Run("getOktaAccountName", func(t *testing.T) {
		assert.Equal(t, p.getOktaAccountName(), "okta-creds")
	})
	/*
		// setup a test server to mock the actual http get
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, req.URL.String(), "/some/path")
			rw.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		org := fmt.Sprintf("%d", rand.Intn(100000))
		usr := fmt.Sprintf("%d", rand.Intn(100000))
		pwd := fmt.Sprintf("%d", rand.Intn(100000))
		oc := OktaCreds{
			Organization: org,
			Username:     usr,
			Password:     pwd,
			Domain:       server.URL,
		}

		err = AddCredsToKeyring("okta-creds", kr, &oc, MFAConfig{})
		if err != nil {
			t.Fatal(err)
		}
			t.Run("getSamlSessionCreds", func(t *testing.T) {
				cv, err := p.getSamlSessionCreds()
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, aki, cv.AccessKeyId)
				assert.Equal(t, sak, cv.SecretAccessKey)
				assert.Equal(t, sst, cv.SessionToken)
				assert.Equal(t, theDistantFuture, cv.Expiration)
			})
	*/
}
