package gcp

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

	"github.com/tractrix/common-go/gcp"
	"github.com/tractrix/common-go/repository"
	"github.com/tractrix/docstand/server/data"
)

const (
	kindDocumentation       = "Documentation"
	kindDocumentationResult = "DocumentationResult"
	kindDocumentationStatus = "DocumentationStatus"
	kindDeletingTarget      = "DeletingTarget"
)

type documentationResultEntity struct {
	Documents []data.Document
}

func newDocumentationResultEntity(result *data.DocumentationResult) (*documentationResultEntity, error) {
	if result == nil {
		return nil, fmt.Errorf("no documentation result specified")
	}

	entity := documentationResultEntity{
		Documents: result.Documents.ListDocuments(),
	}

	return &entity, nil
}

func (entity *documentationResultEntity) toDocumentationResult() *data.DocumentationResult {
	return &data.DocumentationResult{
		Documents: data.MakeDocumentMapFromDocuments(entity.Documents),
	}
}

func (entity *documentationResultEntity) putDocument(document *data.Document) error {
	if document == nil {
		return fmt.Errorf("no document specified")
	}

	for i, existingDoc := range entity.Documents {
		if existingDoc.Type == document.Type {
			// entity.Documents should be updated by using an index because
			// existingDoc is not a pointer and updating it does not affect entity.Documents.
			entity.Documents[i] = *document
			return nil
		}
	}

	entity.Documents = append(entity.Documents, *document)

	return nil
}

type documentationStatusEntity struct {
	CommitID    string
	TaskID      string
	Status      string
	RequestedAt time.Time
	UpdatedAt   time.Time
}

func (entity *documentationStatusEntity) toDocumentationStatus() *data.DocumentationStatus {
	return &data.DocumentationStatus{
		CommitID:    entity.CommitID,
		TaskID:      entity.TaskID,
		Status:      entity.Status,
		RequestedAt: entity.RequestedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

type deletingTargetEntity struct {
	RequestedAt time.Time
}

// DocumentationStore loads/saves documentation data from/to Google Cloud Datastore.
// It implements data.DocumentationStore interface.
type DocumentationStore struct {
	context context.Context
}

// NewDocumentationStore returns a new DocumentationStore bound to the context.
func NewDocumentationStore(ctx context.Context) *DocumentationStore {
	return &DocumentationStore{context: ctx}
}

func (store *DocumentationStore) newRootKey() *datastore.Key {
	return datastore.NewKey(store.context, kindDocumentation, "root", 0, nil)
}

func (store *DocumentationStore) newRepositoryIDKey(id int64) *datastore.Key {
	parentKey := store.newRootKey()
	return datastore.NewKey(store.context, kindDocumentation, "", id, parentKey)
}

func (store *DocumentationStore) newReferenceTypeKey(parentKey *datastore.Key, ref *repository.Reference) (*datastore.Key, error) {
	if ref == nil {
		return nil, fmt.Errorf("no repository reference specified")
	}

	return datastore.NewKey(store.context, kindDocumentation, ref.Type, 0, parentKey), nil
}

func (store *DocumentationStore) newDocumentationTargetKey(target *data.DocumentationTarget, kind string) (*datastore.Key, error) {
	if target == nil {
		return nil, fmt.Errorf("no documentation target specified")
	}

	// newReferenceTypeKey never fail
	// because &target.Reference cannot be nil.
	ancestorKey := store.newRepositoryIDKey(target.RepositoryID)
	parentKey, _ := store.newReferenceTypeKey(ancestorKey, &target.Reference)
	return datastore.NewKey(store.context, kind, target.Reference.Name, 0, parentKey), nil
}

func (store *DocumentationStore) newDocumentationStatusKey(target *data.DocumentationTarget, status *data.DocumentationStatus) (*datastore.Key, error) {
	parentKey, err := store.newDocumentationTargetKey(target, kindDocumentationStatus)
	if err != nil {
		return nil, err
	}

	if status == nil {
		return nil, fmt.Errorf("no documentation status specified")
	}
	if status.SequenceID == 0 {
		return nil, fmt.Errorf("no sequence id specified")
	}

	return datastore.NewKey(store.context, kindDocumentationStatus, "", status.SequenceID, parentKey), nil
}

func (store *DocumentationStore) getDocumentationStatusesByTarget(target *data.DocumentationTarget) ([]*data.DocumentationStatus, error) {
	// To Check : When using this as parent key, double no of entities return as output.
	// parentKey := store.newRepositoryIDKey(target.RepositoryID)

	parentKey, err := store.newDocumentationTargetKey(target, kindDocumentationStatus)
	if err != nil {
		return nil, err
	}

	var statusArray []*data.DocumentationStatus

	err = gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {

		statusQuery := datastore.NewQuery(parentKey.Kind()).Ancestor(parentKey)

		var entities []documentationStatusEntity
		keys, err := statusQuery.GetAll(transactionContext, &entities)
		if err != nil {
			return err
		}

		for index, entity := range entities {
			docStatus := entity.toDocumentationStatus()
			docStatus.SequenceID = keys[index].IntID()
			statusArray = append(statusArray, docStatus)
		}

		return nil

	}, nil)

	return statusArray, err

}

func (store *DocumentationStore) newCommitIDKey(target *data.DocumentationTarget, kind, commitID string) (*datastore.Key, error) {
	parentKey, err := store.newDocumentationTargetKey(target, kind)
	if err != nil {
		return nil, err
	}

	if commitID == "" {
		return nil, fmt.Errorf("no commit id specified")
	}

	return datastore.NewKey(store.context, kind, commitID, 0, parentKey), nil
}

func (store *DocumentationStore) newDeletingTargetKey(target *data.DeletingTarget) (*datastore.Key, error) {
	if target == nil {
		return nil, fmt.Errorf("no deleting target specified")
	}

	// newReferenceTypeKey never fail
	// because &target.Reference cannot be nil.
	ancestorKey := store.newRepositoryIDKey(target.RepositoryID)
	parentKey, _ := store.newReferenceTypeKey(ancestorKey, &target.Reference)
	return datastore.NewKey(store.context, kindDeletingTarget, target.Reference.Name, 0, parentKey), nil
}

// CreateTarget saves a new entity for the documentation target.
// If the documentation target already exists, it returns data.ErrTargetAlreadyExists.
func (store *DocumentationStore) CreateTarget(target *data.DocumentationTarget) error {
	key, err := store.newDocumentationTargetKey(target, kindDocumentationResult)
	if err != nil {
		return err
	}

	return gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {
		err := datastore.Get(transactionContext, key, new(documentationResultEntity))
		if err == nil {
			return data.ErrTargetAlreadyExists
		} else if err != datastore.ErrNoSuchEntity {
			return err
		}

		_, err = datastore.Put(transactionContext, key, new(documentationResultEntity))

		return err
	}, nil)
}

