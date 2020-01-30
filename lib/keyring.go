package lib

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/99designs/keyring"
	log "github.com/sirupsen/logrus"
)

func keyringPrompt(prompt string) (string, error) {
	return PromptWithOutput(prompt, true, os.Stderr)
}

// OpenKeyring compose the keyring.Config and handoff to the keyring library
// to get a functional keyring interaction object
func OpenKeyring(allowedBackends []keyring.BackendType) (kr keyring.Keyring, err error) {
	kr, err = keyring.Open(keyring.Config{
		AllowedBackends:          allowedBackends,
		KeychainTrustApplication: true,
		// this keychain name is for backwards compatibility
		ServiceName:             "aws-okta-login",
		LibSecretCollectionName: "awsvault",
		FileDir:                 "~/.aws-okta/",
		FilePasswordFunc:        keyringPrompt,
	})

	return
}

// AddCredsToKeyring store credentials into OS keyring for later use
func AddCredsToKeyring(key string, kr keyring.Keyring, creds *OktaCreds, mfaConfig MFAConfig) error {

	if err := creds.Validate(mfaConfig); err != nil {
		log.Debugf("Failed to validate credentials: %s", err)
		return fmt.Errorf("Failed to validate credentials %s", err)
	}

	encoded, err := json.Marshal(creds)
	if err != nil {
		return err
	}

	item := keyring.Item{
		Key:                         key,
		Data:                        encoded,
		Label:                       "okta credentials",
		KeychainNotTrustApplication: false,
	}

	if err := kr.Set(item); err != nil {
		log.Debugf("Failed to add user to keyring: %s", err)
		return errors.New("Failed to add user to keyring")
	}

	log.Infof("Added credentials for user %s", creds.Username)
	return nil

}
