package repository

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/tractrix/common-go/cache"
)

// TODO: Consider appropriate expiration for each cache.
const cacheExpiration = 15 * time.Minute

type commit struct {
	Slug Slug
	Ref  Reference
	ID   string
}

// ServiceWithCache provides capability of caching data for an existing repository service.
type ServiceWithCache struct {
	raw        Service
	cacheStore cache.Store
	cacheKey   string
}

// NewServiceWithCache returns a wrapper of the repository service which utilizes cache store.
func NewServiceWithCache(service Service, cacheStore cache.Store, cacheKey string) (*ServiceWithCache, error) {
	if service == nil {
		return nil, fmt.Errorf("no repository service specified")
	}
	if cacheStore == nil {
		return nil, fmt.Errorf("no cache store specified")
	}

	serviceWithCache := &ServiceWithCache{
		raw:        service,
		cacheStore: cacheStore,
		cacheKey:   fmt.Sprintf("repository-service/%s/%s", service.GetServiceName(), cacheKey),
	}

	return serviceWithCache, nil
}

func (service *ServiceWithCache) repositoryFromCacheItem(item *cache.Item) *Repository {
	if item == nil {
		return nil
	}

	repo := new(Repository)
	if err := json.Unmarshal(item.Value, repo); err != nil {
		return nil
	}

	return repo
}

func (service *ServiceWithCache) repositoriesFromCacheItem(item *cache.Item) []*Repository {
	if item == nil {
		return nil
	}

	repos := []*Repository{}
	if err := json.Unmarshal(item.Value, &repos); err != nil {
		return nil
	}

	return repos
}

func (service *ServiceWithCache) commitFromCacheItem(item *cache.Item) *commit {
	if item == nil {
		return nil
	}

	c := new(commit)
	if err := json.Unmarshal(item.Value, c); err != nil {
		return nil
	}

	return c
}

// GetServiceName returns the repository service name.
func (service *ServiceWithCache) GetServiceName() string {
	return service.raw.GetServiceName()
}

// GetUserID returns the authenticated user ID set through SetUserID.
func (service *ServiceWithCache) GetUserID() (string, error) {
	key := fmt.Sprintf("%s/GetUserID", service.cacheKey)
	if item, err := service.cacheStore.Get(key); err == nil {
		return string(item.Value), nil
	}

	userID, err := service.raw.GetUserID()
	if err != nil {
		return "", err
	}

	// The user ID is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	service.cacheStore.Set(&cache.Item{
		Key:        key,
		Value:      []byte(userID),
		Expiration: cacheExpiration,
	})

	return userID, err
}

// GetRepositoryBySlug returns detailed information for a repository identified by the slug.
func (service *ServiceWithCache) GetRepositoryBySlug(slug *Slug) (*Repository, error) {
	if slug == nil {
		return service.raw.GetRepositoryBySlug(slug)
	}

	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", service.cacheKey, slug.Hash())
	if item, err := service.cacheStore.Get(key); err == nil {
		// The cache key might conflict with a cache key for an other item
		// since the cache key is generated based on hash value of the given repository slug.
		// Therefore it is necessary to check if the repository slug in cache data is equal
		// to the given one.
		repo := service.repositoryFromCacheItem(item)
		if repo != nil && repo.Slug == *slug {
			return repo, nil
		}
	}

	repo, err := service.raw.GetRepositoryBySlug(slug)
	if err != nil {
		return nil, err
	}

	// The repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	jsonItem := &cache.JSONItem{
		Key:        key,
		Value:      repo,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return repo, err
}

// GetRepositoryByID returns detailed information for a repository identified by the ID.
func (service *ServiceWithCache) GetRepositoryByID(id *ID) (*Repository, error) {
	if id == nil {
		return service.raw.GetRepositoryByID(id)
	}

	key := fmt.Sprintf("%s/GetRepositoryByID/%d", service.cacheKey, id.Hash())
	if item, err := service.cacheStore.Get(key); err == nil {
		// The cache key might conflict with a cache key for an other item
		// since the cache key is generated based on hash value of the given repository ID.
		// Therefore it is necessary to check if the repository ID in cache data is equal
		// to the given one.
		repo := service.repositoryFromCacheItem(item)
		if repo != nil && repo.ID == *id {
			return repo, nil
		}
	}

	repo, err := service.raw.GetRepositoryByID(id)
	if err != nil {
		return nil, err
	}

	// The repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	jsonItem := &cache.JSONItem{
		Key:        key,
		Value:      repo,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return repo, err
}

