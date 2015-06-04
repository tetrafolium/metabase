package repository

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/tractrix/common-go/cache"
)

// TODO: Consider appropriate expiration for each cache.
const cacheExpiration = 24 * time.Hour

type references struct {
	Slug Slug
	Refs []*Reference
}
type commit struct {
	Slug Slug
	Ref  Reference
	ID   string
}

type serviceWithCacheBase struct {
	raw        Service
	cacheStore cache.Store
	cacheKey   string
}

func newServiceWithCacheBase(service Service, cacheStore cache.Store, cacheKey string) (*serviceWithCacheBase, error) {
	if service == nil {
		return nil, fmt.Errorf("no repository service specified")
	}
	if cacheStore == nil {
		return nil, fmt.Errorf("no cache store specified")
	}

	base := &serviceWithCacheBase{
		raw:        service,
		cacheStore: cacheStore,
		cacheKey:   fmt.Sprintf("repository-service/%s/%s", service.GetServiceName(), cacheKey),
	}

	return base, nil
}

func (service *serviceWithCacheBase) repositoryFromCacheItem(item *cache.Item) *Repository {
	if item == nil {
		return nil
	}

	repo := new(Repository)
	if err := json.Unmarshal(item.Value, repo); err != nil {
		return nil
	}

	return repo
}

func (service *serviceWithCacheBase) repositoriesFromCacheItem(item *cache.Item) []*Repository {
	if item == nil {
		return nil
	}

	repos := []*Repository{}
	if err := json.Unmarshal(item.Value, &repos); err != nil {
		return nil
	}

	return repos
}

func (service *serviceWithCacheBase) referencesFromCacheItem(item *cache.Item) *references {
	if item == nil {
		return nil
	}

	refs := new(references)
	if err := json.Unmarshal(item.Value, refs); err != nil {
		return nil
	}

	return refs
}

func (service *serviceWithCacheBase) commitFromCacheItem(item *cache.Item) *commit {
	if item == nil {
		return nil
	}

	c := new(commit)
	if err := json.Unmarshal(item.Value, c); err != nil {
		return nil
	}

	return c
}

func (service *serviceWithCacheBase) separateReferences(refs []*Reference) ([]*Reference, []*Reference) {
	var branches []*Reference
	var tags []*Reference

	for _, ref := range refs {
		switch {
		case ref.IsBranch():
			branches = append(branches, ref)
		case ref.IsTag():
			tags = append(tags, ref)
		}
	}

	return branches, tags
}

func (service *serviceWithCacheBase) retrieveReferencesCache(slug *Slug, key string) ([]*Reference, error) {
	item, err := service.cacheStore.Get(key)
	if err != nil {
		return nil, err
	}

	// The cache key might conflict with a cache key for an other item
	// since the cache key is generated based on hash value of the given repository slug.
	// Therefore it is necessary to check if the repository slug in cache data is equal
	// to the given one.
	refs := service.referencesFromCacheItem(item)
	if refs == nil || refs.Slug != *slug {
		return nil, cache.ErrCacheMiss
	}

	return refs.Refs, nil
}

func (service *serviceWithCacheBase) storeReferencesCache(slug *Slug, key string, refs []*Reference) {
	// The listed references information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	refsCache := &references{
		Slug: *slug,
		Refs: refs,
	}
	jsonItem := &cache.JSONItem{
		Key:        key,
		Value:      refsCache,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}
}

func (service *serviceWithCacheBase) branchesCacheKey(slug *Slug) string {
	return fmt.Sprintf("%s/ListBranches/%d", service.cacheKey, slug.Hash())
}

func (service *serviceWithCacheBase) retrieveBranchesCache(slug *Slug) ([]*Reference, error) {
	key := service.branchesCacheKey(slug)
	return service.retrieveReferencesCache(slug, key)
}

func (service *serviceWithCacheBase) storeBranchesCache(slug *Slug, branches []*Reference) {
	key := service.branchesCacheKey(slug)
	service.storeReferencesCache(slug, key, branches)
}

func (service *serviceWithCacheBase) tagsCacheKey(slug *Slug) string {
	return fmt.Sprintf("%s/ListTags/%d", service.cacheKey, slug.Hash())
}

