package repository

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"testing"

	mockfw "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/tractrix/common-go/cache"
	"github.com/tractrix/common-go/test/mock"
)

type RepositoryServiceMock struct {
	mockfw.Mock
}

func (m *RepositoryServiceMock) GetServiceName() string {
	return "RepositoryServiceMock"
}

func (m *RepositoryServiceMock) GetUserID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *RepositoryServiceMock) GetRepositoryBySlug(slug *Slug) (*Repository, error) {
	args := m.Called(slug)
	return args.Get(0).(*Repository), args.Error(1)
}

func (m *RepositoryServiceMock) GetRepositoryByID(id *ID) (*Repository, error) {
	args := m.Called(id)
	return args.Get(0).(*Repository), args.Error(1)
}

func (m *RepositoryServiceMock) ListRepositories() ([]*Repository, error) {
	args := m.Called()
	return args.Get(0).([]*Repository), args.Error(1)
}

func (m *RepositoryServiceMock) GetCommitID(slug *Slug, ref *Reference) (string, error) {
	args := m.Called(slug, ref)
	return args.String(0), args.Error(1)
}

func (m *RepositoryServiceMock) RegisterDeployKey(slug *Slug, publicKey, title string) (*DeployKey, error) {
	args := m.Called(slug, publicKey, title)
	return args.Get(0).(*DeployKey), args.Error(1)
}

func (m *RepositoryServiceMock) UnregisterDeployKey(slug *Slug, deployKey *DeployKey) error {
	args := m.Called(slug, deployKey)
	return args.Error(0)
}

func (m *RepositoryServiceMock) ExistsDeployKey(slug *Slug, deployKey *DeployKey) (bool, error) {
	args := m.Called(slug, deployKey)
	return args.Bool(0), args.Error(1)
}

func (m *RepositoryServiceMock) RegisterWebhook(slug *Slug, hookURL string) (*Webhook, error) {
	args := m.Called(slug, hookURL)
	return args.Get(0).(*Webhook), args.Error(1)
}

func (m *RepositoryServiceMock) UnregisterWebhook(slug *Slug, webhook *Webhook) error {
	args := m.Called(slug, webhook)
	return args.Error(0)
}

func (m *RepositoryServiceMock) ExistsWebhook(slug *Slug, webhook *Webhook) (bool, error) {
	args := m.Called(slug, webhook)
	return args.Bool(0), args.Error(1)
}

type ServiceWithCacheTestSuite struct {
	suite.Suite

	service    *ServiceWithCache
	rawService *RepositoryServiceMock
	cacheStore *mock.CacheStoreOnMemoryMock
	cacheKey   string
}

func Test_ServiceWithCacheTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceWithCacheTestSuite))
}

func (suite *ServiceWithCacheTestSuite) SetupTest() {
	suite.rawService = new(RepositoryServiceMock)
	suite.cacheStore = mock.NewCacheStoreOnMemoryMockWithStorage(mock.NewCacheStoreOnMemoryStorage())
	suite.cacheKey = strconv.FormatUint(math.MaxUint64, 10)

	suite.service, _ = NewServiceWithCache(suite.rawService, suite.cacheStore, suite.cacheKey)
}

func (suite *ServiceWithCacheTestSuite) TearDownTest() {
	suite.service = nil

	suite.rawService = nil
	suite.cacheStore = nil
	suite.cacheKey = ""
}

func (suite *ServiceWithCacheTestSuite) newTestRepository(name, id string, admin bool) *Repository {
	return &Repository{
		Slug: Slug{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  name,
		},
		ID: ID{
			Saas:    "github.com",
			OwnerID: "100",
			ID:      id,
		},
		Permissions: Permissions{
			Admin: admin,
		},
	}
}

func (suite *ServiceWithCacheTestSuite) newTestCommit(repoName, branchName, commitID string) *commit {
	return &commit{
		Slug: Slug{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  repoName,
		},
		Ref: Reference{
			Type: ReferenceTypeBranch,
			Name: branchName,
		},
		ID: commitID,
	}
}

func (suite *ServiceWithCacheTestSuite) Test_NewServiceWithCache_WithValidParameters() {
	actual, err := NewServiceWithCache(suite.rawService, suite.cacheStore, suite.cacheKey)
	suite.Assert().NoError(err)
	suite.Assert().Implements((*Service)(nil), actual, "ServiceWithCache should implement Service interface")
}

