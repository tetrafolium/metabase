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
	"github.com/tractrix/common-go/cache/cachetest"
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

func (m *RepositoryServiceMock) ListReferences(slug *Slug) ([]*Reference, error) {
	args := m.Called(slug)
	return args.Get(0).([]*Reference), args.Error(1)
}

func (m *RepositoryServiceMock) ListBranches(slug *Slug) ([]*Reference, error) {
	args := m.Called(slug)
	return args.Get(0).([]*Reference), args.Error(1)
}

func (m *RepositoryServiceMock) ListTags(slug *Slug) ([]*Reference, error) {
	args := m.Called(slug)
	return args.Get(0).([]*Reference), args.Error(1)
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

type ServiceWithCacheTestSuiteBase struct {
	suite.Suite

	rawService *RepositoryServiceMock
	cacheStore *cachetest.Store
	cacheKey   string
}

func (suite *ServiceWithCacheTestSuiteBase) SetupTest() {
	suite.rawService = new(RepositoryServiceMock)
	suite.cacheStore = cachetest.NewStoreWithStorage(cachetest.NewStorage())
	suite.cacheKey = strconv.FormatUint(math.MaxUint64, 10)
}

func (suite *ServiceWithCacheTestSuiteBase) TearDownTest() {
	suite.rawService = nil
	suite.cacheStore = nil
	suite.cacheKey = ""
}

func (suite *ServiceWithCacheTestSuiteBase) newTestSlug(name string) *Slug {
	return &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  name,
	}
}

func (suite *ServiceWithCacheTestSuiteBase) newTestRepository(name, id string, admin bool) *Repository {
	return &Repository{
		Slug: *suite.newTestSlug(name),
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

func (suite *ServiceWithCacheTestSuiteBase) newTestBranch(name string) *Reference {
	return &Reference{
		Type: ReferenceTypeBranch,
		Name: name,
	}
}

func (suite *ServiceWithCacheTestSuiteBase) newTestTag(name string) *Reference {
	return &Reference{
		Type: ReferenceTypeTag,
		Name: name,
	}
}

func (suite *ServiceWithCacheTestSuiteBase) newTestCommit(repoName, branchName, commitID string) *commit {
	return &commit{
		Slug: Slug{
			Saas:  "github.com",
			Owner: "tractrix",
			Name:  repoName,
		},
		Ref: *suite.newTestBranch(branchName),
		ID:  commitID,
	}
}

type ServiceWithCacheBaseTestSuite struct {
	ServiceWithCacheTestSuiteBase

	service *serviceWithCacheBase
}

func Test_ServiceWithCacheBaseTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceWithCacheBaseTestSuite))
}

func (suite *ServiceWithCacheBaseTestSuite) SetupTest() {
	suite.ServiceWithCacheTestSuiteBase.SetupTest()

	suite.service = &serviceWithCacheBase{
		raw:        suite.rawService,
		cacheStore: suite.cacheStore,
		cacheKey:   suite.cacheKey,
	}
}

func (suite *ServiceWithCacheBaseTestSuite) TearDownTest() {
	suite.service = nil

	suite.ServiceWithCacheTestSuiteBase.TearDownTest()
}

