package oauth

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

const (
	verifierLength = 32 // Length of random string value for state parameter
)

// OAuth2Verifier contains OAuth2.0 state parameter
type OAuth2Verifier struct {
	state   string
	Options map[string]interface{}
}

// GetOption gets optional value stored with verifier
func (ver *OAuth2Verifier) GetOption(key string) interface{} {
	return ver.Options[key]
}

// SetOption sets optional value to be stored with verifier
func (ver *OAuth2Verifier) SetOption(key string, val interface{}) {
	if ver.Options == nil {
		ver.Options = make(map[string]interface{})
	}
	ver.Options[key] = val
}

// OAuth2Token contains credential information to access resources through OAuth2.0 protocol
type OAuth2Token struct {
	oauth2.Token
}

func (token *OAuth2Token) version() string {
	return "2.0"
}

// OAuth2Config contains oauth appication information
type OAuth2Config struct {
	ServiceName string

	oauth2.Config
}

// Client returns *http.Client that automatically appends authorization headers to requests
// Access token is automaticlally refreshed when it is expired
func (conf *OAuth2Config) Client(ctx context.Context, tok Token) (*http.Client, error) {
	if ctx == nil {
		return nil, errors.New("invalid context")
	}

	t, ok := tok.(*OAuth2Token)
	if !ok {
		return nil, errors.New("invalid token")
	}

	return conf.Config.Client(ctx, &t.Token), nil
}

// NewVerifier creats random string value to be passed as state parameter
func (conf *OAuth2Config) NewVerifier(_ context.Context) (Verifier, string, error) {
	str := randString(verifierLength)
	return &OAuth2Verifier{state: str}, str, nil
}

// GetVerifier gets verifier given key
func (conf *OAuth2Config) GetVerifier(ctx context.Context, key string) (Verifier, error) {
	val, err := getVerifier(ctx, conf.ServiceName, key)
	if err != nil {
		return nil, err
	}

	var ver OAuth2Verifier
	dec := gob.NewDecoder(bytes.NewReader(val))
	if err := dec.Decode(&ver); err != nil {
		log.Printf("decode error: %+v", err)
		return nil, err
	}

	return &ver, nil
}

// PutVerifier stores verifier with state as a key
func (conf *OAuth2Config) PutVerifier(ctx context.Context, key string, ver Verifier) error {
	return putVerifier(ctx, conf.ServiceName, key, ver)
}

// DeleteVerifier deletes verifier from storage
func (conf *OAuth2Config) DeleteVerifier(ctx context.Context, key string) error {
	return deleteVerifier(ctx, conf.ServiceName, key)
}

// GetVerifierKey obtains verifier key from signin callback
func (conf *OAuth2Config) GetVerifierKey(req *http.Request) (string, error) {
	state := req.FormValue("state")
	if state == "" {
		log.Printf("no state")
		return "", errors.New("no state")
	}

	return state, nil
}

// Exchange exchanges authorization code with access token
func (conf *OAuth2Config) Exchange(ctx context.Context, ver Verifier, authCode string) (Token, error) {
	t, err := conf.Config.Exchange(ctx, authCode)
	if err != nil || t.AccessToken == "" {
		return nil, err
	}

	return &OAuth2Token{Token: *t}, nil
}

// GetExchangeCode obtains access token from signin callback
func (conf *OAuth2Config) GetExchangeCode(req *http.Request) (string, error) {
	authCode := req.FormValue("code")
	if authCode == "" {
		log.Printf("no auth code")
		return "", errors.New("no auth code")
	}

	return authCode, nil
}

// LoginURL returns url for oauth login
func (conf *OAuth2Config) LoginURL(ctx context.Context, ver Verifier) (*url.URL, error) {
	ver2, ok := ver.(*OAuth2Verifier)
	if !ok {
		return nil, errors.New("invalid verifier")
	}
	loginURL, err := url.Parse(conf.Endpoint.AuthURL)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("respons_type", "code")
	params.Set("scope", strings.Join(conf.Scopes, ","))
	params.Set("client_id", conf.ClientID)
	params.Set("state", ver2.state)
	params.Set("redirect_uri", conf.RedirectURL)

	loginURL.RawQuery = params.Encode()

	return loginURL, nil
}
