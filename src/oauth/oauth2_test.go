package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"
)

type OAuth2TestSuite struct {
	suite.Suite

	c      *OAuth2Config
	mux    *http.ServeMux
	server *httptest.Server
}

func Test_OAuth2TestSuite(t *testing.T) {
	suite.Run(t, new(OAuth2TestSuite))
}

func (suite *OAuth2TestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.c = &OAuth2Config{
		ServiceName: "serviceName",
		Config: oauth2.Config{
			ClientID:     "testClientID",
			ClientSecret: "testClientSecret",
			RedirectURL:  suite.server.URL + redirectURLEndpoint,
			Scopes:       []string{"scope1", "scope2"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  suite.server.URL + authURLEndpoint,
				TokenURL: suite.server.URL + tokenURLEndpoint,
			},
		},
	}
}

func (suite *OAuth2TestSuite) TearDownTest() {
	suite.c = nil
}

func (suite *OAuth2TestSuite) Test_OAuth2Config_ClientWithValidParameters() {
	client, err := suite.c.Client(oauth2.NoContext, &OAuth2Token{})
	suite.Assert().NoError(err)
	suite.Assert().NotNil(client)
}

func (suite *OAuth2TestSuite) Test_OAuth2Config_ClientWithNilContext_Error() {
	client, err := suite.c.Client(nil, &OAuth2Token{})
	suite.Assert().Error(err)
	suite.Assert().Nil(client)
}

func (suite *OAuth2TestSuite) Test_OAuth2Config_ClientWithNilToken_Error() {
	client, err := suite.c.Client(oauth2.NoContext, nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(client)
}

func (suite *OAuth2TestSuite) Test_OAuth2Config_NewVerifier() {
	ver, key, err := suite.c.NewVerifier(oauth2.NoContext)

	suite.Assert().NoError(err)
	suite.Assert().NotEmpty(ver)
	suite.Assert().NotEmpty(key, "non empty key should be created")
	suite.Assert().IsType(&OAuth2Verifier{}, ver, "OAuth2Config should return oauth2 type verifier")
	suite.Assert().Equal(key, ver.(*OAuth2Verifier).state, "verifier value and key should be the same")
}

func (suite *OAuth2TestSuite) Test_OAuth2Config_LoginURL() {
	ver, _, err := suite.c.NewVerifier(oauth2.NoContext)
	suite.Assert().NoError(err)

	url, err := suite.c.LoginURL(oauth2.NoContext, ver)

	suite.Assert().NoError(err)
	suite.Assert().NotNil(url)
	suite.Assert().Equal(suite.c.Endpoint.AuthURL+"?"+url.RawQuery, url.String(), "request uri should be pointing to auth url")

	queries := url.Query()
	suite.Assert().Equal(suite.c.ClientID, queries.Get("client_id"))
	suite.Assert().Equal(suite.c.RedirectURL, queries.Get("redirect_uri"))
	suite.Assert().Equal(strings.Join(suite.c.Scopes, ","), queries.Get("scope"))
}
