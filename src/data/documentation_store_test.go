package data

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DocumentationStoreTestSuite struct {
	suite.Suite
}

func Test_DocumentationStoreTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentationStoreTestSuite))
}

func (suite *DocumentationStoreTestSuite) Test_StatusQueuing() {
	// Pay much attention when changing the value of DocumentationStatusQueuing,
	// hence this test case.
	suite.Assert().Equal("queuing", DocumentationStatusQueuing)
}

func (suite *DocumentationStoreTestSuite) Test_StatusQueued() {
	// Pay much attention when changing the value of DocumentationStatusQueued,
	// hence this test case.
	suite.Assert().Equal("queued", DocumentationStatusQueued)
}

func (suite *DocumentationStoreTestSuite) Test_StatusStarted() {
	// Pay much attention when changing the value of DocumentationStatusStarted,
	// hence this test case.
	suite.Assert().Equal("started", DocumentationStatusStarted)
}

func (suite *DocumentationStoreTestSuite) Test_StatusGenerated() {
	// Pay much attention when changing the value of DocumentationStatusSucceeded,
	// hence this test case.
	suite.Assert().Equal("succeeded", DocumentationStatusSucceeded)
}

func (suite *DocumentationStoreTestSuite) Test_StatusFailed() {
	// Pay much attention when changing the value of DocumentationStatusFailed,
	// hence this test case.
	suite.Assert().Equal("failed", DocumentationStatusFailed)
}

func (suite *DocumentationStoreTestSuite) Test_StatusCancelled() {
	// Pay much attention when changing the value of DocumentationStatusCancelled,
	// hence this test case.
	suite.Assert().Equal("cancelled", DocumentationStatusCancelled)
}
