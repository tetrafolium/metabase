package repositorytest

import (
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/tractrix/common-go/repository"
)

// Repository represetns a virtual repository for testing.
type Repository struct {
	mutex      sync.RWMutex
	lastID     int64
	slug       repository.Slug
	id         repository.ID
	userIDs    map[string]bool                 // Key: User ID    , Value: Administrator Flag
	branches   map[string]string               // Key: Branch Name, Value: HEAD Commit ID
	tags       map[string]string               // Key: Tag Name   , Value: Commit ID
	deployKeys map[string]repository.DeployKey // Key: Deploy Key ID
	webhooks   map[string]repository.Webhook   // Key: Webhook ID
}

// NewRepository returns a new Repository.
func NewRepository(slug *repository.Slug, id *repository.ID) *Repository {
	return &Repository{
		slug:       *slug,
		id:         *id,
		userIDs:    make(map[string]bool),
		branches:   make(map[string]string),
		tags:       make(map[string]string),
		deployKeys: make(map[string]repository.DeployKey),
		webhooks:   make(map[string]repository.Webhook),
	}
}

// Slug returns the repository slug.
func (repo *Repository) Slug() *repository.Slug {
	slug := repo.slug
	return &slug
}

// ID returns the repository ID.
func (repo *Repository) ID() *repository.ID {
	id := repo.id
	return &id
}

// Detail returns the detailed repository information for the specified user.
func (repo *Repository) Detail(userID string) *repository.Repository {
	return &repository.Repository{
		Slug: repo.slug,
		ID:   repo.id,
		Permissions: repository.Permissions{
			Admin: repo.IsAdminUser(userID),
		},
	}
}

// RegisterUser registers the specified user with the repository.
func (repo *Repository) RegisterUser(userID string, admin bool) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.userIDs[userID] = admin
}

// IsUser returns true if the specified user is a user of the repository,
// or returns false otherwise.
func (repo *Repository) IsUser(userID string) bool {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	_, ok := repo.userIDs[userID]
	return ok
}

// IsAdminUser returns true if the specified user is an administrator user
// of the repository, or returns false otherwise.
func (repo *Repository) IsAdminUser(userID string) bool {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	return repo.userIDs[userID]
}

// RegisterReference registers the specified reference with the repository.
// If the reference already exists, then RegisterBranch replaces the current one
// with the new one.
func (repo *Repository) RegisterReference(ref *repository.Reference, commitID string) {
	if ref != nil {
		switch {
		case ref.IsBranch():
			repo.RegisterBranch(ref.Name, commitID)
		case ref.IsTag():
			repo.RegisterTag(ref.Name, commitID)
		}
	}
}

// References returns a list of references in the repository.
func (repo *Repository) References() []*repository.Reference {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	var refs []*repository.Reference
	for name := range repo.branches {
		refs = append(refs, &repository.Reference{
			Type: repository.ReferenceTypeBranch,
			Name: name,
		})
	}
	for name := range repo.tags {
		refs = append(refs, &repository.Reference{
			Type: repository.ReferenceTypeTag,
			Name: name,
		})
	}
	return refs
}

// HasReference returns true if the repository has the specified reference,
// or returns false otherwise.
func (repo *Repository) HasReference(ref *repository.Reference) bool {
	if ref != nil {
		for _, existingRef := range repo.References() {
			if *ref == *existingRef {
				return true
			}
		}
	}
	return false
}

// ReferenceCommitIDs returns a map of references and their HEAD commit IDs in the repository.
func (repo *Repository) ReferenceCommitIDs() map[repository.Reference]string {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	commitIDs := make(map[repository.Reference]string)
	for name, commitID := range repo.branches {
		ref := repository.Reference{
			Type: repository.ReferenceTypeBranch,
			Name: name,
		}
		commitIDs[ref] = commitID
	}
	for name, commitID := range repo.tags {
		ref := repository.Reference{
			Type: repository.ReferenceTypeTag,
			Name: name,
		}
		commitIDs[ref] = commitID
	}
	return commitIDs
}

// RegisterBranch registers the specified branch with the repository.
// If the branch already exists, then RegisterBranch replaces the current one
// with the new one.
func (repo *Repository) RegisterBranch(name, headCommitID string) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.branches[name] = headCommitID
}

// Branches returns a list of branch names in the repository.
func (repo *Repository) Branches() []string {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	var names []string
	for name := range repo.branches {
		names = append(names, name)
	}
	return names
}

// BranchCommitIDs returns a map of branches and their HEAD commit IDs in the repository.
func (repo *Repository) BranchCommitIDs() map[string]string {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	commitIDs := make(map[string]string)
	for name, commitID := range repo.branches {
		commitIDs[name] = commitID
	}
	return commitIDs
}

