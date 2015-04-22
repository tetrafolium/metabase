package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"golang.org/x/net/context"

	e "github.com/tractrix/common-go/error"
	cmnfactory "github.com/tractrix/common-go/factory"
	"github.com/tractrix/common-go/queue"
	cmnmock "github.com/tractrix/common-go/test/mock"
	"github.com/tractrix/docstand/server/data"
	"github.com/tractrix/docstand/server/factory"
	"github.com/tractrix/docstand/server/test/mock"
)

type APIHandlerTestSuite struct {
	suite.Suite

	prevContextFactoryFunc       cmnfactory.ContextFactoryFunc
	prevPushTaskQueueFactoryFunc cmnfactory.PushTaskQueueFactoryFunc
	prevDocStoreFactoryFunc      factory.DocumentationStoreFactoryFunc
	prevRepoStoreFactoryFunc     factory.RepositoryStoreFactoryFunc
}

func Test_APIHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(APIHandlerTestSuite))
}

func (suite *APIHandlerTestSuite) SetupTest() {
	suite.prevContextFactoryFunc = cmnfactory.ReplaceContextFactoryFunc(suite.newContextNopMock)
	suite.prevPushTaskQueueFactoryFunc = cmnfactory.ReplacePushTaskQueueFactoryFunc(suite.newPushTaskQueueNopMock)
	suite.prevDocStoreFactoryFunc = factory.ReplaceDocumentationStoreFactoryFunc(suite.newDocumentationStoreOnMemoryMock)
	suite.prevRepoStoreFactoryFunc = factory.ReplaceRepositoryStoreFactoryFunc(suite.newRepositoryStoreOnMemoryMock)
}

func (suite *APIHandlerTestSuite) TearDownTest() {
	factory.ReplaceRepositoryStoreFactoryFunc(suite.prevRepoStoreFactoryFunc)
	suite.prevRepoStoreFactoryFunc = nil
	factory.ReplaceDocumentationStoreFactoryFunc(suite.prevDocStoreFactoryFunc)
	suite.prevDocStoreFactoryFunc = nil
	cmnfactory.ReplacePushTaskQueueFactoryFunc(suite.prevPushTaskQueueFactoryFunc)
	suite.prevPushTaskQueueFactoryFunc = nil
	cmnfactory.ReplaceContextFactoryFunc(suite.prevContextFactoryFunc)
	suite.prevContextFactoryFunc = nil
}

func (suite *APIHandlerTestSuite) newContextNopMock(_ *http.Request) context.Context {
	return cmnmock.NewContextNopMock()
}

func (suite *APIHandlerTestSuite) newPushTaskQueueNopMock(_ context.Context, name string) (queue.PushTaskQueue, error) {
	return cmnmock.NewPushTaskQueueNopMock(name), nil
}

func (suite *APIHandlerTestSuite) newDocumentationStoreOnMemoryMock(_ context.Context) (data.DocumentationStore, error) {
	storage := mock.NewDocumentationStoreOnMemoryStorage()
	return mock.NewDocumentationStoreOnMemoryMockWithStorage(storage), nil
}

func (suite *APIHandlerTestSuite) newRepositoryStoreOnMemoryMock(_ context.Context) (data.RepositoryStore, error) {
	storage := mock.NewRepositoryStoreOnMemoryStorage()
	return mock.NewRepositoryStoreOnMemoryMockWithStorage(storage), nil
}

func (suite *APIHandlerTestSuite) newHTTPRequest(method string, body interface{}) *http.Request {
	if body == nil {
		req, err := http.NewRequest(method, "http://dummy-host.com", nil)
		suite.Assert().NoError(err)
		return req
	}

	var bodyBuffer bytes.Buffer
	if err := json.NewEncoder(&bodyBuffer).Encode(body); err != nil {
		suite.Assert().NoError(err)
		return nil
	}

	req, err := http.NewRequest(method, "http://dummy-host.com", &bodyBuffer)
	suite.Assert().NoError(err)
	return req
}

func (suite *APIHandlerTestSuite) Test_NewAPIHandler_WithValidParameters() {
	actual := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	suite.Assert().NotNil(actual)
}

func (suite *APIHandlerTestSuite) Test_NewAPIHandler_WithNilResponseWriter() {
	suite.Assert().Panics(func() {
		NewAPIHandler(nil, suite.newHTTPRequest("GET", nil))
	})
}

func (suite *APIHandlerTestSuite) Test_NewAPIHandler_WithNilRequest() {
	suite.Assert().Panics(func() {
		NewAPIHandler(httptest.NewRecorder(), nil)
	})
}

func (suite *APIHandlerTestSuite) Test_ResponseWriter() {
	expectedWriter := httptest.NewRecorder()
	actualHandler := NewAPIHandler(expectedWriter, suite.newHTTPRequest("GET", nil))
	suite.Assert().NotNil(actualHandler)
	suite.Assert().True(expectedWriter == actualHandler.ResponseWriter(), "exactly same response writer should be returned")
}

