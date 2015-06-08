package gcp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"google.golang.org/appengine/datastore"

	"github.com/tractrix/common-go/gcp/gcptest"
	"github.com/tractrix/common-go/repository"
	"github.com/tractrix/docstand/server/data"
)

type DocumentationResultEntityTestSuite struct {
	suite.Suite
}

func Test_DocumentationResultEntityTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentationResultEntityTestSuite))
}

func (suite *DocumentationResultEntityTestSuite) Test_putDocument_WithExistingDocument() {
	expectedGoDoc := data.Document{
		Type:     "godoc",
		CommitID: "ada984a02cca9f8362d0c3b75fb50c0591b28924",
	}
	expectedJavaDoc := data.Document{
		Type:     "javadoc",
		CommitID: "37dec2e603e1d24deb263b9f4cd915c716dab24a",
	}
	entity := &documentationResultEntity{
		Documents: []data.Document{
			expectedGoDoc,
			{
				Type:     expectedJavaDoc.Type,
				CommitID: "ada984a02cca9f8362d0c3b75fb50c0591b28924",
			},
		},
	}

	err := entity.putDocument(&expectedJavaDoc)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expectedGoDoc, entity.Documents[0])
	suite.Assert().Equal(expectedJavaDoc, entity.Documents[1])
}

// TODO: Add test cases for DocumentationStore.
type DocumentationStoreTestSuite struct {
	gcptest.AETestSuite

	store *DocumentationStore
}

func Test_DocumentationStoreTestSuite(t *testing.T) {
	suite.Run(t, new(DocumentationStoreTestSuite))
}

func (suite *DocumentationStoreTestSuite) SetupTest() {
	suite.store = NewDocumentationStore(suite.Context())
	suite.clearDatastore()
}

func (suite *DocumentationStoreTestSuite) TearDownTest() {
	suite.clearDatastore()
	suite.store = nil
}

func (suite *DocumentationStoreTestSuite) countEntities(kind string) (int, error) {
	rootKey := suite.store.newRootKey()
	return datastore.NewQuery(kind).Ancestor(rootKey).KeysOnly().Count(suite.Context())
}

func (suite *DocumentationStoreTestSuite) clearDatastore() error {
	// Delete all entities from datastore.
	rootKey := suite.store.newRootKey()
	keys, err := datastore.NewQuery("").Ancestor(rootKey).KeysOnly().GetAll(suite.Context(), nil)
	if err != nil {
		return err
	}
	return datastore.DeleteMulti(suite.Context(), keys)
}

func (suite *DocumentationStoreTestSuite) newRepositoryID() int64 {
	return int64(100)
}

func (suite *DocumentationStoreTestSuite) newReference() *repository.Reference {
	return &repository.Reference{
		Type: repository.ReferenceTypeBranch,
		Name: "master",
	}
}

func (suite *DocumentationStoreTestSuite) newDocumentationTarget() *data.DocumentationTarget {
	return &data.DocumentationTarget{
		RepositoryID: suite.newRepositoryID(),
		Reference:    *suite.newReference(),
	}
}

func (suite *DocumentationStoreTestSuite) newDeletingTarget() *data.DeletingTarget {
	return &data.DeletingTarget{
		RepositoryID: suite.newRepositoryID(),
		Reference:    *suite.newReference(),
	}
}

func (suite *DocumentationStoreTestSuite) Test_newRootKey() {
	actual := suite.store.newRootKey()
	suite.assertRootKey(actual)
}

func (suite *DocumentationStoreTestSuite) assertRootKey(actual *datastore.Key) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal("root", actual.StringID())
	suite.Assert().Equal(kindDocumentation, actual.Kind())
	suite.Assert().Nil(actual.Parent())
}

func (suite *DocumentationStoreTestSuite) Test_newRepositoryIDKey_WithValidRepositoryID() {
	repoID := suite.newRepositoryID()
	actual := suite.store.newRepositoryIDKey(repoID)
	suite.assertRepositoryIDKey(actual, repoID)
}

func (suite *DocumentationStoreTestSuite) assertRepositoryIDKey(actual *datastore.Key, repoID int64) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal(repoID, actual.IntID())
	suite.Assert().Equal(kindDocumentation, actual.Kind())
}

