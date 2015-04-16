package storage

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tractrix/common-go/test/mock"
)

type StorageTestSuite struct {
	suite.Suite

	ac Accessor
}

func (suite *StorageTestSuite) SetupTest() {
	suite.ac = new(mock.StorageAccessorMock)
}

func (suite *StorageTestSuite) TearDownTest() {
	suite.ac = nil
}

func Test_StorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func (suite *StorageTestSuite) Test_CreateFile_ValidAccessor() {
	testFileName := "testfileName.txt"
	testFileBody := []byte{}
	err := CreateFile(suite.ac, testFileName, testFileBody)
	suite.Assert().NoError(err)
}

func (suite *StorageTestSuite) Test_CreateFile_NilAccessor() {
	testFileName := "testfileName.txt"
	testFileBody := []byte{}
	err := CreateFile(nil, testFileName, testFileBody)
	suite.Assert().Error(err)
}

func (suite *StorageTestSuite) Test_ReadFile_ValidAccessor() {
	testFileName := "testfileName.txt"
	body, err := ReadFile(suite.ac, testFileName)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(body)
}

func (suite *StorageTestSuite) Test_ReadFile_NilAccessor() {
	testFileName := "testfileName.txt"
	body, err := ReadFile(nil, testFileName)
	suite.Assert().Error(err)
	suite.Assert().Nil(body)
}
