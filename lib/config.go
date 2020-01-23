package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/mitchellh/go-homedir"
	"github.com/vaughan0/go-ini"
)

const envKeyAWSConfigFile = "AWS_CONFIG_FILE"
const baseNameAWSConfigFile = "/.aws/config"

// Profiles container to store found/existing configuration profiles
type Profiles map[string]map[string]string

type config interface {
	Parse() (Profiles, error)
}

// FileConfig container around the aws config file
type FileConfig struct {
	file string
}

// NewConfigFromEnv initialize a FileConfig struct by collect the file path from environment or use ~/.aws/config.
func NewConfigFromEnv() (*FileConfig, error) {
	file := os.Getenv(envKeyAWSConfigFile)
	if file == "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, err
		}
		file = filepath.Join(home, baseNameAWSConfigFile)
	}
	return &FileConfig{file: file}, nil
}

// Parse load and read the config file, return the profiles found
func (c *FileConfig) Parse() (Profiles, error) {
	if _, err := os.Stat(c.file); os.IsNotExist(err) {
		return nil, err
	}

	log.Debugf("Parsing config file %s", c.file)
	f, err := ini.LoadFile(c.file)
	if err != nil {
		return nil, fmt.Errorf("Error parsing config file %q: %v", c.file, err)
	}

	profiles := Profiles{"okta": map[string]string{}}
	for sectionName, section := range f {
		profiles[strings.TrimPrefix(sectionName, "profile ")] = section
	}

	return profiles, nil
}

// sourceProfile returns either the defined source_profile or p if none exists
func sourceProfile(p string, from Profiles) string {
	if conf, ok := from[p]; ok {
		if source := conf["source_profile"]; source != "" {
			return source
		}
	}
	return p
}

// GetValue return the value found in the config file for the given profile
func (p Profiles) GetValue(profile string, configKey string) (string, string, error) {
	configValue, ok := p[profile][configKey]
	if ok {
		return configValue, profile, nil
	}

	// Lookup from the `source_profile`, if it exists
	profile, ok = p[profile]["source_profile"]
	if ok {
		configValue, ok := p[profile][configKey]
		if ok {
			return configValue, profile, nil
		}

	}

	// Fallback to `okta` if no profile supplies the value
	profile = "okta"
	configValue, ok = p[profile][configKey]
	if ok {
		return configValue, profile, nil
	}

	return "", "", fmt.Errorf("Could not find %s in %s, source profile, or okta", configKey, profile)
}