func (suite *DocumentationStoreTestSuite) Test_newReferenceTypeKey_WithValidParameters() {
	repoID := suite.newRepositoryID()
	parentKey := suite.store.newRepositoryIDKey(repoID)

	ref := suite.newReference()
	actual, err := suite.store.newReferenceTypeKey(parentKey, ref)
	suite.Assert().NoError(err)
	suite.assertReferenceTypeKey(actual, parentKey, ref)
}

func (suite *DocumentationStoreTestSuite) Test_newReferenceTypeKey_WithNilParentKey() {
	ref := suite.newReference()
	actual, err := suite.store.newReferenceTypeKey(nil, ref)
	suite.Assert().NoError(err)
	suite.assertReferenceTypeKey(actual, nil, ref)
}

func (suite *DocumentationStoreTestSuite) Test_newReferenceTypeKey_WithNilReference() {
	repoID := suite.newRepositoryID()
	parentKey := suite.store.newRepositoryIDKey(repoID)

	actual, err := suite.store.newReferenceTypeKey(parentKey, nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) assertReferenceTypeKey(actual *datastore.Key, parentKey *datastore.Key, ref *repository.Reference) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal(ref.Type, actual.StringID())
	suite.Assert().Equal(kindDocumentation, actual.Kind())
	suite.Assert().Equal(parentKey, actual.Parent())
}

func (suite *DocumentationStoreTestSuite) Test_newDocumentationTargetKey_WithValidParameters() {
	target := suite.newDocumentationTarget()
	kind := kindDocumentationResult
	actual, err := suite.store.newDocumentationTargetKey(target, kind)
	suite.Assert().NoError(err)
	suite.assertDocumentationTargetKey(actual, target, kind)
}

func (suite *DocumentationStoreTestSuite) Test_newDocumentationTargetKey_WithNilTarget() {
	actual, err := suite.store.newDocumentationTargetKey(nil, kindDocumentationResult)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) assertDocumentationTargetKey(actual *datastore.Key, target *data.DocumentationTarget, kind string) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal(target.Reference.Name, actual.StringID())
	suite.Assert().Equal(kind, actual.Kind())

	actualRefTypeKey := actual.Parent()
	expectedParentKey := suite.store.newRepositoryIDKey(target.RepositoryID)
	suite.assertReferenceTypeKey(actualRefTypeKey, expectedParentKey, &target.Reference)
}

func (suite *DocumentationStoreTestSuite) Test_newDocumentationStatusKey_WithValidParameters() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{SequenceID: 100}
	actual, err := suite.store.newDocumentationStatusKey(target, status)
	suite.Assert().NoError(err)
	suite.assertDocumentationStatusKey(actual, target, status)
}

func (suite *DocumentationStoreTestSuite) Test_newDocumentationStatustKey_WithNilTarget() {
	status := &data.DocumentationStatus{SequenceID: 100}
	actual, err := suite.store.newDocumentationStatusKey(nil, status)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) Test_newDocumentationStatustKey_WithNilStatus() {
	target := suite.newDocumentationTarget()
	actual, err := suite.store.newDocumentationStatusKey(target, nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) Test_newDocumentationStatustKey_WithEmptySequenceID() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{SequenceID: 0}
	actual, err := suite.store.newDocumentationStatusKey(target, status)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) assertDocumentationStatusKey(actual *datastore.Key, target *data.DocumentationTarget, status *data.DocumentationStatus) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal(status.SequenceID, actual.IntID())
	suite.Assert().Equal(kindDocumentationStatus, actual.Kind())

	actualTargetKey := actual.Parent()
	suite.assertDocumentationTargetKey(actualTargetKey, target, kindDocumentationStatus)
}

func (suite *DocumentationStoreTestSuite) Test_newCommitIDKey_WithValidParameters() {
	target := suite.newDocumentationTarget()
	kind := kindDocumentationResult
	commitID := "ada984a02cca9f8362d0c3b75fb50c0591b28924"
	actual, err := suite.store.newCommitIDKey(target, kind, commitID)
	suite.Assert().NoError(err)
	suite.assertcommitIDKey(actual, target, kind, commitID)
}

