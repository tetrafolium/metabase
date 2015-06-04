package github

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/suite"
	"golang.org/x/oauth2"

	"github.com/tractrix/common-go/oauth"
	"github.com/tractrix/common-go/repository"
)

type ServiceTestSuiteBase struct {
	suite.Suite

	mux     *http.ServeMux
	server  *httptest.Server
	service *Service
}

func (suite *ServiceTestSuiteBase) SetupTest() {
	OAuthConfig = &oauth.OAuth2Config{}

	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)

	suite.service, _ = newServiceWithClient(nil)
	url, _ := url.Parse(suite.server.URL)
	suite.service.client.BaseURL = url
	suite.service.client.UploadURL = url
}

func (suite *ServiceTestSuiteBase) TearDownTest() {
	suite.server.Close()

	suite.service = nil
	suite.server = nil
	suite.mux = nil
}

type GithubServiceTestSuite struct {
	ServiceTestSuiteBase
}

func Test_GithubServiceTestSuite(t *testing.T) {
	suite.Run(t, new(GithubServiceTestSuite))
}

func (suite *GithubServiceTestSuite) Test_NewGithubService_WithContextAndToken() {
	actual, err := NewService(oauth2.NoContext, &oauth.OAuth2Token{})
	suite.Assert().NoError(err)
	suite.Assert().IsType(new(Service), actual, "github.Service object should be created")
	suite.Assert().Implements((*repository.Service)(nil), actual, "github.Service should implement repository.Service")
}

func (suite *GithubServiceTestSuite) Test_translateErrorResponse_WithStatusNotFoundResponse() {
	resp := &github.Response{
		Response: &http.Response{
			StatusCode: http.StatusNotFound,
		},
	}
	actual := suite.service.translateErrorResponse(resp, errors.New("404 Not Found"))
	suite.Assert().Equal(repository.ErrNotFound, actual)
}

func (suite *GithubServiceTestSuite) Test_translateErrorResponse_WithUnknownResponse() {
	resp := &github.Response{}
	expected := errors.New("unknown error")
	actual := suite.service.translateErrorResponse(resp, expected)
	suite.Assert().Equal(expected, actual)
}

func (suite *GithubServiceTestSuite) Test_translateErrorResponse_WithNilResponse() {
	expected := errors.New("nil error")
	actual := suite.service.translateErrorResponse(nil, expected)
	suite.Assert().Equal(expected, actual)
}

func (suite *GithubServiceTestSuite) Test_GetServiceName() {
	suite.Assert().Equal(ServiceName, suite.service.GetServiceName())
}
