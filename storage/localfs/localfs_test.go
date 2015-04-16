package localfs

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
)

type LocalFSTestSuite struct {
	suite.Suite

	tempDir string
}

func (suite *LocalFSTestSuite) SetupTest() {
	dir, err := ioutil.TempDir("", "local_fs_test")
	suite.Assert().NoError(err, "Temporary directory needs to be created before test")
	suite.tempDir = dir
}

func (suite *LocalFSTestSuite) TearDownTest() {
	os.RemoveAll(suite.tempDir)
}

func Test_LocalFSTestSuite(t *testing.T) {
	suite.Run(t, new(LocalFSTestSuite))
}

func (suite *LocalFSTestSuite) Test_CreateOpenDelete_Valid() {
	accessor := NewLocalFS(suite.tempDir)

	testFileName := "localfstest.txt"
	testFileBody := "this is a test"
	// Create
	wc, err := accessor.Create(testFileName)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(wc)

	// Write & Close
	n, err := wc.Write([]byte(testFileBody))
	suite.Assert().NoError(err)
	suite.Assert().Equal(len(testFileBody), n)
	err = wc.Close()
	suite.Assert().NoError(err)

	// Open
	rc, err := accessor.Open(testFileName)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(rc)

	// Read & Close
	actualBody, err := ioutil.ReadAll(rc)
	suite.Assert().NoError(err)
	suite.Assert().Equal(len(testFileBody), n)
	suite.Assert().Equal(testFileBody, string(actualBody))
	err = rc.Close()
	suite.Assert().NoError(err)

	// Delete
	err = accessor.Delete(testFileName)
	suite.Assert().NoError(err)
}