func (suite *DocumentationStoreTestSuite) Test_newCommitIDKey_WithNilTarget() {
	kind := kindDocumentationResult
	commitID := "ada984a02cca9f8362d0c3b75fb50c0591b28924"
	actual, err := suite.store.newCommitIDKey(nil, kind, commitID)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) Test_newCommitIDKey_WithEmptycommitID() {
	target := suite.newDocumentationTarget()
	kind := kindDocumentationResult
	actual, err := suite.store.newCommitIDKey(target, kind, "")
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) assertcommitIDKey(actual *datastore.Key, target *data.DocumentationTarget, kind, commitID string) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal(commitID, actual.StringID())
	suite.Assert().Equal(kind, actual.Kind())

	actualTargetKey := actual.Parent()
	suite.assertDocumentationTargetKey(actualTargetKey, target, kind)
}

func (suite *DocumentationStoreTestSuite) Test_newDeletingTargetKey_WithValidTarget() {
	target := suite.newDeletingTarget()
	actual, err := suite.store.newDeletingTargetKey(target)
	suite.Assert().NoError(err)
	suite.assertDeletingTargetKey(actual, target)
}

func (suite *DocumentationStoreTestSuite) Test_newDeletingTargetKey_WithNilTarget() {
	actual, err := suite.store.newDeletingTargetKey(nil)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) assertDeletingTargetKey(actual *datastore.Key, target *data.DeletingTarget) {
	suite.Assert().NotNil(actual)
	suite.Assert().Equal(target.Reference.Name, actual.StringID())
	suite.Assert().Equal(kindDeletingTarget, actual.Kind())

	actualRefTypeKey := actual.Parent()
	expectedParentKey := suite.store.newRepositoryIDKey(target.RepositoryID)
	suite.assertReferenceTypeKey(actualRefTypeKey, expectedParentKey, &target.Reference)
}

func (suite *DocumentationStoreTestSuite) Test_ExistsTarget_WithUnexistingTarget() {
	target := suite.newDocumentationTarget()
	actual, err := suite.store.ExistsTarget(target)
	suite.Assert().NoError(err)
	suite.Assert().False(actual)
}

func (suite *DocumentationStoreTestSuite) Test_ExistsTarget_WithExistingTarget() {
	target := suite.newDocumentationTarget()
	err := suite.store.CreateTarget(target)
	suite.Assert().NoError(err)

	actual, err := suite.store.ExistsTarget(target)
	suite.Assert().NoError(err)
	suite.Assert().True(actual)
}

func (suite *DocumentationStoreTestSuite) Test_ExistsTarget_WithNil() {
	actual, err := suite.store.ExistsTarget(nil)
	suite.Assert().Error(err)
	suite.Assert().False(actual)
}

func (suite *DocumentationStoreTestSuite) Test_MarkDeletingTarget_WithUnexistingTarget() {
	target := suite.newDeletingTarget()
	startTime := time.Now().UTC()
	err := suite.store.MarkDeletingTarget(target)
	endTime := time.Now().UTC()
	suite.Assert().NoError(err)

	actualNumOfEntities, err := suite.countEntities("")
	suite.Assert().NoError(err)
	suite.Assert().Equal(1, actualNumOfEntities)

	key, _ := suite.store.newDeletingTargetKey(target)
	entity := new(deletingTargetEntity)
	err = datastore.Get(suite.Context(), key, entity)
	suite.Assert().NoError(err)
	suite.Assert().True(entity.RequestedAt.Equal(startTime) || entity.RequestedAt.After(startTime), "the current time should be set to RequestedAt")
	suite.Assert().True(entity.RequestedAt.Equal(endTime) || entity.RequestedAt.Before(endTime), "the current time should be set to RequestedAt")
}

func (suite *DocumentationStoreTestSuite) Test_MarkDeletingTarget_WithExistingTarget() {
	target := suite.newDeletingTarget()
	err := suite.store.MarkDeletingTarget(target)
	suite.Assert().NoError(err)

	key, _ := suite.store.newDeletingTargetKey(target)
	entity := new(deletingTargetEntity)
	err = datastore.Get(suite.Context(), key, entity)
	suite.Assert().NoError(err)

	err = suite.store.MarkDeletingTarget(target)
	suite.Assert().NoError(err)

	actualNumOfEntities, err := suite.countEntities("")
	suite.Assert().NoError(err)
	suite.Assert().Equal(1, actualNumOfEntities)

	actualEntity := new(deletingTargetEntity)
	err = datastore.Get(suite.Context(), key, actualEntity)
	suite.Assert().NoError(err)
	suite.Assert().Equal(entity.RequestedAt, actualEntity.RequestedAt, "existing entity should not be updated")
}

