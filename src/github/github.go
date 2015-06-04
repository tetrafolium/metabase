package github

import (
	"net/http"

	"github.com/google/go-github/github"
	"golang.org/x/net/context"

	"github.com/tractrix/common-go/oauth"
	"github.com/tractrix/common-go/repository"
)

// ServiceName represents the service name for GitHub.
const ServiceName = "github.com"

// See https://developer.github.com/v3/#pagination
const maxPageSize = 100

// OAuthConfig represents oauth client application info
var OAuthConfig oauth.Config

// Service communicates with GitHub.
// It implements repository.Service interface.
type Service struct {
	client *github.Client
}

// NewService returns a new Service bound to the context.
func NewService(ctx context.Context, token oauth.Token) (*Service, error) {
	client, err := OAuthConfig.Client(ctx, token)
	if err != nil {
		return nil, err
	}

	return newServiceWithClient(client)
}

func newServiceWithClient(client *http.Client) (*Service, error) {
	service := &Service{
		client: github.NewClient(client),
	}
	return service, nil
}

func (service *Service) translateErrorResponse(resp *github.Response, err error) error {
	if resp != nil && resp.Response != nil {
		if resp.StatusCode == http.StatusNotFound {
			return repository.ErrNotFound
		}
	}

	return err
}

// GetServiceName returns the repository service name.
func (service *Service) GetServiceName() string {
	return ServiceName
}