// DeleteTarget deletes the documentation target.
func (store *DocumentationStore) DeleteTarget(target *data.DocumentationTarget) error {
	resultKey, err := store.newDocumentationTargetKey(target, kindDocumentationResult)
	if err != nil {
		return err
	}

	// No error must occur.
	statusAncestorKey, _ := store.newDocumentationTargetKey(target, kindDocumentationStatus)

	transactionOptions := &datastore.TransactionOptions{
		XG: true,
	}
	return gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {
		statusQuery := datastore.NewQuery(statusAncestorKey.Kind()).Ancestor(statusAncestorKey).KeysOnly()
		statusKeys, err := statusQuery.GetAll(transactionContext, nil)
		if err != nil {
			return err
		}
		if err := datastore.DeleteMulti(transactionContext, statusKeys); err != nil {
			return err
		}

		return datastore.Delete(transactionContext, resultKey)
	}, transactionOptions)
}

// ExistsTarget returns true if the documentation target already exists.
// It returns false otherwise.
func (store *DocumentationStore) ExistsTarget(target *data.DocumentationTarget) (bool, error) {
	key, err := store.newDocumentationTargetKey(target, kindDocumentationResult)
	if err != nil {
		return false, err
	}

	err = datastore.Get(store.context, key, new(documentationResultEntity))
	switch err {
	case nil:
		return true, nil
	case datastore.ErrNoSuchEntity:
		return false, nil
	}

	return false, err
}

// CountTargetsInRepository returns the number of documentation targets in the repository.
func (store *DocumentationStore) CountTargetsInRepository(id int64) (int, error) {
	key := store.newRepositoryIDKey(id)

	var numOfTargets int
	err := gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {
		count, err := datastore.NewQuery(kindDocumentationResult).Ancestor(key).Count(transactionContext)
		if err != nil {
			return err
		}

		numOfTargets = count
		return nil
	}, nil)

	return numOfTargets, err
}