func (service *serviceWithCacheBase) retrieveTagsCache(slug *Slug) ([]*Reference, error) {
	key := service.tagsCacheKey(slug)
	return service.retrieveReferencesCache(slug, key)
}

func (service *serviceWithCacheBase) storeTagsCache(slug *Slug, tags []*Reference) {
	key := service.tagsCacheKey(slug)
	service.storeReferencesCache(slug, key, tags)
}

// GetServiceName returns the repository service name.
func (service *serviceWithCacheBase) GetServiceName() string {
	return service.raw.GetServiceName()
}

// RegisterDeployKey registers the deploy key with the repository.
// Note that RegisterDeployKey never reads/writes any cache data.
func (service *serviceWithCacheBase) RegisterDeployKey(slug *Slug, publicKey, title string) (*DeployKey, error) {
	// There is nothing cached for this operation.
	return service.raw.RegisterDeployKey(slug, publicKey, title)
}

// UnregisterDeployKey unregisters the deploy key from the repository.
// Note that UnregisterDeployKey never reads/writes any cache data.
func (service *serviceWithCacheBase) UnregisterDeployKey(slug *Slug, deployKey *DeployKey) error {
	// There is nothing cached for this operation.
	return service.raw.UnregisterDeployKey(slug, deployKey)
}

// ExistsDeployKey returns true if the deploy key exists on the repository.
// It returns false otherwise.
// Note that ExistsDeployKey never reads/writes any cache data.
func (service *serviceWithCacheBase) ExistsDeployKey(slug *Slug, deployKey *DeployKey) (bool, error) {
	// There is nothing cached for this operation.
	return service.raw.ExistsDeployKey(slug, deployKey)
}

// RegisterWebhook registers the webhook URL with the repository.
// Note that RegisterWebhook never reads/writes any cache data.
func (service *serviceWithCacheBase) RegisterWebhook(slug *Slug, hookURL string) (*Webhook, error) {
	// There is nothing cached for this operation.
	return service.raw.RegisterWebhook(slug, hookURL)
}

// UnregisterWebhook unregisters the webhook from the repository.
// Note that UnregisterWebhook never reads/writes any cache data.
func (service *serviceWithCacheBase) UnregisterWebhook(slug *Slug, webhook *Webhook) error {
	// There is nothing cached for this operation.
	return service.raw.UnregisterWebhook(slug, webhook)
}

// ExistsWebhook returns true if the webhook exists on the repository.
// It returns false otherwise.
// Note that ExistsWebhook never reads/writes any cache data.
func (service *serviceWithCacheBase) ExistsWebhook(slug *Slug, webhook *Webhook) (bool, error) {
	// There is nothing cached for this operation.
	return service.raw.ExistsWebhook(slug, webhook)
}

// ServiceWithLatest always retrieves the latest data from the actual repository service
// and stores the data to the cache store.
// That means ServiceWithLatest never reads any data from the cache store, so that it takes
// longer time to complete some of operations than ServiceWithCache does.
type ServiceWithLatest struct {
	*serviceWithCacheBase
}

// NewServiceWithLatest returns a new NewServiceWithLatest instance.
func NewServiceWithLatest(service Service, cacheStore cache.Store, cacheKey string) (*ServiceWithLatest, error) {
	base, err := newServiceWithCacheBase(service, cacheStore, cacheKey)
	if err != nil {
		return nil, err
	}

	s := &ServiceWithLatest{
		serviceWithCacheBase: base,
	}

	return s, nil
}

// GetUserID returns the authenticated user ID.
// GetUserID always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) GetUserID() (string, error) {
	userID, err := service.raw.GetUserID()
	if err != nil {
		return "", err
	}

	// The user ID is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	service.cacheStore.Set(&cache.Item{
		Key:        fmt.Sprintf("%s/GetUserID", service.cacheKey),
		Value:      []byte(userID),
		Expiration: cacheExpiration,
	})

	return userID, nil
}

// GetRepositoryBySlug returns detailed information for a repository identified by the slug.
// GetRepositoryBySlug always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) GetRepositoryBySlug(slug *Slug) (*Repository, error) {
	repo, err := service.raw.GetRepositoryBySlug(slug)
	if err != nil {
		return nil, err
	}

	// The repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	jsonItem := &cache.JSONItem{
		Key:        fmt.Sprintf("%s/GetRepositoryBySlug/%d", service.cacheKey, slug.Hash()),
		Value:      repo,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return repo, nil
}

