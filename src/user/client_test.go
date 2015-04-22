package user

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

const getIDAPIPath = "/api/1/"

type ManagerClientTestSuite struct {
	suite.Suite

	mux      *http.ServeMux
	server   *httptest.Server
	endpoint *url.URL
}

func Test_ManagerClientTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerClientTestSuite))
}

func (suite *ManagerClientTestSuite) SetupTest() {
	suite.mux = http.NewServeMux()
	suite.server = httptest.NewServer(suite.mux)
	suite.endpoint, _ = url.Parse(suite.server.URL)
	suite.endpoint.Path = getIDAPIPath
}

func (suite *ManagerClientTestSuite) TearDownTest() {
	suite.server.Close()

	suite.endpoint = nil
	suite.server = nil
	suite.mux = nil
}

func (suite *ManagerClientTestSuite) Test_NewManagerClient_WithNilHTTPClient() {
	client, err := NewManagerClient(nil, suite.endpoint)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(client)
}

func (suite *ManagerClientTestSuite) Test_NewManagerClient_WithValidHTTPClient() {
	client, err := NewManagerClient(http.DefaultClient, suite.endpoint)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(client)
}

func (suite *ManagerClientTestSuite) Test_NewManagerClient_WithNilEndpoint() {
	client, err := NewManagerClient(nil, nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(client)
}

func (suite *ManagerClientTestSuite) Test_GetID_WhenServerSucceeds() {
	expectedID := uint64(100)
	testService := "test-service"
	testServiceID := "test-id"
	suite.mux.HandleFunc(getIDAPIPath+"users/"+testService+"/"+testServiceID, func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Equal("GET", r.Method, "HTTP method should be GET")
		fmt.Fprintf(w, `{"id":"%d", "account":{"service":"%s","id":"%s"}}`, expectedID, testService, testServiceID)
	})

	client, err := NewManagerClient(nil, suite.endpoint)
	suite.Require().NoError(err)

	actualID, err := client.GetID(testService, testServiceID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedID, actualID, "TUID for the service user should be returned")
}

func (suite *ManagerClientTestSuite) Test_GetID_WhenUserNotFound() {
	testService := "test-service"
	testServiceID := "test-id"
	suite.mux.HandleFunc(getIDAPIPath+"users/"+testService+"/"+testServiceID, func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Equal("GET", r.Method, "HTTP method should be GET")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error":{"code":404, "message":"no such user"}}`)
	})

	client, err := NewManagerClient(nil, suite.endpoint)
	suite.Require().NoError(err)

	actualID, err := client.GetID(testService, testServiceID)
	suite.Assert().Error(err)
	suite.Assert().Equal(ErrNoSuchUser, err, "ErrNoSuchUser should be returned when user is not found")
	suite.Assert().Equal(uint64(0), actualID)
}

func (suite *ManagerClientTestSuite) Test_GetID_WhenServerFails() {
	testService := "test-service"
	testServiceID := "test-id"
	suite.mux.HandleFunc(getIDAPIPath+testService+"/"+testServiceID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client, err := NewManagerClient(nil, suite.endpoint)
	suite.Require().NoError(err)

	actualID, err := client.GetID(testService, testServiceID)
	suite.Assert().Error(err)
	suite.Assert().Equal(uint64(0), actualID)
}

func (suite *ManagerClientTestSuite) Test_CreateID_WhenServerSucceeds() {
	expectedID := uint64(100)
	testService := "test-service"
	testServiceID := "test-id"
	suite.mux.HandleFunc(getIDAPIPath+"users", func(w http.ResponseWriter, r *http.Request) {
		suite.Assert().Equal("POST", r.Method, "HTTP method should be POST")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"id":"%d", "account":{"service":"%s","id":"%s"}}`, expectedID, testService, testServiceID)
	})

	client, err := NewManagerClient(nil, suite.endpoint)
	suite.Require().NoError(err)

	actualID, err := client.CreateID(testService, testServiceID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedID, actualID, "TUID for the service user should be returned")
}

func (suite *ManagerClientTestSuite) Test_CreateID_WhenServerFails() {
	testService := "test-service"
	testServiceID := "test-id"
	suite.mux.HandleFunc(getIDAPIPath+testService+"/"+testServiceID, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	client, err := NewManagerClient(nil, suite.endpoint)
	suite.Require().NoError(err)

	actualID, err := client.CreateID(testService, testServiceID)
	suite.Assert().Error(err)
	suite.Assert().Equal(uint64(0), actualID)
}
