package gcp

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type RepositoryManagerClientTestSuite struct {
	suite.Suite
}

func Test_RepositoryManagerClientTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryManagerClientTestSuite))
}

func (suite *RepositoryManagerClientTestSuite) Test_NewRepositoryManagerClient_WithAppEngineContext() {
	suite.T().Skip("TODO: implement new repository manager client test with appengine context")
}

// TODO: Add more test cases for RepositoryManagerClient.
//       As of now, aetest package has not been ported from classic SDK to Go runtime for Managed VMs on App Engine.
//       See below about the Go runtime for Managed VMs on App Engine:
//       https://github.com/golang/appengine