// GetRepositoryByID returns detailed information for a repository identified by the ID.
// GetRepositoryByID always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) GetRepositoryByID(id *ID) (*Repository, error) {
	repo, err := service.raw.GetRepositoryByID(id)
	if err != nil {
		return nil, err
	}

	// The repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	jsonItem := &cache.JSONItem{
		Key:        fmt.Sprintf("%s/GetRepositoryByID/%d", service.cacheKey, id.Hash()),
		Value:      repo,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return repo, nil
}

// ListRepositories retrieves all repositories which the authenticated user can access.
// ListRepositories always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) ListRepositories() ([]*Repository, error) {
	repos, err := service.raw.ListRepositories()
	if err != nil {
		return nil, err
	}

	// The listed repository information is cached with a best effort,
	// so that any errors regarding caching data can be ignored.
	jsonItem := &cache.JSONItem{
		Key:        fmt.Sprintf("%s/ListRepositories", service.cacheKey),
		Value:      repos,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return repos, nil
}

// ListReferences retrieves all branches and tags in the repository.
// ListReferences always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) ListReferences(slug *Slug) ([]*Reference, error) {
	refs, err := service.raw.ListReferences(slug)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	branches, tags := service.separateReferences(refs)
	service.storeBranchesCache(slug, branches)
	service.storeTagsCache(slug, tags)

	// err should be returnd for raising ErrNotFound
	return refs, err
}

// ListBranches retrieves all branches in the repository.
// ListBranches always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) ListBranches(slug *Slug) ([]*Reference, error) {
	branches, err := service.raw.ListBranches(slug)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	service.storeBranchesCache(slug, branches)

	// err should be returnd for raising ErrNotFound
	return branches, err
}

// ListTags retrieves all tags in the repository.
// ListTags always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) ListTags(slug *Slug) ([]*Reference, error) {
	tags, err := service.raw.ListTags(slug)
	if err != nil && err != ErrNotFound {
		return nil, err
	}

	service.storeTagsCache(slug, tags)

	// err should be returnd for raising ErrNotFound
	return tags, err
}

// GetCommitID returns the commit ID for the repository and reference.
// GetCommitID always accesses the actual repository service and stores the latest data
// to the cache store.
func (service *ServiceWithLatest) GetCommitID(slug *Slug, ref *Reference) (string, error) {
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
		Key:        fmt.Sprintf("%s/GetCommitID/%d/%d", service.cacheKey, slug.Hash(), ref.Hash()),
		Value:      commit,
		Expiration: cacheExpiration,
	}
	if item, err := jsonItem.Item(); err == nil {
		service.cacheStore.Set(item)
	}

	return commitID, nil
}

// ServiceWithCache provides capability of caching data for an existing repository service.
type ServiceWithCache struct {
	*serviceWithCacheBase

	latest *ServiceWithLatest
}

// NewServiceWithCache returns a wrapper of the repository service which utilizes cache store.
func NewServiceWithCache(service Service, cacheStore cache.Store, cacheKey string) (*ServiceWithCache, error) {
	latest, err := NewServiceWithLatest(service, cacheStore, cacheKey)
	if err != nil {
		return nil, err
	}

	s := &ServiceWithCache{
		serviceWithCacheBase: latest.serviceWithCacheBase,
		latest:               latest,
	}

	return s, nil
}

// Latest returns a ServiceWithLatest instance that always provides the latest data
// without reading cache data and stores the latest data to the cache store.
func (service *ServiceWithCache) Latest() *ServiceWithLatest {
	return service.latest
}

// GetUserID returns the authenticated user ID.
// GetUserID first reads the cache store and returns the cache data if it exists.
// Otherwise GetUserID tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) GetUserID() (string, error) {
	key := fmt.Sprintf("%s/GetUserID", service.cacheKey)
	if item, err := service.cacheStore.Get(key); err == nil {
		return string(item.Value), nil
	}

	return service.latest.GetUserID()
}

