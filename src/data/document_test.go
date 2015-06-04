package data

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type DocumentMapTestSuite struct {
	suite.Suite
}

func Test_DocumentMapTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentMapTestSuite))
}

func (suite *DocumentMapTestSuite) Test_MakeDocumentMapFromDocuments_WithUniqueDocuments() {
	javadoc := Document{
		Type:     "javadoc",
		CommitID: "javadoc-commit-id",
	}
	godoc := Document{
		Type:     "godoc",
		CommitID: "godoc-commit-id",
	}
	documents := []Document{javadoc, godoc}
	expected := DocumentMap{
		javadoc.Type: javadoc.CommitID,
		godoc.Type:   godoc.CommitID,
	}
	actual := MakeDocumentMapFromDocuments(documents)
	suite.Assert().Equal(expected, actual, "new map having the same as documents specified should be returned")
}

func (suite *DocumentMapTestSuite) Test_MakeDocumentMapFromDocuments_WithRedundantDocuments() {
	ignoredJavadoc := Document{
		Type:     "javadoc",
		CommitID: "javadoc-commit-id-ignored",
	}
	ignoredGodoc := Document{
		Type:     "godoc",
		CommitID: "godoc-commit-id-ignored",
	}
	accepteddJavadoc := Document{
		Type:     ignoredJavadoc.Type,
		CommitID: "javadoc-commit-id-accepted",
	}
	acceptedGodoc := Document{
		Type:     ignoredGodoc.Type,
		CommitID: "godoc-commit-id-accepted",
	}
	documents := []Document{ignoredJavadoc, ignoredGodoc, accepteddJavadoc, acceptedGodoc}
	expected := DocumentMap{
		accepteddJavadoc.Type: accepteddJavadoc.CommitID,
		acceptedGodoc.Type:    acceptedGodoc.CommitID,
	}
	actual := MakeDocumentMapFromDocuments(documents)
	suite.Assert().Equal(expected, actual, "new map having only documents at the largest indexes should be returned")
}

func (suite *DocumentMapTestSuite) Test_MakeDocumentMapFromDocuments_WithEmptyDocuments() {
	actual := MakeDocumentMapFromDocuments([]Document{})
	suite.Assert().Len(actual, 0, "zero-length map should be returned")
}

func (suite *DocumentMapTestSuite) Test_MakeDocumentMapFromDocuments_WithNil() {
	actual := MakeDocumentMapFromDocuments(nil)
	suite.Assert().Len(actual, 0, "zero-length map should be returned")
}

func (suite *DocumentMapTestSuite) Test_ListDocuments_WithDocuments() {
	docMap := DocumentMap{
		"javadoc": "javadoc-commit-id",
		"godoc":   "godoc-commit-id",
		"":        "",
	}
	actual := docMap.ListDocuments()
	suite.Assert().Len(actual, len(docMap), "slice having the same length as the map should be returned")
	for _, actualDoc := range actual {
		suite.Assert().Equal(docMap[actualDoc.Type], actualDoc.CommitID, "document information being held by the map should be returned")
	}
}

func (suite *DocumentMapTestSuite) Test_ListDocuments_WhenEmpty() {
	docMap := make(DocumentMap)
	suite.Assert().Len(docMap.ListDocuments(), 0, "zero-length slice should be returned")
}

func (suite *DocumentMapTestSuite) Test_ListDocuments_WhenNil() {
	var docMap DocumentMap
	suite.Assert().Len(docMap.ListDocuments(), 0, "zero-length slice should be returned")
}