// RegisterTag registers the specified tag with the repository.
// If the tag already exists, then RegisterTag replaces the current one
// with the new one.
func (repo *Repository) RegisterTag(name, commitID string) {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	repo.tags[name] = commitID
}

// Tags returns a list of tag names in the repository.
func (repo *Repository) Tags() []string {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	var names []string
	for name := range repo.tags {
		names = append(names, name)
	}
	return names
}

// TagCommitIDs returns a map of tags and their commit IDs in the repository.
func (repo *Repository) TagCommitIDs() map[string]string {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	tags := make(map[string]string)
	for name, commitID := range repo.tags {
		tags[name] = commitID
	}
	return tags
}

// NewDeployKey creates a new deploy key for the repository.
func (repo *Repository) NewDeployKey() *repository.DeployKey {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	id := atomic.AddInt64(&repo.lastID, 1)
	deployKey := repository.DeployKey{
		ID: strconv.FormatInt(id, 10),
	}
	repo.deployKeys[deployKey.ID] = deployKey
	return &deployKey
}

// DeleteDeployKey deletes the deploy key from the repository.
func (repo *Repository) DeleteDeployKey(deployKey *repository.DeployKey) error {
	if deployKey == nil {
		return nil
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.deployKeys, deployKey.ID)
	return nil
}

// DeployKey returns a deploy having the specified ID.
func (repo *Repository) DeployKey(id string) *repository.DeployKey {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if deployKey, ok := repo.deployKeys[id]; ok {
		return &deployKey
	}
	return nil
}

// NewWebhook creates a new deploy key for the repository.
func (repo *Repository) NewWebhook() *repository.Webhook {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	id := atomic.AddInt64(&repo.lastID, 1)
	webhook := repository.Webhook{
		ID: strconv.FormatInt(id, 10),
	}
	repo.webhooks[webhook.ID] = webhook
	return &webhook
}