// GetRepositoryBySlug returns detailed information for a repository identified by the slug.
// GetRepositoryBySlug first reads the cache store and returns the cache data if it exists.
// Otherwise GetRepositoryBySlug tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) GetRepositoryBySlug(slug *Slug) (*Repository, error) {
	if slug == nil {
		return service.latest.GetRepositoryBySlug(slug)
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

	return service.latest.GetRepositoryBySlug(slug)
}

// GetRepositoryByID returns detailed information for a repository identified by the ID.
// GetRepositoryByID first reads the cache store and returns the cache data if it exists.
// Otherwise GetRepositoryByID tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) GetRepositoryByID(id *ID) (*Repository, error) {
	if id == nil {
		return service.latest.GetRepositoryByID(id)
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

	return service.latest.GetRepositoryByID(id)
}

// ListRepositories retrieves all repositories which the authenticated user can access.
// ListRepositories first reads the cache store and returns the cache data if it exists.
// Otherwise ListRepositories tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) ListRepositories() ([]*Repository, error) {
	key := fmt.Sprintf("%s/ListRepositories", service.cacheKey)
	if item, err := service.cacheStore.Get(key); err == nil {
		if repos := service.repositoriesFromCacheItem(item); repos != nil {
			return repos, nil
		}
	}

	return service.latest.ListRepositories()
}

// ListReferences retrieves all branches and tags in the repository.
// ListReferences first reads the cache store and returns the cache data if it exists.
// Otherwise ListReferences tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) ListReferences(slug *Slug) ([]*Reference, error) {
	if slug == nil {
		return service.latest.ListReferences(slug)
	}

	refs, err := service.listReferences(slug)
	if err != nil {
		return nil, err
	}
	if len(refs) == 0 {
		return nil, ErrNotFound
	}
	return refs, nil
}

func (service *ServiceWithCache) listReferences(slug *Slug) ([]*Reference, error) {
	branches, branchesCacheErr := service.retrieveBranchesCache(slug)
	tags, tagsCacheErr := service.retrieveTagsCache(slug)

	if branchesCacheErr == nil {
		// In the case that both branches and tags are cached,
		// the merged list of branches and tags is returned.
		if tagsCacheErr == nil {
			return append(branches, tags...), nil
		}

		// In the case that branches are cached, but tags are not cached,
		// only the latest tags are retrieved from the actual repository service
		// and then the merged list of branches and tags is returned.
		tags, err := service.latest.ListTags(slug)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
		return append(branches, tags...), nil
	}

	// In the case that branches are not cached, but tags are cached,
	// only the latest branches are retrieved from the actual repository service
	// and then the merged list of branches and tags is returned.
	if tagsCacheErr == nil {
		branches, err := service.latest.ListBranches(slug)
		if err != nil && err != ErrNotFound {
			return nil, err
		}
		return append(branches, tags...), nil
	}

	// In the case that both branches and tags are not cached,
	// the latest branches and tags are retrieved from the actual repository service
	// and then the merged list of branches and tags is returned.
	return service.latest.ListReferences(slug)
}

// ListBranches retrieves all branches in the repository.
// ListBranches first reads the cache store and returns the cache data if it exists.
// Otherwise ListBranches tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) ListBranches(slug *Slug) ([]*Reference, error) {
	if slug == nil {
		return service.latest.ListBranches(slug)
	}

	if branches, err := service.retrieveBranchesCache(slug); err == nil {
		return branches, nil
	}

	return service.latest.ListBranches(slug)
}

// ListTags retrieves all tags in the repository.
// ListTags first reads the cache store and returns the cache data if it exists.
// Otherwise ListTags tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) ListTags(slug *Slug) ([]*Reference, error) {
	if slug == nil {
		return service.latest.ListTags(slug)
	}

	if branches, err := service.retrieveTagsCache(slug); err == nil {
		return branches, nil
	}

	return service.latest.ListTags(slug)
}

// GetCommitID returns the commit ID for the repository and reference.
// GetCommitID first reads the cache store and returns the cache data if it exists.
// Otherwise GetCommitID tries to retrieve the latest data from the actual repository service.
func (service *ServiceWithCache) GetCommitID(slug *Slug, ref *Reference) (string, error) {
	if slug == nil || ref == nil {
		return service.latest.GetCommitID(slug, ref)
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

	return service.latest.GetCommitID(slug, ref)
}
