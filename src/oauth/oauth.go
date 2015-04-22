package oauth

import (
	"encoding/gob"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
)

const (
	verifierSeparator = "_"
	oauthPrefix       = "oauth"

	verifierExpiration = 20 * time.Minute
)

// Verifier describes parameters for CSRF provisioning
// It also holds optional parameters to be passed in signin flow
type Verifier interface {
	GetOption(key string) interface{}
	SetOption(key string, val interface{})
}

// Token represents OAuth1.0/2.0 token
type Token interface {
	version() string
}

// Config describes typical OAuth1.0/2.0 flow
type Config interface {
	Client(context.Context, Token) (*http.Client, error)

	NewVerifier(context.Context) (Verifier, string, error)
	GetVerifier(context.Context, string) (Verifier, error)
	PutVerifier(context.Context, string, Verifier) error
	DeleteVerifier(context.Context, string) error

	Exchange(context.Context, Verifier, string) (Token, error)
	LoginURL(context.Context, Verifier) (*url.URL, error)

	GetVerifierKey(*http.Request) (string, error)
	GetExchangeCode(*http.Request) (string, error)
}

func init() {
	gob.Register(OAuth1Verifier{})
	gob.Register(OAuth2Verifier{})
	gob.Register(&OAuth1Token{})
	gob.Register(&OAuth2Token{})

	// For generating state parameter
	rand.Seed(time.Now().Unix())
}