func (suite *ServiceWithCacheTestSuite) Test_NewServiceWithCache_WithNilService() {
	actual, err := NewServiceWithCache(nil, suite.cacheStore, suite.cacheKey)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_NewServiceWithCache_WithNilCacheStore() {
	actual, err := NewServiceWithCache(suite.rawService, nil, suite.cacheKey)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_NewServiceWithCache_WithEmptyCacheKey() {
	actual, err := NewServiceWithCache(suite.rawService, suite.cacheStore, "")
	suite.Assert().NoError(err)
	suite.Assert().Implements((*Service)(nil), actual, "specifying empty cache key should be allowd")
}

func (suite *ServiceWithCacheTestSuite) Test_repositoryFromCacheItem_WithValidItem() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	actual := suite.service.repositoryFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: expectedInBytes,
	})
	suite.Assert().Equal(expected, actual)
}

func (suite *ServiceWithCacheTestSuite) Test_repositoryFromCacheItem_WithInvalidItem() {
	actual := suite.service.repositoryFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_repositoryFromCacheItem_WithNilItem() {
	actual := suite.service.repositoryFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_repositoriesFromCacheItem_WithValidItem() {
	expected := []*Repository{
		suite.newTestRepository("common-go", "200", true),
		suite.newTestRepository("docstand", "201", false),
	}
	expectedInBytes, _ := json.Marshal(expected)
	actual := suite.service.repositoriesFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: expectedInBytes,
	})
	suite.Assert().Equal(expected, actual)
}

func (suite *ServiceWithCacheTestSuite) Test_repositoriesFromCacheItem_WithInvalidItem() {
	actual := suite.service.repositoriesFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_repositoriesFromCacheItem_WithNilItem() {
	actual := suite.service.repositoriesFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_commitFromCacheItem_WithValidItem() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	actual := suite.service.commitFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: expectedInBytes,
	})
	suite.Assert().Equal(expected, actual)
}

