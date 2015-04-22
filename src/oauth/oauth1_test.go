package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"
)

type OAuth1TestSuite struct {
	suite.Suite

	c      *OAuth1Config
	mux    *http.ServeMux
	server *httptest.Server
}

func Test_OAuth1TestSuite(t *testing.T) {
	suite.Run(t, new(OAuth1TestSuite))
}

func (suite *OAuth1TestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.c = &OAuth1Config{
		ServiceName: "serviceName",
		C: oauth.Client{
			TemporaryCredentialRequestURI: suite.server.URL + requestTokenURLEndpoint,
			ResourceOwnerAuthorizationURI: suite.server.URL + authURLEndpoint,
			TokenRequestURI:               suite.server.URL + tokenURLEndpoint,
			Credentials: oauth.Credentials{
				Token:  "testClientID",
				Secret: "testClientSecret",
			},
		},
		RedirectURL: suite.server.URL + redirectURLEndpoint,
	}
}

func (suite *OAuth1TestSuite) TearDownTest() {
	suite.c = nil
}

func (suite *OAuth1TestSuite) Test_OAuth1Config_ClientWithValidParameters() {
	client, err := suite.c.Client(context.TODO(), &OAuth1Token{})
	suite.Assert().NoError(err)
	suite.Assert().NotNil(client)
}

func (suite *OAuth1TestSuite) Test_OAuth1Config_ClientWithNilContext_Error() {
	client, err := suite.c.Client(nil, &OAuth1Token{})
	suite.Assert().Error(err)
	suite.Assert().Nil(client)
}

func (suite *OAuth1TestSuite) Test_OAuth1Config_ClientWithNilToken_NoError() {
	client, err := suite.c.Client(context.TODO(), nil)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(client)
}
