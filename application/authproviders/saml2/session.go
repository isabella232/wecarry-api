package saml2provider

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/markbates/goth"
)

// Session stores data during the auth process with Facebook.
type Session struct {
	AuthURL     string
	AccessToken string
	ExpiresAt   time.Time
}

// GetAuthURL will return the URL set by calling the `BeginAuth` function on the Facebook provider.
func (s Session) GetAuthURL() (string, error) {
	if s.AuthURL == "" {
		return "", errors.New(goth.NoAuthUrlErrorMessage)
	}
	return s.AuthURL, nil
}

// Authorize the session with Facebook and return the access token to be stored for future use.
func (s *Session) Authorize(provider goth.Provider, params goth.Params) (string, error) {

	return "??", nil
}

// Marshal the session into a string
func (s Session) Marshal() string {
	b, _ := json.Marshal(s)
	return string(b)
}

func (s Session) String() string {
	return s.Marshal()
}