func (suite *DocumentationStoreTestSuite) Test_MarkDeletingTarget_WithNil() {
	err := suite.store.MarkDeletingTarget(nil)
	suite.Assert().Error(err)

	actualNumOfEntities, err := suite.countEntities("")
	suite.Assert().NoError(err)
	suite.Assert().Equal(0, actualNumOfEntities)
}

func (suite *DocumentationStoreTestSuite) Test_UnmarkDeletingTarget_WithUnexistingTarget() {
	target := suite.newDeletingTarget()
	err := suite.store.UnmarkDeletingTarget(target)
	suite.Assert().NoError(err)

	actualNumOfEntities, err := suite.countEntities("")
	suite.Assert().NoError(err)
	suite.Assert().Equal(0, actualNumOfEntities)

	key, _ := suite.store.newDeletingTargetKey(target)
	err = datastore.Get(suite.Context(), key, new(deletingTargetEntity))
	suite.Assert().Equal(datastore.ErrNoSuchEntity, err)
}

func (suite *DocumentationStoreTestSuite) Test_UnmarkDeletingTarget_WithExistingTarget() {
	target := suite.newDeletingTarget()
	err := suite.store.MarkDeletingTarget(target)
	suite.Assert().NoError(err)

	err = suite.store.UnmarkDeletingTarget(target)
	suite.Assert().NoError(err)

	actualNumOfEntities, err := suite.countEntities("")
	suite.Assert().NoError(err)
	suite.Assert().Equal(0, actualNumOfEntities)

	key, _ := suite.store.newDeletingTargetKey(target)
	err = datastore.Get(suite.Context(), key, new(deletingTargetEntity))
	suite.Assert().Equal(datastore.ErrNoSuchEntity, err)
}

func (suite *DocumentationStoreTestSuite) Test_UnmarkDeletingTarget_WithNil() {
	target := suite.newDeletingTarget()
	err := suite.store.MarkDeletingTarget(target)
	suite.Assert().NoError(err)

	err = suite.store.UnmarkDeletingTarget(nil)
	suite.Assert().Error(err)

	actualNumOfEntities, err := suite.countEntities("")
	suite.Assert().NoError(err)
	suite.Assert().Equal(1, actualNumOfEntities)

	key, _ := suite.store.newDeletingTargetKey(target)
	err = datastore.Get(suite.Context(), key, new(deletingTargetEntity))
	suite.Assert().NoError(err)
}

func (suite *DocumentationStoreTestSuite) Test_IsDeletingTarget_WithUnexistingTarget() {
	target := suite.newDeletingTarget()
	isDeleting, err := suite.store.IsDeletingTarget(target)
	suite.Assert().NoError(err)
	suite.Assert().False(isDeleting)
}

func (suite *DocumentationStoreTestSuite) Test_IsDeletingTarget_WithExistingTarget() {
	target := suite.newDeletingTarget()
	err := suite.store.MarkDeletingTarget(target)
	suite.Assert().NoError(err)

	isDeleting, err := suite.store.IsDeletingTarget(target)
	suite.Assert().NoError(err)
	suite.Assert().True(isDeleting)
}

func (suite *DocumentationStoreTestSuite) Test_IsDeletingTarget_WithNil() {
	isDeleting, err := suite.store.IsDeletingTarget(nil)
	suite.Assert().Error(err)
	suite.Assert().False(isDeleting)
}

func (suite *DocumentationStoreTestSuite) Test_PutDocument_WithExistingTarget() {
	target := suite.newDocumentationTarget()
	err := suite.store.CreateTarget(target)
	suite.Assert().NoError(err)

	document := &data.Document{
		Type:     "godoc",
		CommitID: "ada984a02cca9f8362d0c3b75fb50c0591b28924",
	}
	err = suite.store.PutDocument(target, document)
	suite.Assert().NoError(err)

	actualResult, err := suite.store.GetResult(target)
	suite.Assert().NoError(err)
	suite.Assert().NotNil(actualResult)
	suite.Assert().Len(actualResult.Documents, 1)
	suite.Assert().Equal(document.CommitID, actualResult.Documents[document.Type])
}

