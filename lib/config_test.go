package lib

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/mitchellh/go-homedir"
	"github.com/vaughan0/go-ini"
)

func TestNewConfigFromEnv(t *testing.T) {
	fakeEnvPath := "/fake/env/path/to/.file/config"
	if err := os.Setenv(envKeyAWSConfigFile, fakeEnvPath); err != nil {
		t.Error("Unable to set env test value")
	}
	t.Run("get aws path from env", func(t *testing.T) {
		got, err := NewConfigFromEnv()
		if err != nil {
			t.Error(err)
		}
		if got.file != fakeEnvPath {
			t.Errorf("unexpected error! wanted '%s' got '%s'", fakeEnvPath, got.file)
		}
	})
	if err := os.Setenv(envKeyAWSConfigFile, ""); err != nil {
		t.Error("Unable to set empty env value")
	}
	home, err := homedir.Dir()
	if err != nil {
		t.Error(err)
	}
	expectedPath := filepath.Join(home, baseNameAWSConfigFile)
	t.Run("get default config path for OS", func(t *testing.T) {
		got, err := NewConfigFromEnv()
		if err != nil {
			t.Error(err)
		}
		if got.file != expectedPath {
			t.Errorf("unexpected error! wanted '%s' got '%s'", expectedPath, got.file)
		}
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
	if err != nil {
		t.Error(err)
	}
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
		if err != nil {
			t.Error(err)
		}
		if !reflect.DeepEqual(got, expected) {
			t.Errorf("unexpected failure wanted %#v got %#v", expected, got)
		}
	})

}

func TestGetConfigValue(t *testing.T) {
	configProfiles := make(Profiles)

	t.Run("empty profile", func(t *testing.T) {
		_, _, foundError := configProfiles.GetValue("profile_a", "config_key")
		if foundError == nil {
			t.Error("Searching an empty profile set should return an error")
		}
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
		if foundError == nil {
			t.Error("Searching for a missing key should return an error")
		}
	})

	t.Run("fallback to okta", func(t *testing.T) {
		foundValue, foundProfile, foundError := configProfiles.GetValue("profile_a", "key_a")
		if foundError != nil {
			t.Error("Error when searching for key_a")
		}

		if foundProfile != "okta" {
			t.Error("key_a should have come from `okta`")
		}

		if foundValue != "a" {
			t.Error("The proper value for `key_a` should be `a`")
		}
	})

	t.Run("found in current profile", func(t *testing.T) {
		foundValue, foundProfile, foundError := configProfiles.GetValue("profile_b", "key_d")
		if foundError != nil {
			t.Error("Error when searching for key_d")
		}

		if foundProfile != "profile_b" {
			t.Error("key_d should have come from `profile_b`")
		}

		if foundValue != "d-b" {
			t.Error("The proper value for `key_d` should be `d-b`")
		}
	})

	t.Run("traversing from child profile", func(t *testing.T) {
		foundValue, foundProfile, foundError := configProfiles.GetValue("profile_b", "key_a")
		if foundError != nil {
			t.Error("Error when searching for key_a")
		}

		if foundProfile != "okta" {
			t.Error("key_a should have come from `okta`")
		}

		if foundValue != "a" {
			t.Error("The proper value for `key_a` should be `a`")
		}
	})

	t.Run("recursive traversing from child profile", func(t *testing.T) {
		_, _, foundError := configProfiles.GetValue("profile_c", "key_c")
		if foundError == nil {
			t.Error("Recursive searching should not work")
		}
	})
}