// ListRepositories retrieves all repositories which the authenticated user can access.
func (service *ServiceWithCache) ListRepositories() ([]*Repository, error) {
	key := fmt.Sprintf("%s/ListRepositories", service.cacheKey)
	log.Printf("cache key: %s", key)
	if item, err := service.cacheStore.Get(key); err == nil {
		if repos := service.repositoriesFromCacheItem(item); repos != nil {
			return repos, nil
		}
	}

	repos, err := service.raw.ListRepositories()
	if err != nil {
		return nil, err
	}

	// The listed repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	jsonItem := &cache.JSONItem{
		Key:        key,
		Value:      repos,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return repos, err
}

// GetCommitID returns the commit ID for the repository and reference.
func (service *ServiceWithCache) GetCommitID(slug *Slug, ref *Reference) (string, error) {
	if slug == nil || ref == nil {
		return service.raw.GetCommitID(slug, ref)
	}

	key := fmt.Sprintf("%s/GetCommitID/%d/%d", service.cacheKey, slug.Hash(), ref.Hash())
	if item, err := service.cacheStore.Get(key); err == nil {
		// The cache key might conflict with a cache key for an other item
		// since the cache key is generated based on hash values of the given
		// repository slug and reference.
		// Therefore it is necessary to check if the repository slug and reference
		// are equal to the given ones.
		commit := service.commitFromCacheItem(item)
		if commit != nil && commit.Slug == *slug && commit.Ref == *ref {
			return commit.ID, nil
		}
	}

	commitID, err := service.raw.GetCommitID(slug, ref)
	if err != nil {
		return "", err
	}

	// The repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	commit := &commit{
		Slug: *slug,
		Ref:  *ref,
		ID:   commitID,
	}
	jsonItem := &cache.JSONItem{
		Key:        key,
		Value:      commit,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return commitID, err
}

// RegisterDeployKey registers the deploy key with the repository.
func (service *ServiceWithCache) RegisterDeployKey(slug *Slug, publicKey, title string) (*DeployKey, error) {
	// There is nothing cached for this operation.
	return service.raw.RegisterDeployKey(slug, publicKey, title)
}

// UnregisterDeployKey unregisters the deploy key from the repository.
func (service *ServiceWithCache) UnregisterDeployKey(slug *Slug, deployKey *DeployKey) error {
	// There is nothing cached for this operation.
	return service.raw.UnregisterDeployKey(slug, deployKey)
}

// ExistsDeployKey returns true if the deploy key exists on the repository.
// It returns false otherwise.
func (service *ServiceWithCache) ExistsDeployKey(slug *Slug, deployKey *DeployKey) (bool, error) {
	// There is nothing cached for this operation.
	return service.raw.ExistsDeployKey(slug, deployKey)
}

// RegisterWebhook registers the webhook URL with the repository.
func (service *ServiceWithCache) RegisterWebhook(slug *Slug, hookURL string) (*Webhook, error) {
	// There is nothing cached for this operation.
	return service.raw.RegisterWebhook(slug, hookURL)
}

// UnregisterWebhook unregisters the webhook from the repository.
func (service *ServiceWithCache) UnregisterWebhook(slug *Slug, webhook *Webhook) error {
	// There is nothing cached for this operation.
	return service.raw.UnregisterWebhook(slug, webhook)
}

// ExistsWebhook returns true if the webhook exists on the repository.
// It returns false otherwise.
func (service *ServiceWithCache) ExistsWebhook(slug *Slug, webhook *Webhook) (bool, error) {
	// There is nothing cached for this operation.
	return service.raw.ExistsWebhook(slug, webhook)
}