func (suite *DocumentationStoreTestSuite) Test_PutDocument_WithUnexistingTarget() {
	target := suite.newDocumentationTarget()
	document := &data.Document{
		Type:     "godoc",
		CommitID: "ada984a02cca9f8362d0c3b75fb50c0591b28924",
	}
	err := suite.store.PutDocument(target, document)
	suite.Assert().Equal(data.ErrNoSuchTarget, err)
}

func (suite *DocumentationStoreTestSuite) Test_PutDocument_WithNilTarget() {
	document := &data.Document{
		Type:     "godoc",
		CommitID: "ada984a02cca9f8362d0c3b75fb50c0591b28924",
	}
	err := suite.store.PutDocument(nil, document)
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_PutDocument_WithNilDocument() {
	target := suite.newDocumentationTarget()
	err := suite.store.PutDocument(target, nil)
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_CreateStatus_WithUnexistingStatus() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	startTime := time.Now().UTC()
	err := suite.store.CreateStatus(target, status)
	endTime := time.Now().UTC()
	suite.Assert().NoError(err)

	key, _ := suite.store.newDocumentationStatusKey(target, status)
	entity := new(documentationStatusEntity)
	err = datastore.Get(suite.Context(), key, entity)
	suite.Assert().NoError(err)
	suite.Assert().Equal(status.CommitID, entity.CommitID)
	suite.Assert().Equal(status.TaskID, entity.TaskID)
	suite.Assert().Equal(status.Status, entity.Status)
	suite.Assert().True(entity.RequestedAt.UTC().Equal(startTime) || entity.RequestedAt.UTC().After(startTime))
	suite.Assert().True(entity.RequestedAt.UTC().Equal(endTime) || entity.RequestedAt.UTC().Before(endTime))
	suite.Assert().True(entity.UpdatedAt.UTC().Equal(startTime) || entity.UpdatedAt.UTC().After(startTime))
	suite.Assert().True(entity.UpdatedAt.UTC().Equal(endTime) || entity.UpdatedAt.UTC().Before(endTime))
}

func (suite *DocumentationStoreTestSuite) Test_CreateStatus_WithExistingStatus() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	err := suite.store.CreateStatus(target, status)
	suite.Assert().NoError(err)

	err = suite.store.CreateStatus(target, status)
	suite.Assert().Equal(data.ErrStatusAlreadyExists, err)
}

