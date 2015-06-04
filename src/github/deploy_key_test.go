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

type ServiceDeployKeyTestSuite struct {
	ServiceTestSuiteBase
}

func Test_ServiceDeployKeyTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceDeployKeyTestSuite))
}

func (suite *ServiceDeployKeyTestSuite) Test_obtainDeployKeyID_WithValidDeployKey() {
	expectedKeyID := 100
	deployKey := &repository.DeployKey{
		ID: strconv.Itoa(expectedKeyID),
	}
	actualKeyID, err := suite.service.obtainDeployKeyID(deployKey)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedKeyID, actualKeyID)
}

func (suite *ServiceDeployKeyTestSuite) Test_obtainDeployKeyID_WithInvalidID() {
	deployKey := &repository.DeployKey{
		ID: "invalid id",
	}
	actualKeyID, err := suite.service.obtainDeployKeyID(deployKey)
	suite.Assert().Error(err)
	suite.Assert().Equal(0, actualKeyID)
}

func (suite *ServiceDeployKeyTestSuite) Test_obtainDeployKeyID_WithNil() {
	actualKeyID, err := suite.service.obtainDeployKeyID(nil)
	suite.Assert().Error(err)
	suite.Assert().Equal(0, actualKeyID)
}

func (suite *ServiceDeployKeyTestSuite) Test_RegisterDeployKey_WithValidParameters() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	expectedPublicKey := "ssh-rsa AAA..."
	expectedTitle := "key title"
	expectedKeyID := "100"
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		key := new(github.Key)
		err := json.NewDecoder(r.Body).Decode(key)
		suite.Assert().NoError(err)
		suite.Assert().Equal(expectedPublicKey, *key.Key, "public key specified should be sent to github")
		suite.Assert().Equal(expectedTitle, *key.Title, "key title specified should be sent to github")

		fmt.Fprintf(w, `{"id":%s}`, expectedKeyID)
	})

	expectedDeployKey := &repository.DeployKey{
		ID: expectedKeyID,
	}
	actualDeployKey, err := suite.service.RegisterDeployKey(slug, expectedPublicKey, expectedTitle)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedDeployKey, actualDeployKey, "registered deploy key information should be returned")
}

func (suite *ServiceDeployKeyTestSuite) Test_RegisterDeployKey_WithUnsupportedRepositorySlug() {
	slug := &repository.Slug{
		Saas:  "__unsupported__",
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	publicKey := "ssh-rsa AAA..."
	title := "key title"
	actualDeployKey, err := suite.service.RegisterDeployKey(slug, publicKey, title)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualDeployKey)
}

func (suite *ServiceDeployKeyTestSuite) Test_RegisterDeployKey_WithNilRepositorySlug() {
	publicKey := "ssh-rsa AAA..."
	title := "key title"
	actualDeployKey, err := suite.service.RegisterDeployKey(nil, publicKey, title)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualDeployKey)
}

func (suite *ServiceDeployKeyTestSuite) Test_RegisterDeployKey_WhenNoDeployKeyID() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	publicKey := "ssh-rsa AAA..."
	title := "key title"
	actualDeployKey, err := suite.service.RegisterDeployKey(slug, publicKey, title)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualDeployKey)
}

func (suite *ServiceDeployKeyTestSuite) Test_RegisterDeployKey_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	publicKey := "ssh-rsa AAA..."
	title := "key title"
	actualDeployKey, err := suite.service.RegisterDeployKey(slug, publicKey, title)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Nil(actualDeployKey)
}

func (suite *ServiceDeployKeyTestSuite) Test_RegisterDeployKey_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	publicKey := "ssh-rsa AAA..."
	title := "key title"
	actualDeployKey, err := suite.service.RegisterDeployKey(slug, publicKey, title)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualDeployKey)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WithValidSlugAndDeployKey() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().NoError(err)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().Equal(repository.ErrNotFound, err)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().Error(err)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WithInvalidSlug() {
	slug := &repository.Slug{}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().Error(err)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WithNilSlug() {
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	err := suite.service.UnregisterDeployKey(nil, deployKey)
	suite.Assert().Error(err)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WithInvalidDeployKey() {
	slug := &repository.Slug{}
	deployKey := &repository.DeployKey{
		ID: "invalid id",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/0", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().Error(err)
}

func (suite *ServiceDeployKeyTestSuite) Test_UnregisterDeployKey_WithNilDeployKey() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	err := suite.service.UnregisterDeployKey(slug, nil)
	suite.Assert().Error(err)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WithExistingDeployKey() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":%s}`, deployKey.ID)
	})

	exists, err := suite.service.ExistsDeployKey(slug, deployKey)
	suite.Assert().NoError(err)
	suite.Assert().True(exists)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WithUnexistingDeployKey() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	exists, err := suite.service.ExistsDeployKey(slug, deployKey)
	suite.Assert().NoError(err)
	suite.Assert().False(exists)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	exists, err := suite.service.ExistsDeployKey(slug, deployKey)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WithInvalidSlug() {
	slug := &repository.Slug{}
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/%s", slug.Owner, slug.Name, deployKey.ID), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	exists, err := suite.service.ExistsDeployKey(slug, deployKey)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WithNilSlug() {
	deployKey := &repository.DeployKey{
		ID: "100",
	}
	exists, err := suite.service.ExistsDeployKey(nil, deployKey)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WithInvalidDeployKey() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	deployKey := &repository.DeployKey{
		ID: "invalid id",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/keys/0", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	exists, err := suite.service.ExistsDeployKey(slug, deployKey)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceDeployKeyTestSuite) Test_ExistsDeployKey_WithNilDeployKey() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	exists, err := suite.service.ExistsDeployKey(slug, nil)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}
