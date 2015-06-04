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

type ServiceWebhookTestSuite struct {
	ServiceTestSuiteBase
}

func Test_ServiceWebhookTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceWebhookTestSuite))
}

func (suite *ServiceWebhookTestSuite) Test_obtainWebhookID_WithValidWebhook() {
	expectedHookID := 100
	webhook := &repository.Webhook{
		ID: strconv.Itoa(expectedHookID),
	}
	actualHookID, err := suite.service.obtainWebhookID(webhook)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedHookID, actualHookID)
}

func (suite *ServiceWebhookTestSuite) Test_obtainWebhookID_WithInvalidID() {
	webhook := &repository.Webhook{
		ID: "invalid id",
	}
	actualHookID, err := suite.service.obtainWebhookID(webhook)
	suite.Assert().Error(err)
	suite.Assert().Equal(0, actualHookID)
}

func (suite *ServiceWebhookTestSuite) Test_obtainWebhookID_WithNil() {
	actualHookID, err := suite.service.obtainWebhookID(nil)
	suite.Assert().Error(err)
	suite.Assert().Equal(0, actualHookID)
}

func (suite *ServiceWebhookTestSuite) Test_RegisterWebhook_WithValidParameters() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	expectedHookURL := "https://docstand.tractrix.io/github/webhook"
	expectedHookID := "100"
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		hook := new(github.Hook)
		err := json.NewDecoder(r.Body).Decode(hook)
		suite.Assert().NoError(err)
		suite.Assert().Equal(expectedHookURL, hook.Config["url"], "hook url specified should be sent to github")

		fmt.Fprintf(w, `{"id":%s}`, expectedHookID)
	})

	expectedWebhook := &repository.Webhook{
		ID: expectedHookID,
	}
	actualWebhook, err := suite.service.RegisterWebhook(slug, expectedHookURL)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedWebhook, actualWebhook, "registered webhook information should be returned")
}

func (suite *ServiceWebhookTestSuite) Test_RegisterWebhook_WithUnsupportedRepositorySlug() {
	slug := &repository.Slug{
		Saas:  "__unsupported__",
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	hookURL := "https://docstand.tractrix.io/github/webhook"
	actualWebhook, err := suite.service.RegisterWebhook(slug, hookURL)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualWebhook)
}

func (suite *ServiceWebhookTestSuite) Test_RegisterWebhook_WithNilRepositorySlug() {
	hookURL := "https://docstand.tractrix.io/github/webhook"
	actualWebhook, err := suite.service.RegisterWebhook(nil, hookURL)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualWebhook)
}

func (suite *ServiceWebhookTestSuite) Test_RegisterWebhook_WhenNoWebhookID() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{}`)
	})

	hookURL := "https://docstand.tractrix.io/github/webhook"
	actualWebhook, err := suite.service.RegisterWebhook(slug, hookURL)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualWebhook)
}

func (suite *ServiceWebhookTestSuite) Test_RegisterWebhook_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	hookURL := "https://docstand.tractrix.io/github/webhook"
	actualWebhook, err := suite.service.RegisterWebhook(slug, hookURL)
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Nil(actualWebhook)
}

func (suite *ServiceWebhookTestSuite) Test_RegisterWebhook_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	hookURL := "https://docstand.tractrix.io/github/webhook"
	actualWebhook, err := suite.service.RegisterWebhook(slug, hookURL)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualWebhook)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WithValidSlugAndWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	err := suite.service.UnregisterWebhook(slug, webhook)
	suite.Assert().NoError(err)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WhenNotFound() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	err := suite.service.UnregisterWebhook(slug, webhook)
	suite.Assert().Equal(repository.ErrNotFound, err)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	err := suite.service.UnregisterWebhook(slug, webhook)
	suite.Assert().Error(err)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WithInvalidSlug() {
	slug := &repository.Slug{}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	err := suite.service.UnregisterWebhook(slug, webhook)
	suite.Assert().Error(err)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WithNilSlug() {
	webhook := &repository.Webhook{
		ID: "100",
	}
	err := suite.service.UnregisterWebhook(nil, webhook)
	suite.Assert().Error(err)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WithInvalidWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "invalid id",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/0", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	err := suite.service.UnregisterWebhook(slug, webhook)
	suite.Assert().Error(err)
}

func (suite *ServiceWebhookTestSuite) Test_UnregisterWebhook_WithNilWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	err := suite.service.UnregisterWebhook(slug, nil)
	suite.Assert().Error(err)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WithExistingWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"id":%s}`, webhook.ID)
	})

	exists, err := suite.service.ExistsWebhook(slug, webhook)
	suite.Assert().NoError(err)
	suite.Assert().True(exists)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WithUnexistingWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	exists, err := suite.service.ExistsWebhook(slug, webhook)
	suite.Assert().NoError(err)
	suite.Assert().False(exists)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WhenServerError() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	exists, err := suite.service.ExistsWebhook(slug, webhook)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WithInvalidSlug() {
	slug := &repository.Slug{}
	webhook := &repository.Webhook{
		ID: "100",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/%s", slug.Owner, slug.Name, webhook.ID), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	exists, err := suite.service.ExistsWebhook(slug, webhook)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WithNilSlug() {
	webhook := &repository.Webhook{
		ID: "100",
	}
	exists, err := suite.service.ExistsWebhook(nil, webhook)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WithInvalidWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	webhook := &repository.Webhook{
		ID: "invalid id",
	}
	suite.mux.HandleFunc(fmt.Sprintf("/repos/%s/%s/hooks/0", slug.Owner, slug.Name), func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Fail("no github api should be called")
	})

	exists, err := suite.service.ExistsWebhook(slug, webhook)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}

func (suite *ServiceWebhookTestSuite) Test_ExistsWebhook_WithNilWebhook() {
	slug := &repository.Slug{
		Saas:  ServiceName,
		Owner: "tractrix",
		Name:  "docstand",
	}
	exists, err := suite.service.ExistsWebhook(slug, nil)
	suite.Assert().Error(err)
	suite.Assert().False(exists)
}
