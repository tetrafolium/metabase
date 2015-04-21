package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"sync"

	"golang.org/x/net/context"

	e "github.com/tractrix/common-go/error"
	cmnfactory "github.com/tractrix/common-go/factory"
	"github.com/tractrix/docstand/server/data"
	"github.com/tractrix/docstand/server/factory"
)

// APIHandler provides fundamental functionality for handling a web API requests.
type APIHandler struct {
	w                http.ResponseWriter
	req              *http.Request
	ctx              context.Context
	ctxFactory       sync.Once
	docStore         data.DocumentationStore
	docStoreFactory  sync.Once
	repoStore        data.RepositoryStore
	repoStoreFactory sync.Once
}

// NewAPIHandler returns a new APIHandler for the given HTTP request and response writer.
// It panics if the specified HTTP request or response writer is nil.
func NewAPIHandler(w http.ResponseWriter, req *http.Request) *APIHandler {
	if w == nil {
		// TODO: Cause panic by using logging library.
		panic(errors.New("handler: no response writer specified"))
	}
	if req == nil {
		// TODO: Cause panic by using logging library.
		panic(errors.New("handler: no in-flight http request specified"))
	}

	return &APIHandler{
		w:   w,
		req: req,
	}
}

// ResponseWriter returns the HTTP response writer associated with the handler.
func (handler *APIHandler) ResponseWriter() http.ResponseWriter {
	return handler.w
}

// Request returns the in-flight HTTP request associated with the handler.
func (handler *APIHandler) Request() *http.Request {
	return handler.req
}

// Context returns a context associated with the in-flight HTTP request.
func (handler *APIHandler) Context() context.Context {
	handler.ctxFactory.Do(func() {
		handler.ctx = cmnfactory.NewContext(handler.req)
	})

	return handler.ctx
}

// DocumentationStore returns a documentation store working in the context.
func (handler *APIHandler) DocumentationStore() data.DocumentationStore {
	handler.docStoreFactory.Do(func() {
		docStore, err := factory.NewDocumentationStore(handler.Context())
		if err != nil {
			// TODO: Cause panic by using logging library.
			panic(err)
		}
		handler.docStore = docStore
	})

	return handler.docStore
}

// RepositoryStore returns a repository store working in the context.
func (handler *APIHandler) RepositoryStore() data.RepositoryStore {
	handler.repoStoreFactory.Do(func() {
		repoStore, err := factory.NewRepositoryStore(handler.Context())
		if err != nil {
			// TODO: Cause panic by using logging library.
			panic(err)
		}
		handler.repoStore = repoStore
	})

	return handler.repoStore
}

// UnmarshalRequestBodyAsJSON decodes the request body as a JSON object.
func (handler *APIHandler) UnmarshalRequestBodyAsJSON(body interface{}) error {
	if handler.req.Body == nil {
		return nil
	}

	if err := json.NewDecoder(handler.req.Body).Decode(body); err != nil {
		return err
	}

	return nil
}

// SendError writes the error status, code and message to the HTTP response.
func (handler *APIHandler) SendError(status, code int, message string) {
	e.SendSingleError(handler.w, status, code, message)
}

// SendResponseWithNoContent writes the status to the HTTP response.
func (handler *APIHandler) SendResponseWithNoContent(status int) {
	handler.w.WriteHeader(status)
}

// SendResponseWithContent writes the status and content to the HTTP response.
// The contentType is set to "Content-Type" header.
func (handler *APIHandler) SendResponseWithContent(status int, contentType string, contentBytes []byte) {
	handler.w.Header().Set("Content-Type", contentType)
	handler.w.WriteHeader(status)
	handler.w.Write(contentBytes)
}

// SendResponseWithJSONContent writes the status and JSON content to the HTTP response.
// It writes an error response by calling SendError if the given content cannot be
// encoded as JSON object.
func (handler *APIHandler) SendResponseWithJSONContent(status int, content interface{}) {
	bytes, err := json.Marshal(content)
	if err != nil {
		handler.SendError(http.StatusInternalServerError, e.ErrorInvalidJSON, err.Error())
	} else {
		handler.SendResponseWithContent(status, "application/json", bytes)
	}
}
