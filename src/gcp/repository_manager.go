package gcp

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/urlfetch"

	"github.com/tractrix/common-go/client/repository"
	"github.com/tractrix/common-go/gcp"
)

const (
	docstandProjectID       = "docstand-prod"
	repositoryManagerModule = "repository-manager"
	tractrixCoreProjectID   = "planar-oasis-88707"
	localTractrixCoreDomain = "localhost:9003"

	repositoryManagerTimeout = 60 * time.Second
)

// NewRepositoryManagerClient returns a new client for repository-manager module running in the Google App Engine.
func NewRepositoryManagerClient(ctx context.Context) *repository.ManagerClient {
	// repository manager is running in tractrix core project
	hostName := fmt.Sprintf("%s-dot-%s.appspot.com", repositoryManagerModule, tractrixCoreProjectID)
	apiURL := &url.URL{
		Scheme: "https",
		Host:   hostName,
		Path:   "/api/1/",
	}

	deadline := time.Now().Add(repositoryManagerTimeout)
	ctxWithDeadline, _ := context.WithDeadline(ctx, deadline)
	client := &http.Client{
		Transport: &urlfetch.Transport{
			Context: ctxWithDeadline,
		},
	}

	// NOTE: During inter-module communication on development app server,
	//       login information is not shared with the destination module.
	//       Actually, the login information is attached to HTTP request
	//       in the current context as a cookie, so it can be provided to
	//       the destination module also as a cookie as a workaround.
	if appengine.IsDevAppServer() {
		apiURL.Scheme = "http"
		apiURL.Host = localTractrixCoreDomain

		// Error will not occur because valid endpoint is always passed
		client.Jar, _ = gcp.NewCookieJarWithDummyAdminUserInfo(apiURL)
	}

	return repository.NewManagerClient(client, apiURL, docstandProjectID)
}
