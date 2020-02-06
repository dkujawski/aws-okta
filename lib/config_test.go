package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/stretchr/testify/assert"
	"github.com/vaughan0/go-ini"
)

func TestNewConfigFromEnv(t *testing.T) {
	fakeEnvPath := "/fake/env/path/to/.file/config"
	err := os.Setenv(envKeyAWSConfigFile, fakeEnvPath)
	assert.NoError(t, err)
	t.Run("get aws path from env", func(t *testing.T) {
		got, err := NewConfigFromEnv()
		assert.NoError(t, err)
		assert.Equal(t, got.file, fakeEnvPath)
	})
	err = os.Setenv(envKeyAWSConfigFile, "")
	assert.NoError(t, err)

	home, err := homedir.Dir()
	assert.NoError(t, err)

	expectedPath := filepath.Join(home, baseNameAWSConfigFile)
	t.Run("get default config path for OS", func(t *testing.T) {
		got, err := NewConfigFromEnv()
		assert.NoError(t, err)
		assert.Equal(t, got.file, expectedPath)
	})
}

func TestParse(t *testing.T) {
	src := `
[profile nf-sandbox2]
aws_saml_url = home/amazon_aws/SAML/272
role_arn = arn:aws:iam::<account-id>:role/<okta-role-name>
assume_role_ttl = 12h
session_ttl = 12h

[default]
region = us-west-2
output = json	
	`
	f, err := ini.Load(strings.NewReader(src))
	assert.NoError(t, err)

	fc := FileConfig{
		file: "",
		fh:   &f,
	}
	expected := Profiles{
		"okta":    map[string]string{},
		"default": map[string]string{"output": "json", "region": "us-west-2"},
		"nf-sandbox2": map[string]string{
			"assume_role_ttl": "12h",
			"aws_saml_url":    "home/amazon_aws/SAML/272",
			"role_arn":        "arn:aws:iam::<account-id>:role/<okta-role-name>",
			"session_ttl":     "12h",
		},
	}
	t.Run("parse config", func(t *testing.T) {
		got, err := fc.Parse()
		assert.NoError(t, err)
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected failure wanted %#v got %#v", expected, got)
		}
	})

}

func TestSourceProfile(t *testing.T) {
	src := `
[profile nf-sandbox2]
aws_saml_url = home/amazon_aws/SAML/272
role_arn = arn:aws:iam::<account-id>:role/<okta-role-name>
assume_role_ttl = 12h
session_ttl = 12h

[profile nf-sandbox2-extra-role]
source_profile = nf-sandbox2
role_arn = arn:aws:iam::<account-id>:role/<okta-role-extra>
assume_role_ttl = 12h
session_ttl = 12h
	`
	f, err := ini.Load(strings.NewReader(src))
	assert.NoError(t, err)

	fc := FileConfig{
		file: "",
		fh:   &f,
	}
	p, err := fc.Parse()
	assert.NoError(t, err)

	testCases := map[string]string{
		"nf-sandbox2":            "nf-sandbox2",
		"nf-sandbox2-extra-role": "nf-sandbox2",
		"missing-key":            "missing-key",
	}
	for profile, expected := range testCases {
		t.Run(fmt.Sprintf("source profile %s", profile), func(t *testing.T) {
			got := sourceProfile(profile, p)
			assert.Equal(t, got, expected)
		})
	}
}

func TestGetConfigValue(t *testing.T) {
	configProfiles := make(Profiles)

	t.Run("empty profile", func(t *testing.T) {
		_, _, foundError := configProfiles.GetValue("profile_a", "config_key")
		assert.Error(t, foundError, "Searching an empty profile set should error")
	})

	configProfiles["okta"] = map[string]string{
		"key_a": "a",
		"key_b": "b",
	}

	configProfiles["profile_a"] = map[string]string{
		"key_b": "b-a",
		"key_c": "c-a",
		"key_d": "d-a",
	}

	configProfiles["profile_b"] = map[string]string{
		"source_profile": "profile_a",
		"key_d":          "d-b",
		"key_e":          "e-b",
	}

	configProfiles["profile_c"] = map[string]string{
		"source_profile": "profile_b",
		"key_f":          "f-c",
	}

	t.Run("missing key", func(t *testing.T) {
		_, _, foundError := configProfiles.GetValue("profile_a", "config_key")
		assert.Error(t, foundError, "Searching for a missing key should error")
	})

	t.Run("fallback to okta", func(t *testing.T) {
		foundValue, foundProfile, foundError := configProfiles.GetValue("profile_a", "key_a")
		assert.NoError(t, foundError)
		assert.Equal(t, foundProfile, "okta")
		assert.Equal(t, foundValue, "a")
	})

	t.Run("found in current profile", func(t *testing.T) {
		foundValue, foundProfile, foundError := configProfiles.GetValue("profile_b", "key_d")
		assert.NoError(t, foundError, "searching for key_d")
		assert.Equal(t, foundProfile, "profile_b")
		assert.Equal(t, foundValue, "d-b")
	})

	t.Run("traversing from child profile", func(t *testing.T) {
		foundValue, foundProfile, foundError := configProfiles.GetValue("profile_b", "key_a")
		assert.NoError(t, foundError, "searching for key_a")
		assert.Equal(t, foundProfile, "okta", "key_a should come from `okta`")
		assert.Equal(t, foundValue, "a", "`key_a` should b `a`")
	})

	t.Run("recursive traversing from child profile", func(t *testing.T) {
		_, _, foundError := configProfiles.GetValue("profile_c", "key_c")
		assert.Error(t, foundError, "Recursive searching should not work")
	})
}
