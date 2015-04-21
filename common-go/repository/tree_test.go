package repository

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SlugTreeTestSuite struct {
	suite.Suite
}

func Test_SlugTreeTestSuite(t *testing.T) {
	suite.Run(t, new(SlugTreeTestSuite))
}

func (suite *SlugTreeTestSuite) Test_GetData_WithValidRepository() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRepoData := "repo data"
	tree := SlugTree{
		slug.Saas: map[string]map[string]interface{}{
			slug.Owner: map[string]interface{}{
				slug.Name: expectedRepoData,
			},
		},
	}
	actualRepoData := tree.GetData(slug)
	suite.Assert().Equal(expectedRepoData, actualRepoData, "appropriate repository data should be returned")
}

func (suite *SlugTreeTestSuite) Test_GetData_WithNilRepository() {
	tree := SlugTree{
		"github.com": map[string]map[string]interface{}{
			"tractrix": map[string]interface{}{
				"common-go": "repo data",
			},
		},
	}
	suite.Assert().Nil(tree.GetData(nil))
}

func (suite *SlugTreeTestSuite) Test_GetData_WithUnknownSaas() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	tree := SlugTree{
		slug.Saas + "-unknown": map[string]map[string]interface{}{
			slug.Owner: map[string]interface{}{
				slug.Name: "repo data",
			},
		},
	}
	suite.Assert().Nil(tree.GetData(slug))
}

func (suite *SlugTreeTestSuite) Test_GetData_WithUnknownOwner() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	tree := SlugTree{
		slug.Saas: map[string]map[string]interface{}{
			slug.Owner + "-unknown": map[string]interface{}{
				slug.Name: "repo data",
			},
		},
	}
	suite.Assert().Nil(tree.GetData(slug))
}

func (suite *SlugTreeTestSuite) Test_GetData_WithUnknownName() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	tree := SlugTree{
		slug.Saas: map[string]map[string]interface{}{
			slug.Owner: map[string]interface{}{
				slug.Name + "-unknown": "repo data",
			},
		},
	}
	suite.Assert().Nil(tree.GetData(slug))
}

func (suite *SlugTreeTestSuite) Test_GetData_WhenNil() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	var tree SlugTree
	suite.Assert().Nil(tree.GetData(slug))
}

func (suite *SlugTreeTestSuite) Test_PutData_WithNewRepository() {
	tree := make(SlugTree)
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRepoData := "repo data"
	tree.PutData(slug, expectedRepoData)
	actualRepoData := tree.GetData(slug)
	suite.Assert().Equal(expectedRepoData, actualRepoData, "new data should be stored and retrieved for a repository")
}

func (suite *SlugTreeTestSuite) Test_PutData_WithExistingRepository() {
	tree := make(SlugTree)
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRepoData := "repo data"
	tree.PutData(slug, expectedRepoData+"-old")
	tree.PutData(slug, expectedRepoData)
	actualRepoData := tree.GetData(slug)
	suite.Assert().Equal(expectedRepoData, actualRepoData, "data should be overwritten for a repository")
}

func (suite *SlugTreeTestSuite) Test_PutData_WittNilRepository() {
	suite.Assert().NotPanics(func() {
		tree := make(SlugTree)
		tree.PutData(nil, "repo data")
	}, "specifying nil repository should not cause panic")
}

func (suite *SlugTreeTestSuite) Test_PutData_WhenNil() {
	suite.Assert().Panics(func() {
		var tree SlugTree
		tree.PutData(&Slug{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  "common-go",
		}, "repo data")
	}, "calling PutData on nil tree should cause panic")
}

func (suite *SlugTreeTestSuite) Test_PutDataMulti_WithNewRepositories() {
	tree := make(SlugTree)
	slugs := []*Slug{
		{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  "common-go",
		},
		{
			Saas:  "github.com",
			Owner: "golang",
			Name:  "go",
		},
	}
	expectedRepoData := "repo data"
	tree.PutDataMulti(slugs, expectedRepoData)
	for _, slug := range slugs {
		actualRepoData := tree.GetData(slug)
		suite.Assert().Equal(expectedRepoData, actualRepoData, "new data should be stored and retrieved for a repository: %v", *slug)
	}
}