// DeleteWebhook deletes the deploy key from the repository.
func (repo *Repository) DeleteWebhook(webhook *repository.Webhook) error {
	if webhook == nil {
		return nil
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	delete(repo.webhooks, webhook.ID)
	return nil
}

// Webhook returns a deploy having the specified ID.
func (repo *Repository) Webhook(id string) *repository.Webhook {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if webhook, ok := repo.webhooks[id]; ok {
		return &webhook
	}
	return nil
}

// Store represents a store of virtual repositories for testing.
type Store struct {
	mutex sync.RWMutex
	repos map[repository.Slug]*Repository
}

var defaultStore = NewStore()

// DefaultStore returns the default virtual repository store shared across all Service instances.
func DefaultStore() *Store {
	return defaultStore
}

// NewStore returns a new Store.
func NewStore() *Store {
	return &Store{
		repos: make(map[repository.Slug]*Repository),
	}
}

// ResetResources resets all resources held in this package.
func ResetResources() {
	defaultStore.Clear()
}

// Clear deletes all virtual repositories from the store.
func (store *Store) Clear() {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	store.repos = make(map[repository.Slug]*Repository)
}

// RegisterRepository registers the specified virtual repository with the store.
// If the repository already exists, then RegisterRepository replaces the current one
// with the new one.
func (store *Store) RegisterRepository(repo *Repository) {
	if repo != nil {
		store.mutex.Lock()
		defer store.mutex.Unlock()

		store.repos[*repo.Slug()] = repo
	}
}

// UnregisterRepository removes the specified virtual repository from the store.
func (store *Store) UnregisterRepository(repo *Repository) {
	if repo != nil {
		store.mutex.Lock()
		defer store.mutex.Unlock()

		delete(store.repos, *repo.Slug())
	}
}

// RepositoryBySlug returns a virtual repository identified by the specified repository slug.
func (store *Store) RepositoryBySlug(slug *repository.Slug) *Repository {
	if slug == nil {
		return nil
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	return store.repos[*slug]
}

// RepositoryByID returns a virtual repository identified by the specified repository ID.
func (store *Store) RepositoryByID(id *repository.ID) *Repository {
	if id == nil {
		return nil
	}

	store.mutex.RLock()
	defer store.mutex.RUnlock()

	for _, repo := range store.repos {
		if *id == *repo.ID() {
			return repo
		}
	}

	return nil
}

// Repositories returns all virtual repositories added to the store.
func (store *Store) Repositories() []*Repository {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	var repos []*Repository
	for _, repo := range store.repos {
		repos = append(repos, repo)
	}

	return repos
}

// ServiceDelegate holds functions where Service delegates to.
type ServiceDelegate struct {
	GetServiceName func() string

	GetUserID func() (string, error)

	GetRepositoryBySlug func(slug *repository.Slug) (*repository.Repository, error)
	GetRepositoryByID   func(id *repository.ID) (*repository.Repository, error)
	ListRepositories    func() ([]*repository.Repository, error)

	ListReferences func(slug *repository.Slug) ([]*repository.Reference, error)
	ListBranches   func(slug *repository.Slug) ([]*repository.Reference, error)
	ListTags       func(slug *repository.Slug) ([]*repository.Reference, error)
	GetCommitID    func(slug *repository.Slug, ref *repository.Reference) (string, error)

	RegisterDeployKey   func(slug *repository.Slug, publicKey, title string) (*repository.DeployKey, error)
	UnregisterDeployKey func(slug *repository.Slug, deployKey *repository.DeployKey) error
	ExistsDeployKey     func(slug *repository.Slug, deployKey *repository.DeployKey) (bool, error)

	RegisterWebhook   func(slug *repository.Slug, hookURL string) (*repository.Webhook, error)
	UnregisterWebhook func(slug *repository.Slug, webhook *repository.Webhook) error
	ExistsWebhook     func(slug *repository.Slug, webhook *repository.Webhook) (bool, error)
}

// Service is a mock of repository service for testing.
// Service delegates all invocations to Delegate, so that the behavior of Service
// can be customized by updating Delegate.
type Service struct {
	mutex    sync.RWMutex
	userID   string
	store    *Store
	Delegate ServiceDelegate
}

// ServiceName represents the service name for the mock of repository service.
const ServiceName = "mock"

// NewService returns a new mock object of repository service with an empty user ID.
func NewService() *Service {
	return NewServiceWithUserID("")
}

// NewServiceWithUserID returns a new mock object of repository service with the given user ID.
func NewServiceWithUserID(userID string) *Service {
	service := &Service{
		userID: userID,
		store:  defaultStore,
	}
	service.Delegate = ServiceDelegate{
		GetServiceName: service.defaultGetServiceName,

		GetUserID: service.defaultGetUserID,

		GetRepositoryBySlug: service.defaultGetRepositoryBySlug,
		GetRepositoryByID:   service.defaultGetRepositoryByID,
		ListRepositories:    service.defaultListRepositories,

		ListReferences: service.defaultListReferences,
		ListBranches:   service.defaultListBranches,
		ListTags:       service.defaultListTags,
		GetCommitID:    service.defaultGetCommitID,

		RegisterDeployKey:   service.defaultRegisterDeployKey,
		UnregisterDeployKey: service.defaultUnregisterDeployKey,
		ExistsDeployKey:     service.defaultExistsDeployKey,

		RegisterWebhook:   service.defaultRegisterWebhook,
		UnregisterWebhook: service.defaultUnregisterWebhook,
		ExistsWebhook:     service.defaultExistsWebhook,
	}
	return service
}

func (service *Service) userRepositoryBySlug(slug *repository.Slug) (string, *Repository) {
	userID, _ := service.GetUserID()
	if repo := service.store.RepositoryBySlug(slug); repo != nil && repo.IsUser(userID) {
		return userID, repo
	}
	return userID, nil
}

func (service *Service) userRepositoryByID(id *repository.ID) (string, *Repository) {
	userID, _ := service.GetUserID()
	if repo := service.store.RepositoryByID(id); repo != nil && repo.IsUser(userID) {
		return userID, repo
	}
	return userID, nil
}

func (service *Service) userRepositories() (string, []*Repository) {
	userID, _ := service.GetUserID()
	var repos []*Repository
	for _, repo := range service.store.Repositories() {
		if repo.IsUser(userID) {
			repos = append(repos, repo)
		}
	}
	return userID, repos
}

// Store returns the repository information store for testing.
func (service *Service) Store() *Store {
	return service.store
}

// GetServiceName invokes Delegate.GetServiceName.
func (service *Service) GetServiceName() string {
	return service.Delegate.GetServiceName()
}

// defaultGetServiceName returns the service name of the repository service.
func (service *Service) defaultGetServiceName() string {
	return ServiceName
}

// GetUserID invokes Delegate.GetUserID.
func (service *Service) GetUserID() (string, error) {
	return service.Delegate.GetUserID()
}

// defaultGetUserID returns the authenticated user ID set through SetUserID.
func (service *Service) defaultGetUserID() (string, error) {
	service.mutex.RLock()
	defer service.mutex.RUnlock()

	return service.userID, nil
}

// SetUserID sets the given user to the authenticated user for testing.
func (service *Service) SetUserID(userID string) {
	service.mutex.Lock()
	defer service.mutex.Unlock()

	service.userID = userID
}

// GetRepositoryBySlug invokes Delegate.GetRepositoryBySlug.
func (service *Service) GetRepositoryBySlug(slug *repository.Slug) (*repository.Repository, error) {
	return service.Delegate.GetRepositoryBySlug(slug)
}

// defaultGetRepositoryBySlug returns detailed information for a repository identified by the slug.
func (service *Service) defaultGetRepositoryBySlug(slug *repository.Slug) (*repository.Repository, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return nil, err
	}

	userID, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return nil, repository.ErrNotFound
	}

	return &repository.Repository{
		Slug: *repo.Slug(),
		ID:   *repo.ID(),
		Permissions: repository.Permissions{
			Admin: repo.IsAdminUser(userID),
		},
	}, nil
}

