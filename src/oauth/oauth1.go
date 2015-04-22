package oauth

import (
	"bytes"
	"encoding/gob"
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/garyburd/go-oauth/oauth"
	"golang.org/x/net/context"

	"github.com/tractrix/common-go/gcp"
)

// OAuth1Verifier holds temporary credential info retrieved from OAuth1.0 provider.
type OAuth1Verifier struct {
	*oauth.Credentials
	Options map[string]interface{}
}

// GetOption gets optional value stored with verifier
func (ver *OAuth1Verifier) GetOption(key string) interface{} {
	return ver.Options[key]
}

// SetOption sts optional value to be stored with verifier
func (ver *OAuth1Verifier) SetOption(key string, val interface{}) {
	if ver.Options == nil {
		ver.Options = make(map[string]interface{})
	}
	ver.Options[key] = val
}

// OAuth1Token contains credential information to access resources through OAuth1.0 protocol
type OAuth1Token struct {
	*oauth.Credentials
}

func (token *OAuth1Token) version() string {
	return "1.0a"
}

// OAuth1Config contains oauth appication information
type OAuth1Config struct {
	ServiceName string

	C           oauth.Client
	RedirectURL string
}

// Client returns *http.Client to be used for OAuth1.0 request
func (conf *OAuth1Config) Client(ctx context.Context, _ Token) (*http.Client, error) {
	if ctx == nil {
		return nil, errors.New("invalid context")
	}

	return gcp.NewHTTPClient(ctx), nil
}

// NewVerifier creates verufuer by requesting temporary credential to OAuth provider
func (conf *OAuth1Config) NewVerifier(ctx context.Context) (Verifier, string, error) {
	client, _ := conf.Client(ctx, nil)
	tempCredential, err := conf.C.RequestTemporaryCredentials(client, conf.RedirectURL, nil)
	if err != nil {
		log.Printf("error obtaining temporary credentials: %v", err)
		return nil, "", err
	}

	return &OAuth1Verifier{Credentials: tempCredential}, tempCredential.Token, nil
}

// GetVerifier gets verifier given key
func (conf *OAuth1Config) GetVerifier(ctx context.Context, key string) (Verifier, error) {
	val, err := getVerifier(ctx, conf.ServiceName, key)
	if err != nil {
		return nil, err
	}

	var ver OAuth1Verifier
	dec := gob.NewDecoder(bytes.NewReader(val))
	if err := dec.Decode(&ver); err != nil {
		log.Printf("decode error: %+v", err)
		return nil, err
	}

	return &ver, nil
}

// PutVerifier stores verifier for later use
func (conf *OAuth1Config) PutVerifier(ctx context.Context, key string, ver Verifier) error {
	return putVerifier(ctx, conf.ServiceName, key, ver)
}

// DeleteVerifier deletes verifier from storage
func (conf *OAuth1Config) DeleteVerifier(ctx context.Context, key string) error {
	return deleteVerifier(ctx, conf.ServiceName, key)
}

// GetVerifierKey obtains verifier key from signin callback
func (conf *OAuth1Config) GetVerifierKey(req *http.Request) (string, error) {
	authToken := req.FormValue("oauth_token")
	if authToken == "" {
		log.Printf("no oauth token")
		return "", errors.New("no oauth token")
	}

	return authToken, nil
}

// Exchange exchanges temporary token with proper access token
func (conf *OAuth1Config) Exchange(ctx context.Context, ver Verifier, veriCode string) (Token, error) {
	tempCred, ok := ver.(*OAuth1Verifier)
	if !ok {
		return nil, errors.New("invalid verifier")
	}

	client, _ := conf.Client(ctx, nil)
	tokenCredential, _, err := conf.C.RequestToken(client, tempCred.Credentials, veriCode)
	if err != nil {
		log.Printf("error retrieving request token: %v", err)
		return nil, err
	}

	return &OAuth1Token{Credentials: tokenCredential}, nil
}

// GetExchangeCode obtains access token from signin callback
func (conf *OAuth1Config) GetExchangeCode(req *http.Request) (string, error) {
	veriCode := req.FormValue("oauth_verifier")
	if veriCode == "" {
		log.Printf("no oauth verifier")
		return "", errors.New("no oauth verifier")
	}

	return veriCode, nil
}

// LoginURL returns url for oauth login
func (conf *OAuth1Config) LoginURL(ctx context.Context, ver Verifier) (*url.URL, error) {
	cred, ok := ver.(*OAuth1Verifier)
	if !ok {
		return nil, errors.New("invalid verifier")
	}

	loginURL, err := url.Parse(conf.C.AuthorizationURL(cred.Credentials, nil))
	if err != nil {
		log.Printf("error obtaining login url: %v", err)
		return nil, err
	}

	return loginURL, nil
}
