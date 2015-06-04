package github

import (
	"fmt"
	"strconv"

	"github.com/google/go-github/github"

	"github.com/tractrix/common-go/repository"
)

func (service *Service) newRepositorySlug(repo *github.Repository) *repository.Slug {
	slug := &repository.Slug{
		Saas: ServiceName,
	}

	if repo != nil {
		if repo.Name != nil {
			slug.Name = *repo.Name
		}

		switch {
		case repo.Organization != nil && repo.Organization.Login != nil:
			slug.Owner = *repo.Organization.Login
		case repo.Owner != nil && repo.Owner.Login != nil:
			slug.Owner = *repo.Owner.Login
		}
	}

	return slug
}

func (service *Service) newRepositoryID(repo *github.Repository) *repository.ID {
	id := &repository.ID{
		Saas: ServiceName,
	}

	if repo != nil {
		if repo.ID != nil {
			id.ID = strconv.Itoa(*repo.ID)
		}

		switch {
		case repo.Organization != nil && repo.Organization.ID != nil:
			id.OwnerID = strconv.Itoa(*repo.Organization.ID)
		case repo.Owner != nil && repo.Owner.ID != nil:
			id.OwnerID = strconv.Itoa(*repo.Owner.ID)
		}
	}

	return id
}

func (service *Service) newRepositoryPermissions(repo *github.Repository) *repository.Permissions {
	permissions := new(repository.Permissions)

	if repo != nil && repo.Permissions != nil {
		githubPerms := *repo.Permissions
		permissions.Admin = githubPerms["admin"]
	}

	return permissions
}

func (service *Service) newRepository(repo *github.Repository) *repository.Repository {
	return &repository.Repository{
		Slug:        *service.newRepositorySlug(repo),
		ID:          *service.newRepositoryID(repo),
		Permissions: *service.newRepositoryPermissions(repo),
	}
}

func (service *Service) validateRepositoryID(id *repository.ID) error {
	if err := repository.ValidateID(id, ServiceName); err != nil {
		return err
	}
	if _, err := strconv.Atoi(id.OwnerID); err != nil {
		return fmt.Errorf("invalid repository owner id: %s", err.Error())
	}
	if _, err := strconv.Atoi(id.ID); err != nil {
		return fmt.Errorf("invalid repository id: %s", err.Error())
	}
	return nil
}

func (service *Service) listOrganizations(user string) ([]github.Organization, error) {
	var totalOrgs []github.Organization
	var opt github.ListOptions
	for {
		orgs, resp, err := service.client.Organizations.List(user, &opt)
		if err != nil {
			return nil, service.translateErrorResponse(resp, err)
		}
		totalOrgs = append(totalOrgs, orgs...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return totalOrgs, nil
}

func (service *Service) listRepositoriesByOrganization(org string) ([]github.Repository, error) {
	var totalRepos []github.Repository
	var opt github.RepositoryListByOrgOptions
	for {
		repos, resp, err := service.client.Repositories.ListByOrg(org, &opt)
		if err != nil {
			return nil, service.translateErrorResponse(resp, err)
		}
		totalRepos = append(totalRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return totalRepos, nil
}

func (service *Service) listRepositoriesByUser(user string) ([]github.Repository, error) {
	var totalRepos []github.Repository
	var opt github.RepositoryListOptions
	for {
		repos, resp, err := service.client.Repositories.List(user, &opt)
		if err != nil {
			return nil, service.translateErrorResponse(resp, err)
		}
		totalRepos = append(totalRepos, repos...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return totalRepos, nil
}

// GetRepositoryBySlug returns detailed information for a repository identified by the slug.
func (service *Service) GetRepositoryBySlug(slug *repository.Slug) (*repository.Repository, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return nil, err
	}

	githubRepo, resp, err := service.client.Repositories.Get(slug.Owner, slug.Name)
	if err != nil {
		return nil, service.translateErrorResponse(resp, err)
	}

	return service.newRepository(githubRepo), nil
}

// GetRepositoryByID returns detailed information for a repository identified by the ID.
func (service *Service) GetRepositoryByID(id *repository.ID) (*repository.Repository, error) {
	if err := service.validateRepositoryID(id); err != nil {
		return nil, err
	}

	// NOTE: /repositories/:id API does not seem to be supported officially,
	//       since it is not documented anywhere and google/go-github does not
	//       have the ability to call the API.
	//       But we need to call it for making it easy to retrieve repository
	//       information by the ID, otherwise we should list all repositories
	//       and find the expected repository by iterating over them.
	req, err := service.client.NewRequest("GET", fmt.Sprintf("repositories/%s", id.ID), nil)
	if err != nil {
		return nil, err
	}

	githubRepo := new(github.Repository)
	if resp, err := service.client.Do(req, githubRepo); err != nil {
		return nil, service.translateErrorResponse(resp, err)
	}

	repo := service.newRepository(githubRepo)
	if repo.ID != *id {
		return nil, fmt.Errorf("inconsistent repository id")
	}

	return repo, nil
}

// ListRepositories retrieves all repositories which the authenticated user can access.
func (service *Service) ListRepositories() ([]*repository.Repository, error) {
	user := "" // Authenticated user
	orgs, err := service.listOrganizations(user)
	if err != nil {
		return nil, err
	}

	reposChan := make(chan []github.Repository, len(orgs)+1)
	errChan := make(chan error)

	// Retrieve repositories for organizations and user in parallel
	for _, org := range orgs {
		go func(orgName string) {
			repos, err := service.listRepositoriesByOrganization(orgName)
			if err != nil {
				errChan <- err
				return
			}
			reposChan <- repos
		}(*org.Login)
	}
	go func() {
		repos, err := service.listRepositoriesByUser(user)
		if err != nil {
			errChan <- err
			return
		}
		reposChan <- repos
	}()

	var repos []*repository.Repository
	for i := 0; i < cap(reposChan); i++ {
		select {
		case githubRepos := <-reposChan:
			for _, githubRepo := range githubRepos {
				repos = append(repos, service.newRepository(&githubRepo))
			}
		case err := <-errChan:
			return nil, err
		}
	}

	return repos, nil
}