func (suite *DocumentationStoreTestSuite) Test_CreateStatus_WithNilTarget() {
	status := &data.DocumentationStatus{SequenceID: 100}
	err := suite.store.CreateStatus(nil, status)
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_CreateStatus_WithNilStatus() {
	target := suite.newDocumentationTarget()
	err := suite.store.CreateStatus(target, nil)
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) TestGetStatus_WithExistingStatus() {
	target := suite.newDocumentationTarget()
	// Note that expected RequestedAt and UpdatedAt should be expressed in local time.
	expected := &data.DocumentationStatus{
		SequenceID:  100,
		CommitID:    "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:      data.DocumentationStatusQueuing,
		RequestedAt: time.Now().Round(time.Microsecond),
		UpdatedAt:   time.Now().Round(time.Microsecond),
	}
	key, err := suite.store.newDocumentationStatusKey(target, expected)
	suite.Assert().NoError(err)
	_, err = datastore.Put(suite.Context(), key, &documentationStatusEntity{
		CommitID:    expected.CommitID,
		Status:      expected.Status,
		RequestedAt: expected.RequestedAt,
		UpdatedAt:   expected.UpdatedAt,
	})
	suite.Assert().NoError(err)

	actual, err := suite.store.GetStatus(target, expected.SequenceID)
	suite.Assert().NoError(err)
	suite.Assert().Equal(expected, actual)
}

func (suite *DocumentationStoreTestSuite) TestGetStatus_WithUnexistingStatus() {
	target := suite.newDocumentationTarget()
	sequenceID := int64(100)
	actual, err := suite.store.GetStatus(target, sequenceID)
	suite.Assert().Equal(data.ErrNoSuchStatus, err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) TestGetStatus_WithNilTarget() {
	sequenceID := int64(100)
	actual, err := suite.store.GetStatus(nil, sequenceID)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) TestGetStatus_WhenDatastoreError() {
	// Specifying an empty target can cause a datastore error because
	// an invalid datastore key including incomplete keys in ancestors
	// will be created.
	target := new(data.DocumentationTarget)
	sequenceID := int64(100)
	actual, err := suite.store.GetStatus(target, sequenceID)
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatus_WithExistingStatus() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	key, err := suite.store.newDocumentationStatusKey(target, status)
	suite.Assert().NoError(err)
	oldEntity := &documentationStatusEntity{
		CommitID:    status.CommitID,
		Status:      status.Status,
		RequestedAt: time.Now().UTC().Round(time.Microsecond),
		UpdatedAt:   time.Now().UTC().Round(time.Microsecond),
	}
	_, err = datastore.Put(suite.Context(), key, oldEntity)
	suite.Assert().NoError(err)

	status.TaskID = "task1"
	status.Status = data.DocumentationStatusQueued
	startTime := time.Now().UTC()
	err = suite.store.UpdateStatus(target, status)
	endTime := time.Now().UTC()
	suite.Assert().NoError(err)

	entity := new(documentationStatusEntity)
	err = datastore.Get(suite.Context(), key, entity)
	suite.Assert().NoError(err)
	suite.Assert().Equal(status.CommitID, entity.CommitID)
	suite.Assert().Equal(status.TaskID, entity.TaskID)
	suite.Assert().Equal(status.Status, entity.Status)
	suite.Assert().Equal(oldEntity.RequestedAt, entity.RequestedAt.UTC())
	suite.Assert().True(entity.UpdatedAt.UTC().Equal(startTime) || entity.UpdatedAt.UTC().After(startTime))
	suite.Assert().True(entity.UpdatedAt.UTC().Equal(endTime) || entity.UpdatedAt.UTC().Before(endTime))
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatus_WithUnexistingStatus() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	err := suite.store.UpdateStatus(target, status)
	suite.Assert().Equal(data.ErrNoSuchStatus, err)
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatus_WithNilTarget() {
	status := &data.DocumentationStatus{SequenceID: 100}
	err := suite.store.UpdateStatus(nil, status)
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatus_WithNilStatus() {
	target := suite.newDocumentationTarget()
	err := suite.store.UpdateStatus(target, nil)
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatusWithCondition_WithExistingStatus() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	key, err := suite.store.newDocumentationStatusKey(target, status)
	suite.Assert().NoError(err)
	oldEntity := &documentationStatusEntity{
		CommitID:    status.CommitID,
		Status:      status.Status,
		RequestedAt: time.Now().UTC().Round(time.Microsecond),
		UpdatedAt:   time.Now().UTC().Round(time.Microsecond),
	}
	_, err = datastore.Put(suite.Context(), key, oldEntity)
	suite.Assert().NoError(err)

	status.TaskID = "task1"
	status.Status = data.DocumentationStatusQueued
	startTime := time.Now().UTC()
	err = suite.store.UpdateStatusWithCondition(target, status, oldEntity.Status)
	endTime := time.Now().UTC()
	suite.Assert().NoError(err)

	entity := new(documentationStatusEntity)
	err = datastore.Get(suite.Context(), key, entity)
	suite.Assert().NoError(err)
	suite.Assert().Equal(status.CommitID, entity.CommitID)
	suite.Assert().Equal(status.TaskID, entity.TaskID)
	suite.Assert().Equal(status.Status, entity.Status)
	suite.Assert().Equal(oldEntity.RequestedAt, entity.RequestedAt.UTC())
	suite.Assert().True(entity.UpdatedAt.UTC().Equal(startTime) || entity.UpdatedAt.UTC().After(startTime))
	suite.Assert().True(entity.UpdatedAt.UTC().Equal(endTime) || entity.UpdatedAt.UTC().Before(endTime))
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatusWithCondition_WithUnsatisfiedCondition() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	key, err := suite.store.newDocumentationStatusKey(target, status)
	suite.Assert().NoError(err)
	oldEntity := &documentationStatusEntity{
		CommitID:    status.CommitID,
		Status:      status.Status,
		RequestedAt: time.Now().UTC().Round(time.Microsecond),
		UpdatedAt:   time.Now().UTC().Round(time.Microsecond),
	}
	_, err = datastore.Put(suite.Context(), key, oldEntity)
	suite.Assert().NoError(err)

	status.TaskID = "task1"
	status.Status = data.DocumentationStatusQueued
	err = suite.store.UpdateStatusWithCondition(target, status, "unsatisfied-status")
	suite.Assert().Equal(data.ErrUnsatisfiedStatusCondition, err)

	entity := new(documentationStatusEntity)
	err = datastore.Get(suite.Context(), key, entity)
	suite.Assert().NoError(err)
	suite.Assert().Equal(oldEntity.CommitID, entity.CommitID)
	suite.Assert().Equal(oldEntity.TaskID, entity.TaskID)
	suite.Assert().Equal(oldEntity.Status, entity.Status)
	suite.Assert().Equal(oldEntity.RequestedAt, entity.RequestedAt.UTC())
	suite.Assert().Equal(oldEntity.UpdatedAt, entity.UpdatedAt.UTC())
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatusWithCondition_WithUnexistingStatus() {
	target := suite.newDocumentationTarget()
	status := &data.DocumentationStatus{
		SequenceID: 100,
		CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
		Status:     data.DocumentationStatusQueuing,
	}
	err := suite.store.UpdateStatusWithCondition(target, status, "")
	suite.Assert().Equal(data.ErrNoSuchStatus, err)
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatusWithCondition_WithNilTarget() {
	status := &data.DocumentationStatus{SequenceID: 100}
	err := suite.store.UpdateStatusWithCondition(nil, status, "")
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_UpdateStatusWithCondition_WithNilStatus() {
	target := suite.newDocumentationTarget()
	err := suite.store.UpdateStatusWithCondition(target, nil, "")
	suite.Assert().Error(err)
}

func (suite *DocumentationStoreTestSuite) Test_QueryStatus_WithExistingStatus() {
	target := suite.newDocumentationTarget()
	expectedStatus := data.DocumentationStatusQueued
	statusMap := map[int64]*data.DocumentationStatus{
		100: {
			SequenceID: 100,
			CommitID:   "ada984a02cca9f8362d0c3b75fb50c0591b28924",
			TaskID:     "task1",
			Status:     expectedStatus,
		},
		101: {
			SequenceID: 101,
			CommitID:   "0305c4bbcde73657136fc40f1bba0ecda936c1dc",
			Status:     data.DocumentationStatusStarted,
		},
		102: {
			SequenceID: 102,
			CommitID:   "51e601e3e8d316bc163aefb755db086323beeb31",
			TaskID:     "task2",
			Status:     expectedStatus,
		},
	}
	for _, status := range statusMap {
		err := suite.store.CreateStatus(target, status)
		suite.Assert().NoError(err)
	}

	actuals, err := suite.store.QueryStatus(target, expectedStatus)
	suite.Assert().NoError(err)
	suite.Assert().Len(actuals, 2)

	for _, actual := range actuals {
		expected, ok := statusMap[actual.SequenceID]
		suite.Assert().True(ok)
		suite.Assert().NotNil(expected)
		suite.Assert().Equal(expected.CommitID, actual.CommitID)
		suite.Assert().Equal(expected.TaskID, actual.TaskID)
		suite.Assert().Equal(expected.Status, actual.Status)
		suite.Assert().Equal(expectedStatus, actual.Status)
	}
}

func (suite *DocumentationStoreTestSuite) Test_QueryStatus_WithUnexistingStatus() {
	target := suite.newDocumentationTarget()
	actuals, err := suite.store.QueryStatus(target, "unexisting-status")
	suite.Assert().NoError(err)
	suite.Assert().Len(actuals, 0)
}

func (suite *DocumentationStoreTestSuite) Test_QueryStatus_WithNilTarget() {
	actual, err := suite.store.QueryStatus(nil, "")
	suite.Assert().Error(err)
	suite.Assert().Nil(actual)
}
