package lib

import (
	"os"

	"github.com/99designs/keyring"
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