func (suite *APIHandlerTestSuite) Test_Request() {
	expectedReq := suite.newHTTPRequest("GET", nil)
	actualHandler := NewAPIHandler(httptest.NewRecorder(), expectedReq)
	suite.Assert().NotNil(actualHandler)
	suite.Assert().True(expectedReq == actualHandler.Request(), "exactly same request should be returned")
}

func (suite *APIHandlerTestSuite) Test_Context_WhenFirstTime() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	actual := handler.Context()
	suite.Assert().NotNil(actual)
	suite.Assert().IsType(new(cmnmock.ContextNopMock), actual, "context should be instantiated by factory function")
}

func (suite *APIHandlerTestSuite) Test_Context_WhenMultipleTimes() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	expected := handler.Context()
	actual := handler.Context()
	suite.Assert().NotNil(actual)
	suite.Assert().True(expected == actual, "exactly same context should be returned")
}

func (suite *APIHandlerTestSuite) Test_PushTaskQueue_WhenFirstTime() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	actual := handler.PushTaskQueue("test-queue")
	suite.Assert().NotNil(actual)
	suite.Assert().IsType(new(cmnmock.PushTaskQueueNopMock), actual, "push task queue should be instantiated by factory function")
}

func (suite *APIHandlerTestSuite) Test_PushTaskQueue_WhenMultipleTimes() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	queueName := "test-queue"
	expected := handler.PushTaskQueue(queueName)
	actual := handler.PushTaskQueue(queueName)
	suite.Assert().NotNil(actual)
	suite.Assert().True(expected == actual, "exactly same push task queue should be returned")
}

func (suite *APIHandlerTestSuite) Test_PushTaskQueue_WithDifferentQueues() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	actual1 := handler.PushTaskQueue("test-queue1")
	actual2 := handler.PushTaskQueue("test-queue2")
	suite.Assert().NotNil(actual1)
	suite.Assert().NotNil(actual2)
	suite.Assert().True(actual1 != actual2, "different push task queue should be returned")
}

func (suite *APIHandlerTestSuite) Test_PushTaskQueue_WhenFactoryError() {
	cmnfactory.ReplacePushTaskQueueFactoryFunc(func(_ context.Context, name string) (queue.PushTaskQueue, error) {
		return nil, fmt.Errorf("test: intentional error")
	})
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	suite.Assert().Panics(func() {
		handler.PushTaskQueue("test-queue")
	})
}

func (suite *APIHandlerTestSuite) Test_DocumentationStore_WhenFirstTime() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	actual := handler.DocumentationStore()
	suite.Assert().NotNil(actual)
	suite.Assert().IsType(new(mock.DocumentationStoreOnMemoryMock), actual, "documentation store should be instantiated by factory function")
}

func (suite *APIHandlerTestSuite) Test_DocumentationStore_WhenMultipleTimes() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	expected := handler.DocumentationStore()
	actual := handler.DocumentationStore()
	suite.Assert().NotNil(actual)
	suite.Assert().True(expected == actual, "exactly same documentation store should be returned")
}

func (suite *APIHandlerTestSuite) Test_DocumentationStore_WhenFactoryError() {
	factory.ReplaceDocumentationStoreFactoryFunc(func(_ context.Context) (data.DocumentationStore, error) {
		return nil, fmt.Errorf("test: intentional error")
	})
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	suite.Assert().Panics(func() {
		handler.DocumentationStore()
	})
}

func (suite *APIHandlerTestSuite) Test_RepositoryStore_WhenFirstTime() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	actual := handler.RepositoryStore()
	suite.Assert().NotNil(actual)
	suite.Assert().IsType(new(mock.RepositoryStoreOnMemoryMock), actual, "repository store should be instantiated by factory function")
}

func (suite *APIHandlerTestSuite) Test_RepositoryStore_WhenMultipleTimes() {
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	expected := handler.RepositoryStore()
	actual := handler.RepositoryStore()
	suite.Assert().NotNil(actual)
	suite.Assert().True(expected == actual, "exactly same repository store should be returned")
}

func (suite *APIHandlerTestSuite) Test_RepositoryStore_WhenFactoryError() {
	factory.ReplaceRepositoryStoreFactoryFunc(func(_ context.Context) (data.RepositoryStore, error) {
		return nil, fmt.Errorf("test: intentional error")
	})
	handler := NewAPIHandler(httptest.NewRecorder(), suite.newHTTPRequest("GET", nil))
	suite.Assert().Panics(func() {
		handler.RepositoryStore()
	})
}

