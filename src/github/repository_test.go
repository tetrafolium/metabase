package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/go-github/github"
	"github.com/stretchr/testify/suite"

	"github.com/tractrix/common-go/repository"
)

type ServiceRepositoryTestSuite struct {
	ServiceTestSuiteBase
}

func Test_ServiceRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceRepositoryTestSuite))
}

func githubPermissions(admin bool) *map[string]bool {
	return &map[string]bool{
		"admin": admin,
	}
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositorySlug_WithValidUserRepository() {
	expectedSlug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "user",
		Name:  "user-repo",
	}
	githubRepo := &github.Repository{
		Owner: &github.User{
			Login: github.String(expectedSlug.Owner),
			Name:  github.String("display name"),
		},
		Name: github.String(expectedSlug.Name),
	}
	actualSlug := suite.service.newRepositorySlug(githubRepo)
	suite.Assert().Equal(expectedSlug, actualSlug)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositorySlug_WithInvalidUserRepository() {
	expectedSlug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "",
		Name:  "user-repo",
	}
	githubRepo := &github.Repository{
		Owner: &github.User{
			Login: nil, // Having no login field.
			Name:  github.String("display name"),
		},
		Name: github.String(expectedSlug.Name),
	}
	actualSlug := suite.service.newRepositorySlug(githubRepo)
	suite.Assert().Equal(expectedSlug, actualSlug)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositorySlug_WithValidOrganizationRepository() {
	expectedSlug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "org",
		Name:  "org-repo",
	}
	githubRepo := &github.Repository{
		Organization: &github.Organization{
			Login: github.String(expectedSlug.Owner),
			Name:  github.String("display name"),
		},
		Name: github.String(expectedSlug.Name),
	}
	actualSlug := suite.service.newRepositorySlug(githubRepo)
	suite.Assert().Equal(expectedSlug, actualSlug)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositorySlug_WithInvalidUOrganizationRepository() {
	expectedSlug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "",
		Name:  "org-repo",
	}
	githubRepo := &github.Repository{
		Organization: &github.Organization{
			Login: nil, // Having no login field.
			Name:  github.String("display name"),
		},
		Name: github.String(expectedSlug.Name),
	}
	actualSlug := suite.service.newRepositorySlug(githubRepo)
	suite.Assert().Equal(expectedSlug, actualSlug)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositorySlug_WithEmptyRepository() {
	expectedSlug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "",
		Name:  "",
	}
	githubRepo := &github.Repository{}
	actualSlug := suite.service.newRepositorySlug(githubRepo)
	suite.Assert().Equal(expectedSlug, actualSlug)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositorySlug_WithNil() {
	expectedSlug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "",
		Name:  "",
	}
	actualSlug := suite.service.newRepositorySlug(nil)
	suite.Assert().Equal(expectedSlug, actualSlug)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryID_WithValidUserRepository() {
	expectedOwnerID := 100
	expectedRepoID := 200
	expectedID := &repository.ID{
		Saas:    ServiceName,
		OwnerID: strconv.Itoa(expectedOwnerID),
		ID:      strconv.Itoa(expectedRepoID),
	}
	githubRepo := &github.Repository{
		Owner: &github.User{
			ID: github.Int(expectedOwnerID),
		},
		ID: github.Int(expectedRepoID),
	}
	actualID := suite.service.newRepositoryID(githubRepo)
	suite.Assert().Equal(expectedID, actualID)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryID_WithInvalidUserRepository() {
	expectedRepoID := 200
	expectedID := &repository.ID{
		Saas:    ServiceName,
		OwnerID: "",
		ID:      strconv.Itoa(expectedRepoID),
	}
	githubRepo := &github.Repository{
		Owner: &github.User{}, // Having no ID field.
		ID:    github.Int(expectedRepoID),
	}
	actualID := suite.service.newRepositoryID(githubRepo)
	suite.Assert().Equal(expectedID, actualID)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryID_WithValidOrganizationRepository() {
	expectedOwnerID := 100
	expectedRepoID := 200
	expectedID := &repository.ID{
		Saas:    ServiceName,
		OwnerID: strconv.Itoa(expectedOwnerID),
		ID:      strconv.Itoa(expectedRepoID),
	}
	githubRepo := &github.Repository{
		Organization: &github.Organization{
			ID: github.Int(expectedOwnerID),
		},
		ID: github.Int(expectedRepoID),
	}
	actualID := suite.service.newRepositoryID(githubRepo)
	suite.Assert().Equal(expectedID, actualID)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryID_WithInvalidUOrganizationRepository() {
	expectedRepoID := 200
	expectedID := &repository.ID{
		Saas:    ServiceName,
		OwnerID: "",
		ID:      strconv.Itoa(expectedRepoID),
	}
	githubRepo := &github.Repository{
		Organization: &github.Organization{}, // Having no login field.
		ID:           github.Int(expectedRepoID),
	}
	actualID := suite.service.newRepositoryID(githubRepo)
	suite.Assert().Equal(expectedID, actualID)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryID_WithEmptyRepository() {
	expectedID := &repository.ID{
		Saas:    ServiceName,
		OwnerID: "",
		ID:      "",
	}
	githubRepo := &github.Repository{}
	actualID := suite.service.newRepositoryID(githubRepo)
	suite.Assert().Equal(expectedID, actualID)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryID_WithNil() {
	expectedID := &repository.ID{
		Saas:    ServiceName,
		OwnerID: "",
		ID:      "",
	}
	actualID := suite.service.newRepositoryID(nil)
	suite.Assert().Equal(expectedID, actualID)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryPermissions_WithValidPermissions() {
	expectedPermissions := &repository.Permissions{
		Admin: true,
	}
	githubRepo := &github.Repository{
		Permissions: githubPermissions(expectedPermissions.Admin),
	}
	actualPermissions := suite.service.newRepositoryPermissions(githubRepo)
	suite.Assert().Equal(expectedPermissions, actualPermissions)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryPermissions_WithNilPermissions() {
	expectedPermissions := &repository.Permissions{
		Admin: false,
	}
	githubRepo := &github.Repository{
		Permissions: nil,
	}
	actualPermissions := suite.service.newRepositoryPermissions(githubRepo)
	suite.Assert().Equal(expectedPermissions, actualPermissions)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepositoryPermissions_WithNil() {
	expectedPermissions := &repository.Permissions{
		Admin: false,
	}
	actualPermissions := suite.service.newRepositoryPermissions(nil)
	suite.Assert().Equal(expectedPermissions, actualPermissions)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepository_WithCompleteGitHubRepository() {
	expectedOwnerName := "user"
	expectedOwnerID := 100
	expectedRepoName := "user-repo"
	expectedRepoID := 200
	expectedAdmin := true
	expectedRepo := &repository.Repository{
		Slug: repository.Slug{
			Saas:  ServiceName,
			Owner: expectedOwnerName,
			Name:  expectedRepoName,
		},
		ID: repository.ID{
			Saas:    ServiceName,
			OwnerID: strconv.Itoa(expectedOwnerID),
			ID:      strconv.Itoa(expectedRepoID),
		},
		Permissions: repository.Permissions{
			Admin: expectedAdmin,
		},
	}
	githubRepo := &github.Repository{
		Owner: &github.User{
			Login: github.String(expectedOwnerName),
			Name:  github.String("display name"), // This should be ignored.
			ID:    github.Int(expectedOwnerID),
		},
		Name:        github.String(expectedRepoName),
		ID:          github.Int(expectedRepoID),
		Permissions: githubPermissions(expectedAdmin),
	}
	actualRepo := suite.service.newRepository(githubRepo)
	suite.Assert().Equal(expectedRepo, actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_newRepository_WithIncompleteGitHubRepository() {
	expectedRepo := &repository.Repository{
		Slug:        repository.Slug{Saas: ServiceName},
		ID:          repository.ID{Saas: ServiceName},
		Permissions: repository.Permissions{Admin: false},
	}
	githubRepo := &github.Repository{}
	actualRepo := suite.service.newRepository(githubRepo)
	suite.Assert().Equal(expectedRepo, actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_validateRepositoryID_WithValidID() {
	err := suite.service.validateRepositoryID(&repository.ID{
		Saas:    ServiceName,
		OwnerID: "100",
		ID:      "200",
	})
	suite.Assert().NoError(err)
}

func (suite *ServiceRepositoryTestSuite) Test_validateRepositoryID_WithNil() {
	err := suite.service.validateRepositoryID(nil)
	suite.Assert().Error(err)
}

func (suite *ServiceRepositoryTestSuite) Test_validateRepositoryID_WithInvalidOwnerID() {
	err := suite.service.validateRepositoryID(&repository.ID{
		Saas:    ServiceName,
		OwnerID: "invalid owner id",
		ID:      "200",
	})
	suite.Assert().Error(err)
}

func (suite *ServiceRepositoryTestSuite) Test_validateRepositoryID_WithInvalidID() {
	err := suite.service.validateRepositoryID(&repository.ID{
		Saas:    ServiceName,
		OwnerID: "100",
		ID:      "invalid id",
	})
	suite.Assert().Error(err)
}

func (suite *ServiceRepositoryTestSuite) Test_listOrganizations_WhenMultiplePages() {
	authenticatedUser := ""
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header()["Link"] = []string{
				`<https://api.github.com/resource?page=2>; rel="next"`,
				`<https://api.github.com/resource?page=2>; rel="last"`,
			}
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		case "2":
			fmt.Fprint(w, `[{"id":3},{"id":4}]`)
		default:
			suite.Assert().Fail("invalid page: %s", page)
		}
	})

	expectedOrgs := []github.Organization{
		{ID: github.Int(1)},
		{ID: github.Int(2)},
		{ID: github.Int(3)},
		{ID: github.Int(4)},
	}
	actualOrgs, err := suite.service.listOrganizations(authenticatedUser)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedOrgs, actualOrgs, "all organizations should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listOrganizations_WhenNotFound() {
	authenticatedUser := ""
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualOrgs, err := suite.service.listOrganizations(authenticatedUser)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Len(actualOrgs, 0, "no organizations should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listOrganizations_WhenServerError() {
	authenticatedUser := ""
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualOrgs, err := suite.service.listOrganizations(authenticatedUser)
	suite.Assert().Error(err)
	suite.Assert().Len(actualOrgs, 0, "no organizations should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listRepositoriesByOrganization_WhenMultiplePages() {
	orgName := "o"
	suite.mux.HandleFunc(fmt.Sprintf("/orgs/%s/repos", orgName), func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header()["Link"] = []string{
				`<https://api.github.com/resource?page=2>; rel="next"`,
				`<https://api.github.com/resource?page=2>; rel="last"`,
			}
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		case "2":
			fmt.Fprint(w, `[{"id":3},{"id":4}]`)
		default:
			suite.Assert().Fail("invalid page: %s", page)
		}
	})

	expectedRepos := []github.Repository{
		{ID: github.Int(1)},
		{ID: github.Int(2)},
		{ID: github.Int(3)},
		{ID: github.Int(4)},
	}
	actualRepos, err := suite.service.listRepositoriesByOrganization(orgName)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRepos, actualRepos, "all repositories for an organization should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listRepositoriesByOrganization_WhenNotFound() {
	orgName := "o"
	suite.mux.HandleFunc(fmt.Sprintf("/orgs/%s/repos", orgName), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRepos, err := suite.service.listRepositoriesByOrganization(orgName)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listRepositoriesByOrganization_WhenServerError() {
	orgName := "o"
	suite.mux.HandleFunc(fmt.Sprintf("/orgs/%s/repos", orgName), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepos, err := suite.service.listRepositoriesByOrganization(orgName)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listRepositoriesByUser_WhenMultiplePages() {
	authenticatedUser := ""
	suite.mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		switch page {
		case "":
			w.Header()["Link"] = []string{
				`<https://api.github.com/resource?page=2>; rel="next"`,
				`<https://api.github.com/resource?page=2>; rel="last"`,
			}
			fmt.Fprint(w, `[{"id":1},{"id":2}]`)
		case "2":
			fmt.Fprint(w, `[{"id":3},{"id":4}]`)
		default:
			suite.Assert().Fail("invalid page: %s", page)
		}
	})

	expectedRepos := []github.Repository{
		{ID: github.Int(1)},
		{ID: github.Int(2)},
		{ID: github.Int(3)},
		{ID: github.Int(4)},
	}
	actualRepos, err := suite.service.listRepositoriesByUser(authenticatedUser)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRepos, actualRepos, "all repositories for a user should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listRepositoriesByUser_WhenNotFound() {
	authenticatedUser := ""
	suite.mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRepos, err := suite.service.listRepositoriesByUser(authenticatedUser)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_listRepositoriesByUser_WhenServerError() {
	authenticatedUser := ""
	suite.mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepos, err := suite.service.listRepositoriesByUser(authenticatedUser)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryBySlug_WithValidID() {
	expectedRepo := &repository.Repository{
		Slug: repository.Slug{
			Saas:  ServiceName,
			Owner: "org",
			Name:  "org-repo",
		},
		ID: repository.ID{
			Saas:    ServiceName,
			OwnerID: "100",
			ID:      "200",
		},
		Permissions: repository.Permissions{
			Admin: true,
		},
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s", expectedRepo.Slug.Owner, expectedRepo.Slug.Name), func(w http.ResponseWriter, r *http.Request) {
		ownerID, _ := strconv.Atoi(expectedRepo.ID.OwnerID)
		repoID, _ := strconv.Atoi(expectedRepo.ID.ID)
		body, err := json.Marshal(&github.Repository{
			Organization: &github.Organization{
				Login: github.String(expectedRepo.Slug.Owner),
				ID:    github.Int(ownerID),
			},
			ID:          github.Int(repoID),
			Name:        github.String(expectedRepo.Slug.Name),
			Permissions: githubPermissions(expectedRepo.Permissions.Admin),
		})
		suite.Require().NoError(err)
		w.Write(body)
	})

	actualRepo, err := suite.service.GetRepositoryBySlug(&expectedRepo.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRepo, actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryBySlug_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "org",
		Name:  "org-repo",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRepo, err := suite.service.GetRepositoryBySlug(slug)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Nil(actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryBySlug_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "org",
		Name:  "org-repo",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepo, err := suite.service.GetRepositoryBySlug(slug)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryBySlug_WithNil() {
	repo, err := suite.service.GetRepositoryBySlug(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(repo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryByID_WithValidID() {
	expectedRepo := &repository.Repository{
		Slug: repository.Slug{
			Saas:  ServiceName,
			Owner: "org",
			Name:  "org-repo",
		},
		ID: repository.ID{
			Saas:    ServiceName,
			OwnerID: "100",
			ID:      "200",
		},
		Permissions: repository.Permissions{
			Admin: true,
		},
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repositories/%s", expectedRepo.ID.ID), func(w http.ResponseWriter, r *http.Request) {
		ownerID, _ := strconv.Atoi(expectedRepo.ID.OwnerID)
		repoID, _ := strconv.Atoi(expectedRepo.ID.ID)
		body, err := json.Marshal(&github.Repository{
			Organization: &github.Organization{
				Login: github.String(expectedRepo.Slug.Owner),
				ID:    github.Int(ownerID),
			},
			ID:          github.Int(repoID),
			Name:        github.String(expectedRepo.Slug.Name),
			Permissions: githubPermissions(expectedRepo.Permissions.Admin),
		})
		suite.Require().NoError(err)
		w.Write(body)
	})

	actualRepo, err := suite.service.GetRepositoryByID(&expectedRepo.ID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRepo, actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryByID_WithInconsistentID() {
	repo := &repository.Repository{
		Slug: repository.Slug{
			Saas:  ServiceName,
			Owner: "org",
			Name:  "org-repo",
		},
		ID: repository.ID{
			Saas:    ServiceName,
			OwnerID: "100",
			ID:      "200",
		},
		Permissions: repository.Permissions{
			Admin: false,
		},
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repositories/%s", repo.ID.ID), func(w http.ResponseWriter, r *http.Request) {
		ownerID, _ := strconv.Atoi(repo.ID.OwnerID)
		repoID, _ := strconv.Atoi(repo.ID.ID)
		body, err := json.Marshal(&github.Repository{
			Organization: &github.Organization{
				Login: github.String(repo.Slug.Owner),
				ID:    github.Int(ownerID + 1), // Inconsistent owner ID returned.
			},
			ID:          github.Int(repoID),
			Name:        github.String(repo.Slug.Name),
			Permissions: githubPermissions(repo.Permissions.Admin),
		})
		suite.Require().NoError(err)
		w.Write(body)
	})

	actualRepo, err := suite.service.GetRepositoryByID(&repo.ID)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryByID_WhenNotFound() {
	id := &repository.ID{
		Saas:    ServiceName,
		OwnerID: "100",
		ID:      "200",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repositories/%s", id.ID), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualRepo, err := suite.service.GetRepositoryByID(id)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Nil(actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryByID_WhenServerError() {
	id := &repository.ID{
		Saas:    ServiceName,
		OwnerID: "100",
		ID:      "200",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repositories/%s", id.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepo, err := suite.service.GetRepositoryByID(id)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRepo)
}

func (suite *ServiceRepositoryTestSuite) Test_GetRepositoryByID_WithNil() {
	repo, err := suite.service.GetRepositoryByID(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(repo)
}

func (suite *ServiceRepositoryTestSuite) Test_ListRepositories_WhenBothTypesOfRepositories() {
	expectedOrgRepo := github.Repository{
		Organization: &github.Organization{
			Login: github.String("org"),
			ID:    github.Int(100),
		},
		Name:        github.String("org-repo"),
		ID:          github.Int(200),
		Permissions: githubPermissions(false),
	}
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal([]github.Organization{*expectedOrgRepo.Organization})
		suite.Require().NoError(err)
		w.Write(body)
	})
	suite.mux.HandleFunc(fmt.Sprintf("/orgs/%s/repos", *expectedOrgRepo.Organization.Login), func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal([]github.Repository{expectedOrgRepo})
		suite.Require().NoError(err)
		w.Write(body)
	})

	expectedUserRepo := github.Repository{
		Owner: &github.User{
			Login: github.String("user"),
			ID:    github.Int(101),
		},
		Name:        github.String("user-repo"),
		ID:          github.Int(201),
		Permissions: githubPermissions(true),
	}
	suite.mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal([]github.Repository{expectedUserRepo})
		suite.Require().NoError(err)
		w.Write(body)
	})

	expectedDataOrgRepo := &repository.Repository{
		Slug: repository.Slug{
			Saas:  ServiceName,
			Owner: *expectedOrgRepo.Organization.Login,
			Name:  *expectedOrgRepo.Name,
		},
		ID: repository.ID{
			Saas:    ServiceName,
			OwnerID: strconv.Itoa(*expectedOrgRepo.Organization.ID),
			ID:      strconv.Itoa(*expectedOrgRepo.ID),
		},
		Permissions: repository.Permissions{
			Admin: (*expectedOrgRepo.Permissions)["admin"],
		},
	}
	expectedDataUserRepo := &repository.Repository{
		Slug: repository.Slug{
			Saas:  ServiceName,
			Owner: *expectedUserRepo.Owner.Login,
			Name:  *expectedUserRepo.Name,
		},
		ID: repository.ID{
			Saas:    ServiceName,
			OwnerID: strconv.Itoa(*expectedUserRepo.Owner.ID),
			ID:      strconv.Itoa(*expectedUserRepo.ID),
		},
		Permissions: repository.Permissions{
			Admin: (*expectedUserRepo.Permissions)["admin"],
		},
	}
	actualRepos, err := suite.service.ListRepositories()
	suite.Assert().NoError(err)
	// Retrieving organization repositories and user repositories are done in parallel,
	// thus the order of repositories in the list returned is not fixed.
	suite.Assert().Len(actualRepos, 2, "all repositories should be returned")
	suite.Assert().Contains(actualRepos, expectedDataOrgRepo, "organization repositories should be returned")
	suite.Assert().Contains(actualRepos, expectedDataUserRepo, "user repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_ListRepositories_WhenListingOrganizationsFails() {
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepos, err := suite.service.ListRepositories()
	suite.Assert().Error(err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_ListRepositories_WhenListingOrganizationRepositoriesFails() {
	org := github.Organization{
		Login: github.String("org"),
	}
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal([]github.Organization{org})
		suite.Require().NoError(err)
		w.Write(body)
	})
	suite.mux.HandleFunc(fmt.Sprintf("/orgs/%s/repos", *org.Login), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepos, err := suite.service.ListRepositories()
	suite.Assert().Error(err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}

func (suite *ServiceRepositoryTestSuite) Test_ListRepositories_WhenListingUserRepositoriesFails() {
	orgRepo := github.Repository{
		Organization: &github.Organization{
			Login: github.String("org"),
		},
		Name: github.String("org-repo"),
	}
	suite.mux.HandleFunc("/user/orgs", func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal([]github.Organization{*orgRepo.Organization})
		suite.Require().NoError(err)
		w.Write(body)
	})
	suite.mux.HandleFunc(fmt.Sprintf("/orgs/%s/repos", *orgRepo.Organization.Login), func(w http.ResponseWriter, r *http.Request) {
		body, err := json.Marshal([]github.Repository{orgRepo})
		suite.Require().NoError(err)
		w.Write(body)
	})

	suite.mux.HandleFunc("/user/repos", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualRepos, err := suite.service.ListRepositories()
	suite.Assert().Error(err)
	suite.Assert().Len(actualRepos, 0, "no repositories should be returned")
}