// GetRepositoryByID invokes Delegate.GetRepositoryByID.
func (service *Service) GetRepositoryByID(id *repository.ID) (*repository.Repository, error) {
	return service.Delegate.GetRepositoryByID(id)
}

// defaultGetRepositoryByID returns detailed information for a repository identified by the ID.
func (service *Service) defaultGetRepositoryByID(id *repository.ID) (*repository.Repository, error) {
	if err := repository.ValidateID(id, service.GetServiceName()); err != nil {
		return nil, err
	}

	userID, repo := service.userRepositoryByID(id)
	if repo == nil {
		return nil, repository.ErrNotFound
	}

	return &repository.Repository{
		Slug: *repo.Slug(),
		ID:   *repo.ID(),
		Permissions: repository.Permissions{
			Admin: repo.IsAdminUser(userID),
		},
	}, nil
}

// ListRepositories invokes Delegate.ListRepositories.
func (service *Service) ListRepositories() ([]*repository.Repository, error) {
	return service.Delegate.ListRepositories()
}

// defaultListRepositories retrieves all repositories which the authenticated user can access.
func (service *Service) defaultListRepositories() ([]*repository.Repository, error) {
	userID, repos := service.userRepositories()
	var results []*repository.Repository
	for _, repo := range repos {
		results = append(results, &repository.Repository{
			Slug: *repo.Slug(),
			ID:   *repo.ID(),
			Permissions: repository.Permissions{
				Admin: repo.IsAdminUser(userID),
			},
		})
	}

	return results, nil
}

// ListReferences invokes Delegate.ListReferences.
func (service *Service) ListReferences(slug *repository.Slug) ([]*repository.Reference, error) {
	return service.Delegate.ListReferences(slug)
}

// defaultListReferences retrieves all branches and tags in the repository.
func (service *Service) defaultListReferences(slug *repository.Slug) ([]*repository.Reference, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return nil, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return nil, nil
	}

	refs := repo.References()
	if len(refs) == 0 {
		return nil, repository.ErrNotFound
	}
	return refs, nil
}

// ListBranches invokes Delegate.ListBranches.
func (service *Service) ListBranches(slug *repository.Slug) ([]*repository.Reference, error) {
	return service.Delegate.ListBranches(slug)
}

// defaultListBranches retrieves all branches in the repository.
func (service *Service) defaultListBranches(slug *repository.Slug) ([]*repository.Reference, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return nil, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return nil, nil
	}

	var branches []*repository.Reference
	for _, name := range repo.Branches() {
		branches = append(branches, &repository.Reference{
			Type: repository.ReferenceTypeBranch,
			Name: name,
		})
	}
	if len(branches) == 0 {
		return nil, repository.ErrNotFound
	}
	return branches, nil
}

// ListTags invokes Delegate.ListTags.
func (service *Service) ListTags(slug *repository.Slug) ([]*repository.Reference, error) {
	return service.Delegate.ListTags(slug)
}

// defaultListTags retrieves all tags in the repository.
func (service *Service) defaultListTags(slug *repository.Slug) ([]*repository.Reference, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return nil, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return nil, nil
	}

	var tags []*repository.Reference
	for _, name := range repo.Tags() {
		tags = append(tags, &repository.Reference{
			Type: repository.ReferenceTypeTag,
			Name: name,
		})
	}
	if len(tags) == 0 {
		return nil, repository.ErrNotFound
	}
	return tags, nil
}

// GetCommitID invokes Delegate.GetCommitID.
func (service *Service) GetCommitID(slug *repository.Slug, ref *repository.Reference) (string, error) {
	return service.Delegate.GetCommitID(slug, ref)
}