// MarkDeletingTarget marks the documentation target as deleting.
func (store *DocumentationStore) MarkDeletingTarget(target *data.DeletingTarget) error {
	key, err := store.newDeletingTargetKey(target)
	if err != nil {
		return err
	}

	return gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {
		err := datastore.Get(transactionContext, key, new(deletingTargetEntity))
		if err == nil {
			// If the entity already exists, nothing has to be done.
			return nil
		}
		if err != datastore.ErrNoSuchEntity {
			return nil
		}

		_, err = datastore.Put(transactionContext, key, &deletingTargetEntity{
			RequestedAt: time.Now().UTC(),
		})

		return err
	}, nil)
}

// UnmarkDeletingTarget removes a mark of deleting from the documentation target.
func (store *DocumentationStore) UnmarkDeletingTarget(target *data.DeletingTarget) error {
	key, err := store.newDeletingTargetKey(target)
	if err != nil {
		return err
	}

	return datastore.Delete(store.context, key)
}

// IsDeletingTarget checks if the target is marked as a target being deleted.
func (store *DocumentationStore) IsDeletingTarget(target *data.DeletingTarget) (bool, error) {
	key, err := store.newDeletingTargetKey(target)
	if err != nil {
		return false, err
	}

	err = datastore.Get(store.context, key, new(deletingTargetEntity))
	if err == datastore.ErrNoSuchEntity {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetResult loads the latest documentation information for the documentation target.
func (store *DocumentationStore) GetResult(target *data.DocumentationTarget) (*data.DocumentationResult, error) {
	key, err := store.newDocumentationTargetKey(target, kindDocumentationResult)
	if err != nil {
		return nil, err
	}

	var entity documentationResultEntity
	if err := datastore.Get(store.context, key, &entity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, data.ErrNoSuchTarget
		}

		return nil, err
	}

	return entity.toDocumentationResult(), nil
}

// GetResultMulti is a batch version of GetLatestResult.
func (store *DocumentationStore) GetResultMulti(targets []*data.DocumentationTarget) (map[*data.DocumentationTarget]*data.DocumentationResult, error) {
	var keys []*datastore.Key
	keys2Targets := make(map[*datastore.Key]*data.DocumentationTarget)
	for _, target := range targets {
		key, err := store.newDocumentationTargetKey(target, kindDocumentationResult)
		if err != nil {
			return nil, err
		}

		keys = append(keys, key)
		keys2Targets[key] = target
	}

	// TODO: Reconsider how to handle MultiError.
	entities := make([]*documentationResultEntity, len(keys))
	if err := datastore.GetMulti(store.context, keys, entities); err != nil {
		if _, ok := err.(appengine.MultiError); !ok {
			return nil, err
		}
	}

	results := make(map[*data.DocumentationTarget]*data.DocumentationResult)
	for index, entity := range entities {
		if entity != nil {
			target := keys2Targets[keys[index]]
			results[target] = entity.toDocumentationResult()
		}
	}

	return results, nil
}

// PutDocument saves the document information as the latest one for the documentation target.
func (store *DocumentationStore) PutDocument(target *data.DocumentationTarget, document *data.Document) error {
	key, err := store.newDocumentationTargetKey(target, kindDocumentationResult)
	if err != nil {
		return err
	}

	return gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {
		var entity documentationResultEntity
		if err := datastore.Get(transactionContext, key, &entity); err != nil {
			if err == datastore.ErrNoSuchEntity {
				return data.ErrNoSuchTarget
			}
			return err
		}

		if err := entity.putDocument(document); err != nil {
			return err
		}

		_, err := datastore.Put(transactionContext, key, &entity)

		return err
	}, nil)
}

// CreateStatus saves a new entity for the documentation status.
func (store *DocumentationStore) CreateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus) error {
	key, err := store.newDocumentationStatusKey(target, status)
	if err != nil {
		return err
	}

	return gcp.NewDefaultDatastoreTransaction().Run(store.context, func(tc context.Context) error {
		err := datastore.Get(tc, key, new(documentationStatusEntity))
		if err == nil {
			return data.ErrStatusAlreadyExists
		}
		if err != datastore.ErrNoSuchEntity {
			return err
		}

		// time.Time is stored with microsecond precision into datastore.
		// See https://cloud.google.com/appengine/docs/go/datastore/reference also.
		currentUTCTime := time.Now().UTC().Round(time.Microsecond)
		entity := &documentationStatusEntity{
			CommitID:    status.CommitID,
			TaskID:      status.TaskID,
			Status:      status.Status,
			RequestedAt: currentUTCTime,
			UpdatedAt:   currentUTCTime,
		}
		_, err = datastore.Put(store.context, key, entity)
		return err
	}, nil)
}