func (suite *APIHandlerTestSuite) Test_UnmarshalRequestBodyAsJSON_WithValidBody() {
	expectedBody := map[string]interface{}{
		"param": "value",
	}
	req := suite.newHTTPRequest("POST", expectedBody)
	handler := NewAPIHandler(httptest.NewRecorder(), req)

	actualBody := make(map[string]interface{})
	err := handler.UnmarshalRequestBodyAsJSON(&actualBody)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBody, actualBody)
}

func (suite *APIHandlerTestSuite) Test_UnmarshalRequestBodyAsJSON_WhenUnmarshalFails() {
	req := suite.newHTTPRequest("POST", "")
	handler := NewAPIHandler(httptest.NewRecorder(), req)

	err := handler.UnmarshalRequestBodyAsJSON(nil)
	suite.Assert().Error(err)
}

func (suite *APIHandlerTestSuite) Test_UnmarshalRequestBodyAsJSON_WhenNilRequestBody() {
	req := suite.newHTTPRequest("GET", nil)
	handler := NewAPIHandler(httptest.NewRecorder(), req)

	actualBody := make(map[string]interface{})
	err := handler.UnmarshalRequestBodyAsJSON(&actualBody)
	suite.Assert().NoError(err)
	suite.Assert().Empty(actualBody)
}

func (suite *APIHandlerTestSuite) Test_SendError() {
	w := httptest.NewRecorder()
	handler := NewAPIHandler(w, suite.newHTTPRequest("GET", nil))
	expectedStatus := http.StatusBadRequest
	expectedError := e.Content{
		Code:    e.ErrorInvalidJSON,
		Message: "test: an error has occurred",
	}
	expectedContent := &e.Response{
		Errors: []e.Content{expectedError},
	}
	handler.SendError(expectedStatus, expectedError.Code, expectedError.Message)
	suite.Assert().Equal(expectedStatus, w.Code)

	actualContent := new(e.Response)
	err := json.NewDecoder(w.Body).Decode(actualContent)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedContent, actualContent)
}

func (suite *APIHandlerTestSuite) Test_SendResponseWithNoContent() {
	w := httptest.NewRecorder()
	handler := NewAPIHandler(w, suite.newHTTPRequest("GET", nil))
	expectedStatus := http.StatusOK
	handler.SendResponseWithNoContent(expectedStatus)
	suite.Assert().Equal(expectedStatus, w.Code)
	suite.Assert().Equal(0, w.Body.Len())
}

func (suite *APIHandlerTestSuite) Test_SendResponseWithContent() {
	w := httptest.NewRecorder()
	handler := NewAPIHandler(w, suite.newHTTPRequest("GET", nil))
	expectedStatus := http.StatusOK
	expectedContentType := "text/plain"
	expectedContentBytes := []byte("ok")
	handler.SendResponseWithContent(expectedStatus, expectedContentType, expectedContentBytes)
	suite.Assert().Equal(expectedStatus, w.Code)

	actualContentBytes, err := ioutil.ReadAll(w.Body)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedContentBytes, actualContentBytes)
}

func (suite *APIHandlerTestSuite) Test_SendResponseWithJSONContent_WithValidParameters() {
	w := httptest.NewRecorder()
	handler := NewAPIHandler(w, suite.newHTTPRequest("GET", nil))
	expectedStatus := http.StatusOK
	expectedContent := map[string]interface{}{
		"test": "success",
	}
	handler.SendResponseWithJSONContent(expectedStatus, expectedContent)
	suite.Assert().Equal(expectedStatus, w.Code)
	suite.Assert().Equal("application/json", w.Header().Get("Content-Type"))

	actualContent := make(map[string]interface{})
	err := json.NewDecoder(w.Body).Decode(&actualContent)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedContent, actualContent)
}

func (suite *APIHandlerTestSuite) Test_SendResponseWithJSONContent_WithNilBody() {
	w := httptest.NewRecorder()
	handler := NewAPIHandler(w, suite.newHTTPRequest("GET", nil))
	expectedStatus := http.StatusOK
	handler.SendResponseWithJSONContent(expectedStatus, nil)
	suite.Assert().Equal(expectedStatus, w.Code)

	actualContent, err := ioutil.ReadAll(w.Body)
	suite.Assert().NoError(err)
	suite.Assert().Equal("null", string(actualContent))
}

func (suite *APIHandlerTestSuite) Test_SendResponseWithJSONContent_WithInvalidBody() {
	w := httptest.NewRecorder()
	handler := NewAPIHandler(w, suite.newHTTPRequest("GET", nil))
	code := http.StatusOK
	invalidBody := func() {}
	handler.SendResponseWithJSONContent(code, invalidBody)
	suite.Assert().Equal(http.StatusInternalServerError, w.Code)

	actualContent := new(e.Response)
	err := json.NewDecoder(w.Body).Decode(actualContent)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualContent.Errors, 1)
	suite.Assert().Equal(e.ErrorInvalidJSON, actualContent.Errors[0].Code)
	suite.Assert().NotEmpty(actualContent.Errors[0].Message)
}
