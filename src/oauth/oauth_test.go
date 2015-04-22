package oauth

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

const (
	requestTokenURLEndpoint = "/request_token"
	authURLEndpoint         = "/authorize"
	tokenURLEndpoint        = "/token"
	redirectURLEndpoint     = "/redirect"
)

type OAuthTestSuite struct {
	suite.Suite
}

func Test_OAuthTestSuite(t *testing.T) {
	suite.Run(t, new(OAuthTestSuite))
}

func (suite *OAuthTestSuite) SetupTest() {
}

func (suite *OAuthTestSuite) TearDownTest() {
}
