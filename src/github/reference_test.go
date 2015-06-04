package github

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tractrix/common-go/repository"
)

type ServiceReferenceTestSuite struct {
	ServiceTestSuiteBase
}

func Test_ServiceReferenceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceReferenceTestSuite))
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WithValidSlug() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeBranch, Name: "master"},
		{Type: repository.ReferenceTypeBranch, Name: "list-repository-branches"},
		{Type: repository.ReferenceTypeTag, Name: "v1.0"},
		{Type: repository.ReferenceTypeTag, Name: "list-repository-tags"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  [
		    {"ref": "refs/heads/%s"},
		    {"ref": "refs/heads/%s"},
		    {"ref": "refs/tags/%s"},
		    {"ref": "refs/tags/%s"}
		  ]`, expectedRefs[0].Name, expectedRefs[1].Name, expectedRefs[2].Name, expectedRefs[3].Name)
	})

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references in refs/heads should be returned as branches")
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WhenMultiplePages() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeBranch, Name: "master"},
		{Type: repository.ReferenceTypeTag, Name: "v1.0"},
		{Type: repository.ReferenceTypeBranch, Name: "list-repository-branches"},
		{Type: repository.ReferenceTypeTag, Name: "list-repository-tags"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header()["Link"] = []string{
				`<https://api.github.com/resource?page=2>; rel="next"`,
				`<https://api.github.com/resource?page=2>; rel="last"`,
			}
			fmt.Fprintf(w, `
			  [
			    {"ref": "refs/heads/%s"},
			    {"ref": "refs/tags/%s"}
			  ]`, expectedRefs[0].Name, expectedRefs[1].Name)
		case "2":
			fmt.Fprintf(w, `
			  [
			    {"ref": "refs/heads/%s"},
			    {"ref": "refs/tags/%s"}
			  ]`, expectedRefs[2].Name, expectedRefs[3].Name)
		default:
			suite.Assert().Fail("invalid page: %s", page)
		}
	})

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references in refs/heads or refs/tags should be returned")
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WhenNoRefs() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[]`)
	})

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().NoError(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WhenInvalidRefs() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeBranch, Name: "master"},
		{Type: repository.ReferenceTypeTag, Name: "v1.0"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  [
		    {"ref": "refs/heads/%s"},
		    {"ref": "refs/tags/%s"},
		    {"ref": "refs/unknown/should-be-ignored"},
		    {}
		  ]`, expectedRefs[0].Name, expectedRefs[1].Name)
	})

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references which don't have valid refs should be ignored")
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().Error(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListReferences_WithNilSlug() {
	actualRefs, err := suite.service.ListReferences(nil)
	suite.Assert().Error(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WithValidSlug() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeBranch, Name: "master"},
		{Type: repository.ReferenceTypeBranch, Name: "list-repository-branches"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  [
		    {"ref": "refs/heads/%s"},
		    {"ref": "refs/heads/%s"}
		  ]`, expectedRefs[0].Name, expectedRefs[1].Name)
	})

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references in refs/heads should be returned as branches")
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WhenMultiplePages() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeBranch, Name: "master"},
		{Type: repository.ReferenceTypeBranch, Name: "develop"},
		{Type: repository.ReferenceTypeBranch, Name: "release"},
		{Type: repository.ReferenceTypeBranch, Name: "list-repository-branch"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header()["Link"] = []string{
				`<https://api.github.com/resource?page=2>; rel="next"`,
				`<https://api.github.com/resource?page=2>; rel="last"`,
			}
			fmt.Fprintf(w, `
			  [
			    {"ref": "refs/heads/%s"},
			    {"ref": "refs/heads/%s"}
			  ]`, expectedRefs[0].Name, expectedRefs[1].Name)
		case "2":
			fmt.Fprintf(w, `
			  [
			    {"ref": "refs/heads/%s"},
			    {"ref": "refs/heads/%s"}
			  ]`, expectedRefs[2].Name, expectedRefs[3].Name)
		default:
			suite.Assert().Fail("invalid page: %s", page)
		}
	})

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references in refs/heads should be returned as branch")
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WhenNoBranches() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[]`)
	})

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().NoError(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WhenInvalidRefs() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeBranch, Name: "master"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  [
		    {"ref": "refs/heads/%s"},
		    {"ref": "refs/tags/should-be-ignored"},
		    {}
		  ]`, expectedRefs[0].Name)
	})

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references which don't have valid refs should be ignored")
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().Error(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListBranches_WithNilSlug() {
	actualRefs, err := suite.service.ListBranches(nil)
	suite.Assert().Error(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WithValidSlug() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeTag, Name: "v1.0"},
		{Type: repository.ReferenceTypeTag, Name: "list-repository-tag"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  [
		    {"ref": "refs/tags/%s"},
		    {"ref": "refs/tags/%s"}
		  ]`, expectedRefs[0].Name, expectedRefs[1].Name)
	})

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references in refs/tags should be returned as tag")
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WhenMultiplePages() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "common-go",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeTag, Name: "v1.0"},
		{Type: repository.ReferenceTypeTag, Name: "v1.1"},
		{Type: repository.ReferenceTypeTag, Name: "v1.2"},
		{Type: repository.ReferenceTypeTag, Name: "list-repository-tag"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header()["Link"] = []string{
				`<https://api.github.com/resource?page=2>; rel="next"`,
				`<https://api.github.com/resource?page=2>; rel="last"`,
			}
			fmt.Fprintf(w, `
			  [
			    {"ref": "refs/tags/%s"},
			    {"ref": "refs/tags/%s"}
			  ]`, expectedRefs[0].Name, expectedRefs[1].Name)
		case "2":
			fmt.Fprintf(w, `
			  [
			    {"ref": "refs/tags/%s"},
			    {"ref": "refs/tags/%s"}
			  ]`, expectedRefs[2].Name, expectedRefs[3].Name)
		default:
			suite.Assert().Fail("invalid page: %s", page)
		}
	})

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references in refs/tags should be returned as tag")
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WhenNoTags() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `[]`)
	})

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().NoError(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WhenInvalidRefs() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	expectedRefs := []*repository.Reference{
		{Type: repository.ReferenceTypeTag, Name: "v1.0"},
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  [
		    {"ref": "refs/tags/%s"},
		    {"ref": "refs/heads/should-be-ignored"},
		    {}
		  ]`, expectedRefs[0].Name)
	})

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "references which don't have valid refs should be ignored")
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags", slug.Owner, slug.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().Error(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_ListTags_WithNilSlug() {
	actualRefs, err := suite.service.ListTags(nil)
	suite.Assert().Error(err)
	suite.Assert().Empty(actualRefs)
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WithValidBranch() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	ref := &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
	expectedCommitID := "8e7e232202560331e4476c6ce20468ba5f1749dc"
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", slug.Owner, slug.Name, ref.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  {
		    "object": {
		      "sha": "%s"
		    }
		  }`, expectedCommitID)
	})

	actualCommitID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedCommitID, actualCommitID, "sha hash valud should be returned as commit id")
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WithValidTag() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	ref := &repository.Reference{
		Type: repository.ReferenceTypeTag,
		Name: "0.0.1",
	}
	expectedCommitID := "8e7e232202560331e4476c6ce20468ba5f1749dc"
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/tags/%s", slug.Owner, slug.Name, ref.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  {
		    "object": {
		      "sha": "%s"
		    }
		  }`, expectedCommitID)
	})

	actualCommitID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedCommitID, actualCommitID, "sha hash valud should be returned as commit id")
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WhenNoCommitObject() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	ref := &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", slug.Owner, slug.Name, ref.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{}`)
	})

	commitID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().Error(err)
	suite.Assert().Empty(commitID)
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WhenNoCommitSHA() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	ref := &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", slug.Owner, slug.Name, ref.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `
		  {
		    "object": {}
		  }`)
	})

	commitID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().Error(err)
	suite.Assert().Empty(commitID)
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	ref := &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", slug.Owner, slug.Name, ref.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	commitID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Empty(commitID)
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	ref := &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
	githubAPIPath := fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", slug.Owner, slug.Name, ref.Name)
	suite.mux.HandleFunc(githubAPIPath, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	commitID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().Error(err)
	suite.Assert().Empty(commitID)
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WithNilSlug() {
	ref := &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
	commitID, err := suite.service.GetCommitID(nil, ref)
	suite.Assert().Error(err)
	suite.Assert().Empty(commitID)
}

func (suite *ServiceReferenceTestSuite) Test_GetCommitID_WithNilReference() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	commitID, err := suite.service.GetCommitID(slug, nil)
	suite.Assert().Error(err)
	suite.Assert().Empty(commitID)
}
