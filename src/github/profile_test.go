package github

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/tractrix/common-go/repository"
)

type ServiceProfileTestSuite struct {
	ServiceTestSuiteBase
}

func Test_ServiceProfileTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceProfileTestSuite))
}

func (suite *ServiceProfileTestSuite) Test_GetUserID_WithValidWebhook() {
	expectedUserID := "100"
	suite.mux.HandleFunc(fmt.Sprintf("/user"), func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":100}`)
	})

	actualUserID, err := suite.service.GetUserID()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedUserID, actualUserID, "int ID should be translated into string ID")
}

func (suite *ServiceProfileTestSuite) Test_GetUserID_WhenNotFound() {
	suite.mux.HandleFunc(fmt.Sprintf("/user"), func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	actualUserID, err := suite.service.GetUserID()
	suite.Assert().Equal(repository.ErrNotFound, err)
	suite.Assert().Empty(actualUserID)
}

func (suite *ServiceProfileTestSuite) Test_GetUserID_WhenServerError() {
	suite.mux.HandleFunc(fmt.Sprintf("/user"), func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	actualUserID, err := suite.service.GetUserID()
	suite.Assert().Error(err)
	suite.Assert().Empty(actualUserID)
}