// defaultGetCommitID returns the commit ID for the repository and reference.
func (service *Service) defaultGetCommitID(slug *repository.Slug, ref *repository.Reference) (string, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return "", err
	}
	if err := repository.ValidateReference(ref); err != nil {
		return "", err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return "", repository.ErrNotFound
	}

	var commitIDs map[string]string
	switch {
	case ref.IsBranch():
		commitIDs = repo.BranchCommitIDs()
	case ref.IsTag():
		commitIDs = repo.TagCommitIDs()
	}
	if commitID, ok := commitIDs[ref.Name]; ok {
		return commitID, nil
	}

	return "", fmt.Errorf("repositorytest: no such reference")
}

// RegisterDeployKey invokes Delegate.RegisterDeployKey.
func (service *Service) RegisterDeployKey(slug *repository.Slug, publicKey, title string) (*repository.DeployKey, error) {
	return service.Delegate.RegisterDeployKey(slug, publicKey, title)
}

// defaultRegisterDeployKey registers the deploy key with the repository.
func (service *Service) defaultRegisterDeployKey(slug *repository.Slug, publicKey, title string) (*repository.DeployKey, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return nil, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return nil, repository.ErrNotFound
	}

	return repo.NewDeployKey(), nil
}

// UnregisterDeployKey invokes Delegate.UnregisterDeployKey.
func (service *Service) UnregisterDeployKey(slug *repository.Slug, deployKey *repository.DeployKey) error {
	return service.Delegate.UnregisterDeployKey(slug, deployKey)
}

// defaultUnregisterDeployKey unregisters the deploy key from the repository.
func (service *Service) defaultUnregisterDeployKey(slug *repository.Slug, deployKey *repository.DeployKey) error {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return err
	}
	if err := repository.ValidateDeployKey(deployKey); err != nil {
		return err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return repository.ErrNotFound
	}

	return repo.DeleteDeployKey(deployKey)
}

// ExistsDeployKey invokes Delegate.ExistsDeployKey.
func (service *Service) ExistsDeployKey(slug *repository.Slug, deployKey *repository.DeployKey) (bool, error) {
	return service.Delegate.ExistsDeployKey(slug, deployKey)
}

// defaultExistsDeployKey returns true if the deploy key exists on the repository.
// It returns false otherwise.
func (service *Service) defaultExistsDeployKey(slug *repository.Slug, deployKey *repository.DeployKey) (bool, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return false, err
	}
	if err := repository.ValidateDeployKey(deployKey); err != nil {
		return false, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return false, repository.ErrNotFound
	}

	return repo.DeployKey(deployKey.ID) != nil, nil
}

// RegisterWebhook invokes Delegate.RegisterWebhook.
func (service *Service) RegisterWebhook(slug *repository.Slug, hookURL string) (*repository.Webhook, error) {
	return service.Delegate.RegisterWebhook(slug, hookURL)
}

// defaultRegisterWebhook registers the webhook URL with the repository.
func (service *Service) defaultRegisterWebhook(slug *repository.Slug, hookURL string) (*repository.Webhook, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return nil, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return nil, repository.ErrNotFound
	}

	return repo.NewWebhook(), nil
}

// UnregisterWebhook invokes Delegate.UnregisterWebhook.
func (service *Service) UnregisterWebhook(slug *repository.Slug, webhook *repository.Webhook) error {
	return service.Delegate.UnregisterWebhook(slug, webhook)
}

// defaultUnregisterWebhook unregisters the webhook from the repository.
func (service *Service) defaultUnregisterWebhook(slug *repository.Slug, webhook *repository.Webhook) error {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return err
	}
	if err := repository.ValidateWebhook(webhook); err != nil {
		return err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return repository.ErrNotFound
	}

	return repo.DeleteWebhook(webhook)
}

// ExistsWebhook invokes Delegate.ExistsWebhook.
func (service *Service) ExistsWebhook(slug *repository.Slug, webhook *repository.Webhook) (bool, error) {
	return service.Delegate.ExistsWebhook(slug, webhook)
}

// defaultExistsWebhook returns true if the webhook exists on the repository.
// It returns false otherwise.
func (service *Service) defaultExistsWebhook(slug *repository.Slug, webhook *repository.Webhook) (bool, error) {
	if err := repository.ValidateSlug(slug, service.GetServiceName()); err != nil {
		return false, err
	}
	if err := repository.ValidateWebhook(webhook); err != nil {
		return false, err
	}

	_, repo := service.userRepositoryBySlug(slug)
	if repo == nil {
		return false, repository.ErrNotFound
	}

	return repo.Webhook(webhook.ID) != nil, nil
}