func (suite *ServiceWithCacheTestSuite) Test_commitFromCacheItem_WithInvalidItem() {
	actual := suite.service.commitFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_commitFromCacheItem_WithNilItem() {
	actual := suite.service.commitFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_GetUserID_WhenCacheMiss() {
	expected := "test user id"
	suite.rawService.On("GetUserID").Return(expected, nil).Once()

	actual, err := suite.service.GetUserID()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	key := fmt.Sprintf("%s/GetUserID", suite.service.cacheKey)
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, string(item.Value), "obtained user id should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetUserID_WhenCacheHit() {
	expected := "test user id"
	key := fmt.Sprintf("%s/GetUserID", suite.service.cacheKey)
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte(expected),
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetUserID()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.rawService.AssertNotCalled(suite.T(), "GetUserID")
}

func (suite *ServiceWithCacheTestSuite) Test_GetUserID_WhenRawServiceError() {
	suite.rawService.On("GetUserID").Return("", fmt.Errorf("test: unexpected error")).Once()

	actual, err := suite.service.GetUserID()
	suite.Assert().Error(err)
	suite.Assert().Equal("", actual)

	key := fmt.Sprintf("%s/GetUserID", suite.service.cacheKey)
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WithCacheMiss() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryBySlug", &expected.Slug).Return(expected, nil).Once()

	actual, err := suite.service.GetRepositoryBySlug(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", suite.service.cacheKey, expected.Slug.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WithValidCacheHit() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: expectedInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryBySlug(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.rawService.AssertNotCalled(suite.T(), "GetRepositoryBySlug")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WithCacheConflict() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryBySlug", &expected.Slug).Return(expected, nil).Once()

	conflict := *expected
	conflict.Name = "conflict-repo"
	conflictInBytes, _ := json.Marshal(conflict)
	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: conflictInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryBySlug(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WithInvalidCacheHit() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryBySlug", &expected.Slug).Return(expected, nil).Once()

	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte("non json value"),
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryBySlug(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WhenRawServiceError() {
	slug := new(Slug)
	suite.rawService.On("GetRepositoryBySlug", slug).Return((*Repository)(nil), fmt.Errorf("test: unexpected error")).Once()

	actual, err := suite.service.GetRepositoryBySlug(slug)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)

	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", suite.service.cacheKey, slug.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WithNilSlug() {
	suite.rawService.On("GetRepositoryBySlug", (*Slug)(nil)).Return((*Repository)(nil), fmt.Errorf("no slug specified")).Once()

	actual, err := suite.service.GetRepositoryBySlug(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WithCacheMiss() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryByID", &expected.ID).Return(expected, nil).Once()

	actual, err := suite.service.GetRepositoryByID(&expected.ID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	key := fmt.Sprintf("%s/GetRepositoryByID/%d", suite.service.cacheKey, expected.ID.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WithValidCacheHit() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	key := fmt.Sprintf("%s/GetRepositoryByID/%d", suite.service.cacheKey, expected.ID.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: expectedInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryByID(&expected.ID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.rawService.AssertNotCalled(suite.T(), "GetRepositoryByID")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WithCacheConflict() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryByID", &expected.ID).Return(expected, nil).Once()

	conflict := *expected
	conflict.ID.ID = "201"
	conflictInBytes, _ := json.Marshal(conflict)
	key := fmt.Sprintf("%s/GetRepositoryByID/%d", suite.service.cacheKey, expected.ID.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: conflictInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryByID(&expected.ID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WithInvalidCacheHit() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryByID", &expected.ID).Return(expected, nil).Once()

	key := fmt.Sprintf("%s/GetRepositoryByID/%d", suite.service.cacheKey, expected.ID.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte("non json value"),
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryByID(&expected.ID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WhenRawServiceError() {
	id := new(ID)
	suite.rawService.On("GetRepositoryByID", id).Return((*Repository)(nil), fmt.Errorf("test: unexpected error")).Once()

	actual, err := suite.service.GetRepositoryByID(id)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)

	key := fmt.Sprintf("%s/GetRepositoryByID/%d", suite.service.cacheKey, id.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WithNilSlug() {
	suite.rawService.On("GetRepositoryByID", (*ID)(nil)).Return((*Repository)(nil), fmt.Errorf("no id specified")).Once()

	actual, err := suite.service.GetRepositoryByID(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheTestSuite) Test_ListRepositories_WhenCacheMiss() {
	expected := []*Repository{
		suite.newTestRepository("common-go", "200", true),
		suite.newTestRepository("docstand", "201", false),
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListRepositories").Return(expected, nil).Once()

	actual, err := suite.service.ListRepositories()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	key := fmt.Sprintf("%s/ListRepositories", suite.service.cacheKey)
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListRepositories_WhenCacheHit() {
	expected := []*Repository{
		suite.newTestRepository("common-go", "200", true),
		suite.newTestRepository("docstand", "201", false),
	}
	expectedInBytes, _ := json.Marshal(expected)
	key := fmt.Sprintf("%s/ListRepositories", suite.service.cacheKey)
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: expectedInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.ListRepositories()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.rawService.AssertNotCalled(suite.T(), "ListRepositories")
}

func (suite *ServiceWithCacheTestSuite) Test_ListRepositories_WhenRawServiceError() {
	suite.rawService.On("ListRepositories").Return(([]*Repository)(nil), fmt.Errorf("test: unexpected error")).Once()

	actual, err := suite.service.ListRepositories()
	suite.Assert().Error(err)
	suite.Assert().Len(actual, 0)

	key := fmt.Sprintf("%s/ListRepositories", suite.service.cacheKey)
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithCacheMiss() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetCommitID", &expected.Slug, &expected.Ref).Return(expected.ID, nil).Once()

	actualID, err := suite.service.GetCommitID(&expected.Slug, &expected.Ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.ID, actualID)

	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, expected.Slug.Hash(), expected.Ref.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained commit id should be cached with slug and reference")
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithValidCacheHit() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, expected.Slug.Hash(), expected.Ref.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: expectedInBytes,
	})
	suite.Assert().NoError(err)

	actualID, err := suite.service.GetCommitID(&expected.Slug, &expected.Ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.ID, actualID)

	suite.rawService.AssertNotCalled(suite.T(), "GetCommitID")
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithCacheConflictOnSlug() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetCommitID", &expected.Slug, &expected.Ref).Return(expected.ID, nil).Once()

	conflict := *expected
	conflict.Slug.Name = "conflict-repo"
	conflictInBytes, _ := json.Marshal(conflict)
	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, expected.Slug.Hash(), expected.Ref.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: conflictInBytes,
	})
	suite.Assert().NoError(err)

	actualID, err := suite.service.GetCommitID(&expected.Slug, &expected.Ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.ID, actualID)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained commit id should be cached with slug and reference")
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithCacheConflictOnReference() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetCommitID", &expected.Slug, &expected.Ref).Return(expected.ID, nil).Once()

	conflict := *expected
	conflict.Ref.Name = "conflict-branch"
	conflictInBytes, _ := json.Marshal(conflict)
	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, expected.Slug.Hash(), expected.Ref.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: conflictInBytes,
	})
	suite.Assert().NoError(err)

	actualID, err := suite.service.GetCommitID(&expected.Slug, &expected.Ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.ID, actualID)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained commit id should be cached with slug and reference")
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithInvalidCacheHit() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetCommitID", &expected.Slug, &expected.Ref).Return(expected.ID, nil).Once()

	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, expected.Slug.Hash(), expected.Ref.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte("non json value"),
	})
	suite.Assert().NoError(err)

	actualID, err := suite.service.GetCommitID(&expected.Slug, &expected.Ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.ID, actualID)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained commit id should be cached with slug and reference")
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WhenRawServiceError() {
	slug := new(Slug)
	ref := new(Reference)
	suite.rawService.On("GetCommitID", slug, ref).Return("", fmt.Errorf("test: unexpected error")).Once()

	actualID, err := suite.service.GetCommitID(slug, ref)
	suite.Assert().Error(err)
	suite.Assert().Equal("", actualID)

	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, slug.Hash(), ref.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithNilSlug() {
	ref := new(Reference)
	suite.rawService.On("GetCommitID", (*Slug)(nil), ref).Return("", fmt.Errorf("no slug specified")).Once()

	actualID, err := suite.service.GetCommitID(nil, ref)
	suite.Assert().Error(err)
	suite.Assert().Equal("", actualID)
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WithNilReference() {
	slug := new(Slug)
	suite.rawService.On("GetCommitID", slug, (*Reference)(nil)).Return("", fmt.Errorf("no reference specified")).Once()

	actualID, err := suite.service.GetCommitID(slug, nil)
	suite.Assert().Error(err)
	suite.Assert().Equal("", actualID)
}

func (suite *ServiceWithCacheTestSuite) Test_RegisterDeployKey() {
	slug := new(Slug)
	publicKey := "ssh public key"
	title := "test deploy key"
	expected := &DeployKey{
		ID: "300",
	}
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("RegisterDeployKey", slug, publicKey, title).Return(expected, nil).Once()

	actual, err := suite.service.RegisterDeployKey(slug, publicKey, title)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_UnregisterDeployKey() {
	slug := new(Slug)
	deployKey := new(DeployKey)
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("UnregisterDeployKey", slug, deployKey).Return(nil).Once()

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().NoError(err)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ExistsDeployKey() {
	slug := new(Slug)
	deployKey := new(DeployKey)
	expected := true
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("ExistsDeployKey", slug, deployKey).Return(expected, nil).Once()

	actual, err := suite.service.ExistsDeployKey(slug, deployKey)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_RegisterWebhook() {
	slug := new(Slug)
	hookURL := "https://docstand.tractrix.io/github/webhook"
	expected := &Webhook{
		ID: "300",
	}
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("RegisterWebhook", slug, hookURL).Return(expected, nil).Once()

	actual, err := suite.service.RegisterWebhook(slug, hookURL)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_UnregisterWebhook() {
	slug := new(Slug)
	webhooK := new(Webhook)
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("UnregisterWebhook", slug, webhooK).Return(nil).Once()

	err := suite.service.UnregisterWebhook(slug, webhooK)
	suite.Assert().NoError(err)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ExistsWebhook() {
	slug := new(Slug)
	webhooK := new(Webhook)
	expected := true
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("ExistsWebhook", slug, webhooK).Return(expected, nil).Once()

	actual, err := suite.service.ExistsWebhook(slug, webhooK)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}