// GetStatus loads the documentation status for the documentation target.
func (store *DocumentationStore) GetStatus(target *data.DocumentationTarget, sequenceID int64) (*data.DocumentationStatus, error) {
	key, err := store.newDocumentationStatusKey(target, &data.DocumentationStatus{SequenceID: sequenceID})
	if err != nil {
		return nil, err
	}

	var entity documentationStatusEntity
	if err := datastore.Get(store.context, key, &entity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, data.ErrNoSuchStatus
		}

		return nil, err
	}

	status := entity.toDocumentationStatus()
	status.SequenceID = sequenceID
	return status, nil
}

// GetStatusMulti loads the documentation status for the multiple documentation targets of a given repository
func (store *DocumentationStore) GetStatusMulti(targets []*data.DocumentationTarget) (map[*data.DocumentationTarget][]*data.DocumentationStatus, error) {
	targets2StatusArrays := make(map[*data.DocumentationTarget][]*data.DocumentationStatus)

	for _, target := range targets {
		statusArray, err := store.getDocumentationStatusesByTarget(target)
		if err != nil {
			return nil, err
		}
		targets2StatusArrays[target] = statusArray
	}

	return targets2StatusArrays, nil
}

// UpdateStatus updates the existing documentation status with the given one.
func (store *DocumentationStore) UpdateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus) error {
	return store.updateStatus(target, status, func(entity *documentationStatusEntity) error {
		return nil
	})
}

// UpdateStatusWithCondition updates the existing documentation status with the given one
// if the current status is equal to prevStatus, or returns an error otherwise.
func (store *DocumentationStore) UpdateStatusWithCondition(target *data.DocumentationTarget, status *data.DocumentationStatus, prevStatus string) error {
	return store.updateStatus(target, status, func(entity *documentationStatusEntity) error {
		if entity.Status != prevStatus {
			return data.ErrUnsatisfiedStatusCondition
		}
		return nil
	})
}

func (store *DocumentationStore) updateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus, conditionCheckFunc func(*documentationStatusEntity) error) error {
	key, err := store.newDocumentationStatusKey(target, status)
	if err != nil {
		return err
	}

	return gcp.NewDefaultDatastoreTransaction().Run(store.context, func(tc context.Context) error {
		entity := new(documentationStatusEntity)
		err := datastore.Get(tc, key, entity)
		if err == datastore.ErrNoSuchEntity {
			return data.ErrNoSuchStatus
		}
		if err != nil {
			return err
		}
		if err := conditionCheckFunc(entity); err != nil {
			return err
		}

		// time.Time is stored with microsecond precision into datastore.
		// See https://cloud.google.com/appengine/docs/go/datastore/reference also.
		// And furthermore, RequestedAt should not be updated.
		entity.CommitID = status.CommitID
		entity.TaskID = status.TaskID
		entity.Status = status.Status
		entity.UpdatedAt = time.Now().UTC().Round(time.Microsecond)
		_, err = datastore.Put(store.context, key, entity)
		return err
	}, nil)
}

// QueryStatus returns a list of documentation status as the result of querying on the documentation target by the status.
func (store *DocumentationStore) QueryStatus(target *data.DocumentationTarget, status string) ([]*data.DocumentationStatus, error) {
	key, err := store.newDocumentationTargetKey(target, kindDocumentationStatus)
	if err != nil {
		return nil, err
	}

	var statuses []*data.DocumentationStatus
	err = gcp.NewDefaultDatastoreTransaction().Run(store.context, func(transactionContext context.Context) error {
		query := datastore.NewQuery(key.Kind()).Ancestor(key).Filter("Status =", status)
		var entities []documentationStatusEntity
		keys, err := query.GetAll(transactionContext, &entities)
		if err != nil {
			return err
		}

		for index, entity := range entities {
			docStatus := entity.toDocumentationStatus()
			docStatus.SequenceID = keys[index].IntID()
			statuses = append(statuses, docStatus)
		}

		return nil
	}, nil)

	return statuses, err
}