func (suite *ServiceWithCacheBaseTestSuite) Test_newServiceWithCacheBase_WithValidParameters() {
	actual, err := newServiceWithCacheBase(suite.rawService, suite.cacheStore, suite.cacheKey)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(actual)
	suite.Assert().True(suite.rawService == actual.raw)
	suite.Assert().True(suite.cacheStore == actual.cacheStore)
	suite.Assert().Equal(fmt.Sprintf("repository-service/%s/%s", suite.rawService.GetServiceName(), suite.cacheKey), actual.cacheKey)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_newServiceWithCacheBaseWithNilService() {
	actual, err := newServiceWithCacheBase(nil, suite.cacheStore, suite.cacheKey)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_newServiceWithCacheBase_WithNilCacheStore() {
	actual, err := newServiceWithCacheBase(suite.rawService, nil, suite.cacheKey)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_newServiceWithCacheBase_WithEmptyCacheKey() {
	actual, err := newServiceWithCacheBase(suite.rawService, suite.cacheStore, "")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(actual)
	suite.Assert().True(suite.rawService == actual.raw)
	suite.Assert().True(suite.cacheStore == actual.cacheStore)
	suite.Assert().Equal(fmt.Sprintf("repository-service/%s/", suite.rawService.GetServiceName()), actual.cacheKey)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_repositoryFromCacheItem_WithValidItem() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	actual := suite.service.repositoryFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: expectedInBytes,
	})
	suite.Assert().Equal(expected, actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_repositoryFromCacheItem_WithInvalidItem() {
	actual := suite.service.repositoryFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_repositoryFromCacheItem_WithNilItem() {
	actual := suite.service.repositoryFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_repositoriesFromCacheItem_WithValidItem() {
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

func (suite *ServiceWithCacheBaseTestSuite) Test_repositoriesFromCacheItem_WithInvalidItem() {
	actual := suite.service.repositoriesFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_repositoriesFromCacheItem_WithNilItem() {
	actual := suite.service.repositoriesFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_referencesFromCacheItem_WithValidItem() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	actual := suite.service.referencesFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: expectedInBytes,
	})
	suite.Assert().Equal(expected, actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_referencesFromCacheItem_WithInvalidItem() {
	actual := suite.service.referencesFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_referencesFromCacheItem_WithNilItem() {
	actual := suite.service.referencesFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_commitFromCacheItem_WithValidItem() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	actual := suite.service.commitFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: expectedInBytes,
	})
	suite.Assert().Equal(expected, actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_commitFromCacheItem_WithInvalidItem() {
	actual := suite.service.commitFromCacheItem(&cache.Item{
		Key:   "test key",
		Value: []byte("non json value"),
	})
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_commitFromCacheItem_WithNilItem() {
	actual := suite.service.commitFromCacheItem(nil)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithCacheBaseTestSuite) Test_GetServiceName() {
	suite.Assert().Equal(suite.rawService.GetServiceName(), suite.service.GetServiceName())
}

func (suite *ServiceWithCacheBaseTestSuite) Test_RegisterDeployKey() {
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

func (suite *ServiceWithCacheBaseTestSuite) Test_UnregisterDeployKey() {
	slug := new(Slug)
	deployKey := new(DeployKey)
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("UnregisterDeployKey", slug, deployKey).Return(nil).Once()

	err := suite.service.UnregisterDeployKey(slug, deployKey)
	suite.Assert().NoError(err)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheBaseTestSuite) Test_ExistsDeployKey() {
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

func (suite *ServiceWithCacheBaseTestSuite) Test_RegisterWebhook() {
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

func (suite *ServiceWithCacheBaseTestSuite) Test_UnregisterWebhook() {
	slug := new(Slug)
	webhooK := new(Webhook)
	expectedCacheSize := suite.cacheStore.Storage().Size()
	suite.rawService.On("UnregisterWebhook", slug, webhooK).Return(nil).Once()

	err := suite.service.UnregisterWebhook(slug, webhooK)
	suite.Assert().NoError(err)

	suite.Assert().Equal(expectedCacheSize, suite.cacheStore.Storage().Size(), "no data should be cached")
}

func (suite *ServiceWithCacheBaseTestSuite) Test_ExistsWebhook() {
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

type ServiceWithLatestTestSuite struct {
	ServiceWithCacheTestSuiteBase

	service *ServiceWithLatest
}

func Test_ServiceWithLatestTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceWithLatestTestSuite))
}

func (suite *ServiceWithLatestTestSuite) SetupTest() {
	suite.ServiceWithCacheTestSuiteBase.SetupTest()

	suite.service, _ = NewServiceWithLatest(suite.rawService, suite.cacheStore, suite.cacheKey)
}

func (suite *ServiceWithLatestTestSuite) TearDownTest() {
	suite.service = nil

	suite.ServiceWithCacheTestSuiteBase.TearDownTest()
}

func (suite *ServiceWithLatestTestSuite) Test_NewServiceWithLatest_WithValidParameters() {
	actual, err := NewServiceWithLatest(suite.rawService, suite.cacheStore, suite.cacheKey)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(actual)
	suite.Assert().Implements((*Service)(nil), actual, "ServiceWithLatest should implement Service interface")

	expectedBase, _ := newServiceWithCacheBase(suite.rawService, suite.cacheStore, suite.cacheKey)
	suite.Assert().Equal(expectedBase, actual.serviceWithCacheBase)
}

func (suite *ServiceWithLatestTestSuite) Test_NewServiceWithLatest_WithNilService() {
	actual, err := NewServiceWithLatest(nil, suite.cacheStore, suite.cacheKey)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithLatestTestSuite) Test_NewServiceWithLatest_WithNilCacheStore() {
	actual, err := NewServiceWithLatest(suite.rawService, nil, suite.cacheKey)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithLatestTestSuite) Test_NewServiceWithLatest_WithEmptyCacheKey() {
	actual, err := NewServiceWithLatest(suite.rawService, suite.cacheStore, "")
	suite.Assert().NoError(err)
	suite.Assert().NotNil(actual)
	suite.Assert().Implements((*Service)(nil), actual, "ServiceWithLatest should implement Service interface")

	expectedBase, _ := newServiceWithCacheBase(suite.rawService, suite.cacheStore, "")
	suite.Assert().Equal(expectedBase, actual.serviceWithCacheBase)
}

func (suite *ServiceWithLatestTestSuite) Test_GetUserID_WhenNoCache() {
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

func (suite *ServiceWithLatestTestSuite) Test_GetUserID_WhenCache() {
	expected := "test user id"
	suite.rawService.On("GetUserID").Return(expected, nil).Once()

	key := fmt.Sprintf("%s/GetUserID", suite.service.cacheKey)
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte("should not be read"),
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetUserID()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual, "cache data should not be read")

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, string(item.Value), "obtained user id should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_GetUserID_WhenRawServiceError() {
	suite.rawService.On("GetUserID").Return("", fmt.Errorf("test: unexpected error")).Once()

	actual, err := suite.service.GetUserID()
	suite.Assert().Error(err)
	suite.Assert().Equal("", actual)

	key := fmt.Sprintf("%s/GetUserID", suite.service.cacheKey)
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryBySlug_WhenNoCache() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WhenCache() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryBySlug", &expected.Slug).Return(expected, nil).Once()

	cached := suite.newTestRepository("docstand", "201", true)
	cachedInBytes, _ := json.Marshal(cached)
	key := fmt.Sprintf("%s/GetRepositoryBySlug/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: cachedInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryBySlug(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual, "cache data should not be read")

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryBySlug_WhenRawServiceError() {
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

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryBySlug_WithNilSlug() {
	suite.rawService.On("GetRepositoryBySlug", (*Slug)(nil)).Return((*Repository)(nil), fmt.Errorf("no slug specified")).Once()

	actual, err := suite.service.GetRepositoryBySlug(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryByID_WhenNoCache() {
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

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryByID_WhenCache() {
	expected := suite.newTestRepository("common-go", "200", true)
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetRepositoryByID", &expected.ID).Return(expected, nil).Once()

	cached := suite.newTestRepository("docstand", "201", true)
	cachedInBytes, _ := json.Marshal(cached)
	key := fmt.Sprintf("%s/GetRepositoryByID/%d", suite.service.cacheKey, expected.ID.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: cachedInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.GetRepositoryByID(&expected.ID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual, "cache data should not be read")

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryByID_WhenRawServiceError() {
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

func (suite *ServiceWithLatestTestSuite) Test_GetRepositoryByID_WithNilSlug() {
	suite.rawService.On("GetRepositoryByID", (*ID)(nil)).Return((*Repository)(nil), fmt.Errorf("no id specified")).Once()

	actual, err := suite.service.GetRepositoryByID(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *ServiceWithLatestTestSuite) Test_ListRepositories_WhenNoCache() {
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

func (suite *ServiceWithLatestTestSuite) Test_ListRepositories_WhenCache() {
	expected := []*Repository{
		suite.newTestRepository("common-go", "200", true),
		suite.newTestRepository("docstand", "201", false),
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListRepositories").Return(expected, nil).Once()

	cached := []*Repository{
		suite.newTestRepository("loadroid", "202", false),
	}
	cachedInBytes, _ := json.Marshal(cached)
	key := fmt.Sprintf("%s/ListRepositories", suite.service.cacheKey)
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: cachedInBytes,
	})
	suite.Assert().NoError(err)

	actual, err := suite.service.ListRepositories()
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual, "cache data should not be read")

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained repository information should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListRepositories_WhenRawServiceError() {
	suite.rawService.On("ListRepositories").Return(([]*Repository)(nil), fmt.Errorf("test: unexpected error")).Once()

	actual, err := suite.service.ListRepositories()
	suite.Assert().Error(err)
	suite.Assert().Len(actual, 0)

	key := fmt.Sprintf("%s/ListRepositories", suite.service.cacheKey)
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithLatestTestSuite) Test_ListReferences_WhenNoCache() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}

	expectedTags := &references{
		Slug: expectedBranches.Slug,
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}

	expectedRefs := append(expectedBranches.Refs, expectedTags.Refs...)
	suite.rawService.On("ListReferences", &expectedBranches.Slug).Return(expectedRefs, nil).Once()

	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListReferences_WhenCache() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.service.storeBranchesCache(&expectedBranches.Slug, expectedBranches.Refs)

	expectedTags := &references{
		Slug: expectedBranches.Slug,
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.service.storeTagsCache(&expectedTags.Slug, expectedTags.Refs)

	expectedRefs := append(expectedBranches.Refs, expectedTags.Refs...)
	suite.rawService.On("ListReferences", &expectedBranches.Slug).Return(expectedRefs, nil).Once()

	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs, "cache data should not be read")

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListReferences_WhenRawServiceNotFoundError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListReferences", slug).Return(([]*Reference)(nil), ErrNotFound).Once()

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().Equal(ErrNotFound, err)
	suite.Assert().Len(actualRefs, 0)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualBranchesCache, 0, "empty list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualTagsCache, 0, "empty list of tags should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListReferences_WhenRawServiceError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListReferences", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(actualBranchesCache)

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(actualTagsCache)
}

func (suite *ServiceWithLatestTestSuite) Test_ListReferences_WithNilSlug() {
	suite.rawService.On("ListReferences", (*Slug)(nil)).Return(([]*Reference)(nil), fmt.Errorf("test: no slug specified")).Once()

	actualRefs, err := suite.service.ListReferences(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRefs)

	suite.rawService.AssertCalled(suite.T(), "ListReferences", (*Slug)(nil))
}

func (suite *ServiceWithLatestTestSuite) Test_ListBranches_WhenNoCache() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.rawService.On("ListBranches", &expected.Slug).Return(expected.Refs, nil).Once()

	actualRefs, err := suite.service.ListBranches(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualBranchesCache, "obtained list of branches should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListBranches_WhenCache() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.service.storeBranchesCache(&expected.Slug, expected.Refs)
	suite.rawService.On("ListBranches", &expected.Slug).Return(expected.Refs, nil).Once()

	actualRefs, err := suite.service.ListBranches(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs, "cache data should not be read")

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualBranchesCache, "obtained list of branches should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListBranches_WhenRawServiceNotFoundError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListBranches", slug).Return(([]*Reference)(nil), ErrNotFound).Once()

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().Equal(ErrNotFound, err)
	suite.Assert().Len(actualRefs, 0)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualBranchesCache, 0, "empty list of branches should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListBranches_WhenRawServiceError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListBranches", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(actualBranchesCache)
}

func (suite *ServiceWithLatestTestSuite) Test_ListBranches_WithNilSlug() {
	suite.rawService.On("ListBranches", (*Slug)(nil)).Return(([]*Reference)(nil), fmt.Errorf("test: no slug specified")).Once()

	actualRefs, err := suite.service.ListBranches(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRefs)

	suite.rawService.AssertCalled(suite.T(), "ListBranches", (*Slug)(nil))
}

func (suite *ServiceWithLatestTestSuite) Test_ListTags_WhenNoCache() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.rawService.On("ListTags", &expected.Slug).Return(expected.Refs, nil).Once()

	actualRefs, err := suite.service.ListTags(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	actualTagsCache, err := suite.service.retrieveTagsCache(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListTags_WhenCache() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.service.storeTagsCache(&expected.Slug, expected.Refs)
	suite.rawService.On("ListTags", &expected.Slug).Return(expected.Refs, nil).Once()

	actualRefs, err := suite.service.ListTags(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs, "cache data should not be read")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListTags_WhenRawServiceNotFoundError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListTags", slug).Return(([]*Reference)(nil), ErrNotFound).Once()

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().Equal(ErrNotFound, err)
	suite.Assert().Len(actualRefs, 0)

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualTagsCache, 0, "empty list of tags should be cached")
}

func (suite *ServiceWithLatestTestSuite) Test_ListTags_WhenRawServiceError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListTags", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(actualTagsCache)
}

func (suite *ServiceWithLatestTestSuite) Test_ListTags_WithNilSlug() {
	suite.rawService.On("ListTags", (*Slug)(nil)).Return(([]*Reference)(nil), fmt.Errorf("test: no slug specified")).Once()

	actualRefs, err := suite.service.ListTags(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRefs)

	suite.rawService.AssertCalled(suite.T(), "ListTags", (*Slug)(nil))
}

func (suite *ServiceWithLatestTestSuite) Test_GetCommitID_WhenNoCache() {
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

func (suite *ServiceWithLatestTestSuite) Test_GetCommitID_WhenCache() {
	expected := suite.newTestCommit("common-go", "master", "0000000000000000000000000000000000000000")
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("GetCommitID", &expected.Slug, &expected.Ref).Return(expected.ID, nil).Once()

	cached := suite.newTestCommit("common-go", "master", "1234567890abcdefghijklmnopqrstuvwxyz")
	cachedInBytes, _ := json.Marshal(cached)
	key := fmt.Sprintf("%s/GetCommitID/%d/%d", suite.service.cacheKey, expected.Slug.Hash(), expected.Ref.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: cachedInBytes,
	})
	suite.Assert().NoError(err)

	actualID, err := suite.service.GetCommitID(&expected.Slug, &expected.Ref)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.ID, actualID, "cache data should not be read")

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained commit id should be cached with slug and reference")
}

func (suite *ServiceWithLatestTestSuite) Test_GetCommitID_WhenRawServiceError() {
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

func (suite *ServiceWithLatestTestSuite) Test_GetCommitID_WithNilSlug() {
	ref := new(Reference)
	suite.rawService.On("GetCommitID", (*Slug)(nil), ref).Return("", fmt.Errorf("no slug specified")).Once()

	actualID, err := suite.service.GetCommitID(nil, ref)
	suite.Assert().Error(err)
	suite.Assert().Equal("", actualID)
}

func (suite *ServiceWithLatestTestSuite) Test_GetCommitID_WithNilReference() {
	slug := new(Slug)
	suite.rawService.On("GetCommitID", slug, (*Reference)(nil)).Return("", fmt.Errorf("no reference specified")).Once()

	actualID, err := suite.service.GetCommitID(slug, nil)
	suite.Assert().Error(err)
	suite.Assert().Equal("", actualID)
}

type ServiceWithCacheTestSuite struct {
	ServiceWithCacheTestSuiteBase

	service *ServiceWithCache
}

func Test_ServiceWithCacheTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceWithCacheTestSuite))
}

func (suite *ServiceWithCacheTestSuite) SetupTest() {
	suite.ServiceWithCacheTestSuiteBase.SetupTest()

	suite.service, _ = NewServiceWithCache(suite.rawService, suite.cacheStore, suite.cacheKey)
}

func (suite *ServiceWithCacheTestSuite) TearDownTest() {
	suite.service = nil

	suite.ServiceWithCacheTestSuiteBase.TearDownTest()
}

func (suite *ServiceWithCacheTestSuite) Test_NewServiceWithCache_WithValidParameters() {
	actual, err := NewServiceWithCache(suite.rawService, suite.cacheStore, suite.cacheKey)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(actual)
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
	suite.Assert().NotNil(actual)
	suite.Assert().Implements((*Service)(nil), actual, "ServiceWithCache should implement Service interface")
}

func (suite *ServiceWithCacheTestSuite) Test_Latest() {
	actual := suite.service.Latest()
	suite.Assert().NotNil(actual)
	suite.Assert().True(suite.service.latest == actual)
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WhenCacheMiss() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WhenValidCacheHit() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WhenCacheConflict() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryBySlug_WhenInvalidCacheHit() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WhenCacheMiss() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WhenValidCacheHit() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WhenCacheConflict() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetRepositoryByID_WhenInvalidCacheHit() {
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

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenBothCacheMiss() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}

	expectedTags := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}

	expectedRefs := append(expectedBranches.Refs, expectedTags.Refs...)
	suite.rawService.On("ListReferences", &expectedTags.Slug).Return(expectedRefs, nil).Once()

	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenBranchesCacheMiss() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.rawService.On("ListBranches", &expectedBranches.Slug).Return(expectedBranches.Refs, nil).Once()

	expectedTags := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.service.storeTagsCache(&expectedTags.Slug, expectedTags.Refs)

	expectedRefs := append(expectedBranches.Refs, expectedTags.Refs...)
	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenTagsCacheMiss() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.service.storeBranchesCache(&expectedBranches.Slug, expectedBranches.Refs)

	expectedTags := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.rawService.On("ListTags", &expectedTags.Slug).Return(expectedTags.Refs, nil).Once()

	expectedRefs := append(expectedBranches.Refs, expectedTags.Refs...)
	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenBothCacheHit() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.service.storeBranchesCache(&expectedBranches.Slug, expectedBranches.Refs)

	expectedTags := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.service.storeTagsCache(&expectedTags.Slug, expectedTags.Refs)

	expectedRefs := append(expectedBranches.Refs, expectedTags.Refs...)
	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedRefs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenBothEmptyCacheHit() {
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
	}
	suite.service.storeBranchesCache(&expectedBranches.Slug, expectedBranches.Refs)

	expectedTags := &references{
		Slug: expectedBranches.Slug,
	}
	suite.service.storeTagsCache(&expectedTags.Slug, expectedTags.Refs)

	actualRefs, err := suite.service.ListReferences(&expectedBranches.Slug)
	suite.Assert().Equal(ErrNotFound, err)
	suite.Assert().Len(actualRefs, 0)
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenListBranchesNotFoundError() {
	slug := suite.newTestSlug("common-go")
	expectedTags := &references{
		Slug: *slug,
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	suite.service.storeTagsCache(&expectedTags.Slug, expectedTags.Refs)
	suite.rawService.On("ListBranches", slug).Return(([]*Reference)(nil), ErrNotFound).Once()

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualBranchesCache, 0, "empty list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(&expectedTags.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedTags.Refs, actualTagsCache, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenListBranchesError() {
	slug := suite.newTestSlug("common-go")
	suite.service.storeTagsCache(slug, nil)
	suite.rawService.On("ListBranches", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(actualBranchesCache)

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualTagsCache, 0)
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenListTagsNotFoundError() {
	slug := suite.newTestSlug("common-go")
	expectedBranches := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	suite.service.storeBranchesCache(&expectedBranches.Slug, expectedBranches.Refs)
	suite.rawService.On("ListTags", slug).Return(([]*Reference)(nil), ErrNotFound).Once()

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualRefs)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(&expectedBranches.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedBranches.Refs, actualBranchesCache, "obtained list of branches should be cached")

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualTagsCache, 0, "empty list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WhenListTagsError() {
	slug := suite.newTestSlug("common-go")
	suite.service.storeBranchesCache(slug, nil)
	suite.rawService.On("ListTags", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListReferences(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	actualBranchesCache, err := suite.service.retrieveBranchesCache(slug)
	suite.Assert().NoError(err)
	suite.Assert().Len(actualBranchesCache, 0)

	actualTagsCache, err := suite.service.retrieveTagsCache(slug)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(actualTagsCache)
}

func (suite *ServiceWithCacheTestSuite) Test_ListReferences_WithNilSlug() {
	suite.rawService.On("ListReferences", (*Slug)(nil)).Return(([]*Reference)(nil), fmt.Errorf("test: no slug specified")).Once()

	actualRefs, err := suite.service.ListReferences(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRefs)

	suite.rawService.AssertCalled(suite.T(), "ListReferences", (*Slug)(nil))
}

func (suite *ServiceWithCacheTestSuite) Test_ListBranches_WhenCacheMiss() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListBranches", &expected.Slug).Return(expected.Refs, nil).Once()

	actualRefs, err := suite.service.ListBranches(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	key := fmt.Sprintf("%s/ListBranches/%d", suite.service.cacheKey, expected.Slug.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained list of branches should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListBranches_WhenValidCacheHit() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	key := fmt.Sprintf("%s/ListBranches/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: expectedInBytes,
	})
	suite.Assert().NoError(err)

	actualRefs, err := suite.service.ListBranches(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	suite.rawService.AssertNotCalled(suite.T(), "ListBranches")
}

func (suite *ServiceWithCacheTestSuite) Test_ListBranches_WhenCacheConflict() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListBranches", &expected.Slug).Return(expected.Refs, nil).Once()

	conflict := &references{
		Slug: *suite.newTestSlug("conflict-repo"),
		Refs: []*Reference{suite.newTestBranch("conflict")},
	}
	conflictInBytes, _ := json.Marshal(conflict)
	key := fmt.Sprintf("%s/ListBranches/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: conflictInBytes,
	})
	suite.Assert().NoError(err)

	actualRefs, err := suite.service.ListBranches(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained list of branches should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListBranches_WhenInvalidCacheHit() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestBranch("master"),
			suite.newTestBranch("list-repository-branches"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListBranches", &expected.Slug).Return(expected.Refs, nil).Once()

	key := fmt.Sprintf("%s/ListBranches/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte("non json value"),
	})
	suite.Assert().NoError(err)

	actualRefs, err := suite.service.ListBranches(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained list of branches should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListBranches_WhenRawServiceError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListBranches", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListBranches(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	key := fmt.Sprintf("%s/ListBranches/%d", suite.service.cacheKey, slug.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_ListBranches_WithNilSlug() {
	suite.rawService.On("ListBranches", (*Slug)(nil)).Return(([]*Reference)(nil), fmt.Errorf("test: no slug specified")).Once()

	actualRefs, err := suite.service.ListBranches(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRefs)
}

func (suite *ServiceWithCacheTestSuite) Test_ListTags_WhenCacheMiss() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListTags", &expected.Slug).Return(expected.Refs, nil).Once()

	actualRefs, err := suite.service.ListTags(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	key := fmt.Sprintf("%s/ListTags/%d", suite.service.cacheKey, expected.Slug.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListTags_WhenValidCacheHit() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	key := fmt.Sprintf("%s/ListTags/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: expectedInBytes,
	})
	suite.Assert().NoError(err)

	actualRefs, err := suite.service.ListTags(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	suite.rawService.AssertNotCalled(suite.T(), "ListTags")
}

func (suite *ServiceWithCacheTestSuite) Test_ListTags_WhenCacheConflict() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListTags", &expected.Slug).Return(expected.Refs, nil).Once()

	conflict := &references{
		Slug: *suite.newTestSlug("conflict-repo"),
		Refs: []*Reference{suite.newTestTag("conflict")},
	}
	conflictInBytes, _ := json.Marshal(conflict)
	key := fmt.Sprintf("%s/ListTags/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: conflictInBytes,
	})
	suite.Assert().NoError(err)

	actualRefs, err := suite.service.ListTags(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListTags_WhenInvalidCacheHit() {
	expected := &references{
		Slug: *suite.newTestSlug("common-go"),
		Refs: []*Reference{
			suite.newTestTag("v1.0"),
			suite.newTestTag("list-repository-tags"),
		},
	}
	expectedInBytes, _ := json.Marshal(expected)
	suite.rawService.On("ListTags", &expected.Slug).Return(expected.Refs, nil).Once()

	key := fmt.Sprintf("%s/ListTags/%d", suite.service.cacheKey, expected.Slug.Hash())
	err := suite.cacheStore.Set(&cache.Item{
		Key:   key,
		Value: []byte("non json value"),
	})
	suite.Assert().NoError(err)

	actualRefs, err := suite.service.ListTags(&expected.Slug)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected.Refs, actualRefs)

	item, err := suite.cacheStore.Get(key)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedInBytes, item.Value, "obtained list of tags should be cached")
}

func (suite *ServiceWithCacheTestSuite) Test_ListTags_WhenRawServiceError() {
	slug := suite.newTestSlug("common-go")
	suite.rawService.On("ListTags", slug).Return(([]*Reference)(nil), fmt.Errorf("test: unexpected error")).Once()

	actualRefs, err := suite.service.ListTags(slug)
	suite.Assert().Error(err)
	suite.Assert().Len(actualRefs, 0)

	key := fmt.Sprintf("%s/ListTags/%d", suite.service.cacheKey, slug.Hash())
	item, err := suite.cacheStore.Get(key)
	suite.Assert().Equal(cache.ErrCacheMiss, err)
	suite.Assert().Nil(item)
}

func (suite *ServiceWithCacheTestSuite) Test_ListTags_WithNilSlug() {
	suite.rawService.On("ListTags", (*Slug)(nil)).Return(([]*Reference)(nil), fmt.Errorf("test: no slug specified")).Once()

	actualRefs, err := suite.service.ListTags(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actualRefs)
}

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WhenCacheMiss() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WhenValidCacheHit() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WhenCacheConflictOnSlug() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WhenCacheConflictOnReference() {
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

func (suite *ServiceWithCacheTestSuite) Test_GetCommitID_WhenInvalidCacheHit() {
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