func (suite *SlugTreeTestSuite) Test_PutDataMulti_WithExistingRepositories() {
	tree := make(SlugTree)
	slugs := []*Slug{
		{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  "common-go",
		},
		{
			Saas:  "github.com",
			Owner: "golang",
			Name:  "go",
		},
	}
	expectedRepoData := "repo data"
	tree.PutDataMulti(slugs, expectedRepoData+"-old")
	tree.PutDataMulti(slugs, expectedRepoData)
	for _, slug := range slugs {
		actualRepoData := tree.GetData(slug)
		suite.Assert().Equal(expectedRepoData, actualRepoData, "data should be overwritten for a repository: %v", *slug)
	}
}

func (suite *SlugTreeTestSuite) Test_PutDataMulti_WithValidAndNilRepositories() {
	tree := make(SlugTree)
	slugs := []*Slug{
		nil,
		{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  "common-go",
		},
		nil,
		{
			Saas:  "github.com",
			Owner: "golang",
			Name:  "go",
		},
		nil,
	}
	expectedRepoData := "repo data"
	tree.PutDataMulti(slugs, expectedRepoData+"-old")
	tree.PutDataMulti(slugs, expectedRepoData)
	for _, slug := range slugs {
		if slug != nil {
			actualRepoData := tree.GetData(slug)
			suite.Assert().Equal(expectedRepoData, actualRepoData, "data should be stored for a valud repository: %v", *slug)
		}
	}
}

func (suite *SlugTreeTestSuite) Test_PutDataMulti_WittNilRepositorySlice() {
	suite.Assert().NotPanics(func() {
		tree := make(SlugTree)
		tree.PutDataMulti(nil, "repo data")
	}, "specifying nil repository slice should not cause panic")
}

func (suite *SlugTreeTestSuite) Test_PutDataMulti_WhenNil() {
	suite.Assert().Panics(func() {
		var tree SlugTree
		tree.PutDataMulti([]*Slug{
			{
				Saas:  "github.com",
				Owner: "tractrix",
				Name:  "common-go",
			},
		}, "repo data")
	}, "calling PutDataMulti with valid repositories on nil tree should cause panic")
}

func (suite *SlugTreeTestSuite) Test_Empty_WhenSomeData() {
	tree := make(SlugTree)
	tree.PutData(&Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}, "repo data")
	suite.Assert().False(tree.Empty(), "false should be returned from tree holding some data")
}

func (suite *SlugTreeTestSuite) Test_Empty_WhenNoData() {
	tree := make(SlugTree)
	suite.Assert().True(tree.Empty(), "true should be returned from empty tree")
}

func (suite *SlugTreeTestSuite) Test_Empty_WhenNil() {
	var tree SlugTree
	suite.Assert().True(tree.Empty(), "true should be returned from nil tree")
}

type ReferenceTreeTestSuite struct {
	suite.Suite
}

func Test_ReferenceTreeTestSuite(t *testing.T) {
	suite.Run(t, new(ReferenceTreeTestSuite))
}

func (suite *ReferenceTreeTestSuite) Test_GetData_WithValidReference() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	expectedRefData := "ref data"
	tree := ReferenceTree{
		ref.Type: map[string]interface{}{
			ref.Name: expectedRefData,
		},
	}
	actualRefData := tree.GetData(ref)
	suite.Assert().Equal(expectedRefData, actualRefData, "appropriate reference data should be returned")
}

func (suite *ReferenceTreeTestSuite) Test_GetData_WithNilReference() {
	tree := ReferenceTree{
		ReferenceTypeBranch: map[string]interface{}{
			"master": "ref data",
		},
	}
	suite.Assert().Nil(tree.GetData(nil))
}

func (suite *ReferenceTreeTestSuite) Test_GetData_WithUnknownType() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	tree := ReferenceTree{
		ref.Type + "-unknown": map[string]interface{}{
			ref.Name: "ref data",
		},
	}
	suite.Assert().Nil(tree.GetData(ref))
}

func (suite *ReferenceTreeTestSuite) Test_GetData_WithUnknownName() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	tree := ReferenceTree{
		ref.Type: map[string]interface{}{
			ref.Name + "-unknown": "ref data",
		},
	}
	suite.Assert().Nil(tree.GetData(ref))
}

func (suite *ReferenceTreeTestSuite) Test_GetData_WhenNil() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	var tree ReferenceTree
	suite.Assert().Nil(tree.GetData(ref))
}

func (suite *ReferenceTreeTestSuite) Test_PutData_WithNewReference() {
	tree := make(ReferenceTree)
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	expectedRefData := "ref data"
	tree.PutData(ref, expectedRefData)
	actualRefData := tree.GetData(ref)
	suite.Assert().Equal(expectedRefData, actualRefData, "new data should be stored and retrieved for a reference")
}

func (suite *ReferenceTreeTestSuite) Test_PutData_WithExistingReference() {
	tree := make(ReferenceTree)
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	expectedRefData := "ref data"
	tree.PutData(ref, expectedRefData+"-old")
	tree.PutData(ref, expectedRefData)
	actualRefData := tree.GetData(ref)
	suite.Assert().Equal(expectedRefData, actualRefData, "data should be overwritten for a reference")
}

func (suite *ReferenceTreeTestSuite) Test_PutData_WittNilReference() {
	suite.Assert().NotPanics(func() {
		tree := make(ReferenceTree)
		tree.PutData(nil, "ref data")
	}, "specifying nil reference should not cause panic")
}

func (suite *ReferenceTreeTestSuite) Test_PutData_WhenNil() {
	suite.Assert().Panics(func() {
		var tree ReferenceTree
		tree.PutData(&Reference{
			Type: ReferenceTypeBranch,
			Name: "master",
		}, "ref data")
	}, "calling PutData on nil tree should cause panic")
}

func (suite *ReferenceTreeTestSuite) Test_PutDataMulti_WithNewReferences() {
	tree := make(ReferenceTree)
	refs := []*Reference{
		{
			Type: ReferenceTypeBranch,
			Name: "master",
		},
		{
			Type: ReferenceTypeTag,
			Name: "0.0.1",
		},
	}
	expectedRefData := "ref data"
	tree.PutDataMulti(refs, expectedRefData)
	for _, ref := range refs {
		actualRefData := tree.GetData(ref)
		suite.Assert().Equal(expectedRefData, actualRefData, "new data should be stored and retrieved for a reference: %v", *ref)
	}
}

func (suite *ReferenceTreeTestSuite) Test_PutDataMulti_WithExistingReferences() {
	tree := make(ReferenceTree)
	refs := []*Reference{
		{
			Type: ReferenceTypeBranch,
			Name: "master",
		},
		{
			Type: ReferenceTypeTag,
			Name: "0.0.1",
		},
	}
	expectedRefData := "ref data"
	tree.PutDataMulti(refs, expectedRefData+"-old")
	tree.PutDataMulti(refs, expectedRefData)
	for _, ref := range refs {
		actualRefData := tree.GetData(ref)
		suite.Assert().Equal(expectedRefData, actualRefData, "data should be overwritten for a reference: %v", *ref)
	}
}

func (suite *ReferenceTreeTestSuite) Test_PutDataMulti_WithValidAndNilReferences() {
	tree := make(ReferenceTree)
	refs := []*Reference{
		nil,
		{
			Type: ReferenceTypeBranch,
			Name: "master",
		},
		nil,
		{
			Type: ReferenceTypeTag,
			Name: "0.0.1",
		},
		nil,
	}
	expectedRefData := "ref data"
	tree.PutDataMulti(refs, expectedRefData)
	for _, ref := range refs {
		if ref != nil {
			actualRefData := tree.GetData(ref)
			suite.Assert().Equal(expectedRefData, actualRefData, "data should be store for a valid reference: %v", *ref)
		}
	}
}

func (suite *ReferenceTreeTestSuite) Test_PutDataMulti_WithNilReferenceSlice() {
	suite.Assert().NotPanics(func() {
		tree := make(ReferenceTree)
		tree.PutDataMulti(nil, "ref data")
	}, "specifying nil reference slice should not cause panic")
}

func (suite *ReferenceTreeTestSuite) Test_PutDataMulti_WhenNil() {
	suite.Assert().Panics(func() {
		var tree ReferenceTree
		tree.PutDataMulti([]*Reference{
			{
				Type: ReferenceTypeBranch,
				Name: "master",
			},
		}, "ref data")
	}, "calling PutDataMulti with valid references on nil tree should cause panic")
}

func (suite *ReferenceTreeTestSuite) Test_Empty_WhenSomeData() {
	tree := make(ReferenceTree)
	tree.PutData(&Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}, "ref data")
	suite.Assert().False(tree.Empty(), "false should be returned from tree holding some data")
}

func (suite *ReferenceTreeTestSuite) Test_Empty_WhenNoData() {
	tree := make(ReferenceTree)
	suite.Assert().True(tree.Empty(), "true should be returned from empty tree")
}

func (suite *ReferenceTreeTestSuite) Test_Empty_WhenNil() {
	var tree ReferenceTree
	suite.Assert().True(tree.Empty(), "true should be returned from nil tree")
}
