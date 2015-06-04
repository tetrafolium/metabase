package datatest

import (
	"fmt"
	"sync"
	"time"

	"github.com/tractrix/common-go/repository"
	"github.com/tractrix/docstand/server/data"
)

type documentationResultEntity struct {
	result data.DocumentationResult
}

func (entity *documentationResultEntity) clone() *documentationResultEntity {
	clonedEntity := new(documentationResultEntity)
	if entity.result.Documents != nil {
		clonedEntity.result.Documents = data.MakeDocumentMapFromDocuments(entity.result.Documents.ListDocuments())
	}
	return clonedEntity
}

type documentationStatusEntity struct {
	status data.DocumentationStatus
}

func (entity *documentationStatusEntity) clone() *documentationStatusEntity {
	clonedEntity := new(documentationStatusEntity)
	clonedEntity.status = entity.status
	return clonedEntity
}

type deletingTargetEntity struct {
	RequestedAt time.Time
}

func (entity *deletingTargetEntity) clone() *deletingTargetEntity {
	clonedEntity := *entity
	return &clonedEntity
}

// DocumentationStorage represents a on-memory storage where documentation data is stored.
type DocumentationStorage struct {
	mutex           sync.RWMutex
	results         map[int64]map[repository.Reference]*documentationResultEntity
	status          map[int64]map[repository.Reference]map[int64]*documentationStatusEntity
	deletingTargets map[int64]map[repository.Reference]*deletingTargetEntity
}

var defaultDocumentationStorage = NewDocumentationStorage()

// DefaultDocumentationStorage returns the default on-memory documentation data storage used by default.
func DefaultDocumentationStorage() *DocumentationStorage {
	return defaultDocumentationStorage
}

// NewDocumentationStorage returns a new DocumentationStorage.
func NewDocumentationStorage() *DocumentationStorage {
	return &DocumentationStorage{
		results:         make(map[int64]map[repository.Reference]*documentationResultEntity),
		status:          make(map[int64]map[repository.Reference]map[int64]*documentationStatusEntity),
		deletingTargets: make(map[int64]map[repository.Reference]*deletingTargetEntity),
	}
}

// ResetResources resets all resources held in this package.
func ResetResources() {
	defaultDocumentationStorage.Clear()
}

// Clear deletes all documentation data from the storage.
func (storage *DocumentationStorage) Clear() {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	storage.results = make(map[int64]map[repository.Reference]*documentationResultEntity)
	storage.status = make(map[int64]map[repository.Reference]map[int64]*documentationStatusEntity)
	storage.deletingTargets = make(map[int64]map[repository.Reference]*deletingTargetEntity)
}

func (storage *DocumentationStorage) getResultEntity(target *data.DocumentationTarget) (*documentationResultEntity, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	if target == nil {
		return nil, fmt.Errorf("mock: no documentation target specified")
	}

	refs2Entities, ok := storage.results[target.RepositoryID]
	if !ok {
		return nil, data.ErrNoSuchTarget
	}

	entity, ok := refs2Entities[target.Reference]
	if !ok {
		return nil, data.ErrNoSuchTarget
	}

	return entity.clone(), nil
}

func (storage *DocumentationStorage) listResultEntitiesInRepository(id int64) (map[repository.Reference]*documentationResultEntity, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	if id == 0 {
		return nil, fmt.Errorf("mock: no repository id specified")
	}

	refs2Entities, ok := storage.results[id]
	if !ok {
		return nil, nil
	}

	entities := make(map[repository.Reference]*documentationResultEntity)
	for ref, entity := range refs2Entities {
		entities[ref] = entity.clone()
	}

	return entities, nil
}

func (storage *DocumentationStorage) putResultEntity(target *data.DocumentationTarget, entity *documentationResultEntity) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if target == nil {
		return fmt.Errorf("mock: no documentation target specified")
	}
	if entity == nil {
		return fmt.Errorf("mock: no documentation result entity specified")
	}

	refs2Entities, ok := storage.results[target.RepositoryID]
	if !ok {
		refs2Entities = make(map[repository.Reference]*documentationResultEntity)
		storage.results[target.RepositoryID] = refs2Entities
	}

	refs2Entities[target.Reference] = entity.clone()

	return nil
}

func (storage *DocumentationStorage) deleteResultEntity(target *data.DocumentationTarget) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if target == nil {
		return fmt.Errorf("mock: no documentation target specified")
	}

	refs2Entities, ok := storage.results[target.RepositoryID]
	if !ok {
		return nil
	}

	_, ok = refs2Entities[target.Reference]
	if !ok {
		return nil
	}

	delete(refs2Entities, target.Reference)

	return nil
}

func (storage *DocumentationStorage) getStatusEntity(target *data.DocumentationTarget, sequenceID int64) (*documentationStatusEntity, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	if target == nil {
		return nil, fmt.Errorf("mock: no documentation target specified")
	}

	refs2SeqIDs, ok := storage.status[target.RepositoryID]
	if !ok {
		return nil, data.ErrNoSuchStatus
	}

	seqIDs2Entities, ok := refs2SeqIDs[target.Reference]
	if !ok {
		return nil, data.ErrNoSuchStatus
	}

	entity, ok := seqIDs2Entities[sequenceID]
	if !ok {
		return nil, data.ErrNoSuchStatus
	}

	return entity.clone(), nil
}

func (storage *DocumentationStorage) listStatusEntitiesInTarget(target *data.DocumentationTarget) (map[int64]*documentationStatusEntity, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	if target == nil {
		return nil, fmt.Errorf("mock: no documentation target specified")
	}

	refs2SeqIDs, ok := storage.status[target.RepositoryID]
	if !ok {
		return nil, nil
	}

	seqIDs2Entities, ok := refs2SeqIDs[target.Reference]
	if !ok {
		return nil, nil
	}

	entities := make(map[int64]*documentationStatusEntity)
	for sequenceID, entity := range seqIDs2Entities {
		entities[sequenceID] = entity.clone()
	}

	return entities, nil
}

func (storage *DocumentationStorage) putStatusEntity(target *data.DocumentationTarget, entity *documentationStatusEntity) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if target == nil {
		return fmt.Errorf("mock: no documentation target specified")
	}
	if entity == nil {
		return fmt.Errorf("mock: no documentation status entity specified")
	}

	refs2SeqIDs, ok := storage.status[target.RepositoryID]
	if !ok {
		refs2SeqIDs = make(map[repository.Reference]map[int64]*documentationStatusEntity)
		storage.status[target.RepositoryID] = refs2SeqIDs
	}

	seqIDs2Entities, ok := refs2SeqIDs[target.Reference]
	if !ok {
		seqIDs2Entities = make(map[int64]*documentationStatusEntity)
		refs2SeqIDs[target.Reference] = seqIDs2Entities
	}

	seqIDs2Entities[entity.status.SequenceID] = entity.clone()

	return nil
}

func (storage *DocumentationStorage) deleteStatusEntity(target *data.DocumentationTarget, sequenceID int64) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if target == nil {
		return fmt.Errorf("mock: no documentation target specified")
	}

	refs2SeqIDs, ok := storage.status[target.RepositoryID]
	if !ok {
		return nil
	}

	seqIDs2Entities, ok := refs2SeqIDs[target.Reference]
	if !ok {
		return nil
	}

	_, ok = seqIDs2Entities[sequenceID]
	if !ok {
		return nil
	}

	delete(seqIDs2Entities, sequenceID)

	return nil
}

func (storage *DocumentationStorage) getDeletingTargetEntity(target *data.DeletingTarget) (*deletingTargetEntity, error) {
	storage.mutex.RLock()
	defer storage.mutex.RUnlock()

	if target == nil {
		return nil, fmt.Errorf("mock: no documentation target specified")
	}

	refs2Entities, ok := storage.deletingTargets[target.RepositoryID]
	if !ok {
		return nil, data.ErrNoSuchTarget
	}

	entity, ok := refs2Entities[target.Reference]
	if !ok {
		return nil, data.ErrNoSuchTarget
	}

	return entity.clone(), nil
}

func (storage *DocumentationStorage) putDeletingTargetEntity(target *data.DeletingTarget, entity *deletingTargetEntity) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if target == nil {
		return fmt.Errorf("mock: no documentation target specified")
	}
	if entity == nil {
		return fmt.Errorf("mock: no deleting target entity specified")
	}

	refs2Entities, ok := storage.deletingTargets[target.RepositoryID]
	if !ok {
		refs2Entities = make(map[repository.Reference]*deletingTargetEntity)
		storage.deletingTargets[target.RepositoryID] = refs2Entities
	}

	refs2Entities[target.Reference] = entity.clone()

	return nil
}

func (storage *DocumentationStorage) deleteDeletingTargetEntity(target *data.DeletingTarget) error {
	storage.mutex.Lock()
	defer storage.mutex.Unlock()

	if target == nil {
		return fmt.Errorf("mock: no documentation target specified")
	}

	refs2Entities, ok := storage.deletingTargets[target.RepositoryID]
	if !ok {
		return nil
	}

	_, ok = refs2Entities[target.Reference]
	if !ok {
		return nil
	}

	delete(refs2Entities, target.Reference)

	return nil
}

// DocumentationStoreDelegate holds functions where DocumentationStore delegates to.
type DocumentationStoreDelegate struct {
	CreateTarget             func(target *data.DocumentationTarget) error
	DeleteTarget             func(target *data.DocumentationTarget) error
	ExistsTarget             func(target *data.DocumentationTarget) (bool, error)
	CountTargetsInRepository func(id int64) (int, error)

	MarkDeletingTarget   func(target *data.DeletingTarget) error
	UnmarkDeletingTarget func(target *data.DeletingTarget) error
	IsDeletingTarget     func(target *data.DeletingTarget) (bool, error)

	GetResult      func(target *data.DocumentationTarget) (*data.DocumentationResult, error)
	GetResultMulti func(targets []*data.DocumentationTarget) (map[*data.DocumentationTarget]*data.DocumentationResult, error)
	PutDocument    func(target *data.DocumentationTarget, document *data.Document) error

	CreateStatus              func(target *data.DocumentationTarget, status *data.DocumentationStatus) error
	GetStatus                 func(target *data.DocumentationTarget, sequenceID int64) (*data.DocumentationStatus, error)
	UpdateStatus              func(target *data.DocumentationTarget, status *data.DocumentationStatus) error
	UpdateStatusWithCondition func(target *data.DocumentationTarget, status *data.DocumentationStatus, prevStatus string) error
	QueryStatus               func(target *data.DocumentationTarget, status string) ([]*data.DocumentationStatus, error)
}

// DocumentationStore is a mock of documentation store for testing.
// DocumentationStore delegates all invocations to Delegate,
// so that the behavior of DocumentationStore can be customized by updating Delegate.
// DocumentationStore holds all data on-memory by default.
type DocumentationStore struct {
	mutex    sync.RWMutex
	storage  *DocumentationStorage
	Delegate DocumentationStoreDelegate
}

// NewDocumentationStore returns a new mock object of documentation store.
// All mock objects created by this function share the default on-memory storage.
func NewDocumentationStore() *DocumentationStore {
	return NewDocumentationStoreWithStorage(defaultDocumentationStorage)
}

// NewDocumentationStoreWithStorage returns a new mock object of documentation store
// that stores documentation data to the given storage.
func NewDocumentationStoreWithStorage(storage *DocumentationStorage) *DocumentationStore {
	store := &DocumentationStore{
		storage: storage,
	}
	store.Delegate = DocumentationStoreDelegate{
		CreateTarget:             store.defaultCreateTarget,
		DeleteTarget:             store.defaultDeleteTarget,
		ExistsTarget:             store.defaultExistsTarget,
		CountTargetsInRepository: store.defaultCountTargetsInRepository,

		MarkDeletingTarget:   store.defaultMarkDeletingTarget,
		UnmarkDeletingTarget: store.defaultUnmarkDeletingTarget,
		IsDeletingTarget:     store.defaultIsDeletingTarget,

		GetResult:      store.defaultGetResult,
		GetResultMulti: store.defaultGetResultMulti,
		PutDocument:    store.defaultPutDocument,

		CreateStatus:              store.defaultCreateStatus,
		GetStatus:                 store.defaultGetStatus,
		UpdateStatus:              store.defaultUpdateStatus,
		UpdateStatusWithCondition: store.defaultUpdateStatusWithCondition,
		QueryStatus:               store.defaultQueryStatus,
	}
	return store
}

// Storage returns the storage associated with the documentation store.
func (store *DocumentationStore) Storage() *DocumentationStorage {
	return store.storage
}

// CreateTarget invokes Delegate.CreateTarget.
func (store *DocumentationStore) CreateTarget(target *data.DocumentationTarget) error {
	return store.Delegate.CreateTarget(target)
}

// defaultCreateTarget saves a new entity for the documentation target.
// If the documentation target already exists, it returns data.ErrTargetAlreadyExists.
func (store *DocumentationStore) defaultCreateTarget(target *data.DocumentationTarget) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	_, err := store.storage.getResultEntity(target)
	if err == nil {
		return data.ErrTargetAlreadyExists
	}
	if err != data.ErrNoSuchTarget {
		return err
	}

	return store.storage.putResultEntity(target, new(documentationResultEntity))
}

// DeleteTarget invokes Delegate.DeleteTarget.
func (store *DocumentationStore) DeleteTarget(target *data.DocumentationTarget) error {
	return store.Delegate.DeleteTarget(target)
}

// defaultDeleteTarget deletes the documentation target.
func (store *DocumentationStore) defaultDeleteTarget(target *data.DocumentationTarget) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if err := store.storage.deleteResultEntity(target); err != nil {
		return err
	}

	statusEntities, err := store.storage.listStatusEntitiesInTarget(target)
	if err != nil {
		return err
	}

	for sequenceID := range statusEntities {
		if err := store.storage.deleteStatusEntity(target, sequenceID); err != nil {
			return err
		}
	}

	return nil
}

// ExistsTarget invokes Delegate.ExistsTarget.
func (store *DocumentationStore) ExistsTarget(target *data.DocumentationTarget) (bool, error) {
	return store.Delegate.ExistsTarget(target)
}

// defaultExistsTarget returns true if the documentation target already exists.
// It returns false otherwise.
func (store *DocumentationStore) defaultExistsTarget(target *data.DocumentationTarget) (bool, error) {
	_, err := store.storage.getResultEntity(target)
	switch err {
	case nil:
		return true, nil
	case data.ErrNoSuchTarget:
		return false, nil
	}
	return false, err
}

// CountTargetsInRepository invokes Delegate.CountTargetsInRepository.
func (store *DocumentationStore) CountTargetsInRepository(id int64) (int, error) {
	return store.Delegate.CountTargetsInRepository(id)
}

// defaultCountTargetsInRepository returns the number of documentation targets in the repository.
func (store *DocumentationStore) defaultCountTargetsInRepository(id int64) (int, error) {
	entities, err := store.storage.listResultEntitiesInRepository(id)
	if err != nil {
		return 0, err
	}

	return len(entities), nil
}

// MarkDeletingTarget invokes Delegate.MarkDeletingTarget.
func (store *DocumentationStore) MarkDeletingTarget(target *data.DeletingTarget) error {
	return store.Delegate.MarkDeletingTarget(target)
}

// defaultMarkDeletingTarget marks the documentation target as deleting.
func (store *DocumentationStore) defaultMarkDeletingTarget(target *data.DeletingTarget) error {
	return store.storage.putDeletingTargetEntity(target, &deletingTargetEntity{
		RequestedAt: time.Now().UTC(),
	})
}

// UnmarkDeletingTarget invokes Delegate.UnmarkDeletingTarget.
func (store *DocumentationStore) UnmarkDeletingTarget(target *data.DeletingTarget) error {
	return store.Delegate.UnmarkDeletingTarget(target)
}

// defaultUnmarkDeletingTarget removes a mark of deleting from the documentation target.
func (store *DocumentationStore) defaultUnmarkDeletingTarget(target *data.DeletingTarget) error {
	return store.storage.deleteDeletingTargetEntity(target)
}

// IsDeletingTarget invokes Delegate.IsDeletingTarget.
func (store *DocumentationStore) IsDeletingTarget(target *data.DeletingTarget) (bool, error) {
	return store.Delegate.IsDeletingTarget(target)
}

// defaultIsDeletingTarget checks if the target is marked as a target being deleted.
func (store *DocumentationStore) defaultIsDeletingTarget(target *data.DeletingTarget) (bool, error) {
	entity, err := store.storage.getDeletingTargetEntity(target)
	if err == data.ErrNoSuchTarget {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return (entity != nil), nil
}

// GetResult invokes Delegate.GetResult.
func (store *DocumentationStore) GetResult(target *data.DocumentationTarget) (*data.DocumentationResult, error) {
	return store.Delegate.GetResult(target)
}

// defaultGetResult loads the latest documentation information for the documentation target.
func (store *DocumentationStore) defaultGetResult(target *data.DocumentationTarget) (*data.DocumentationResult, error) {
	entity, err := store.storage.getResultEntity(target)
	if err != nil {
		return nil, err
	}

	return &entity.result, nil
}

// GetResultMulti invokes Delegate.GetResultMulti.
func (store *DocumentationStore) GetResultMulti(targets []*data.DocumentationTarget) (map[*data.DocumentationTarget]*data.DocumentationResult, error) {
	return store.Delegate.GetResultMulti(targets)
}

// defaultGetResultMulti is a batch version of GetLatestResult.
func (store *DocumentationStore) defaultGetResultMulti(targets []*data.DocumentationTarget) (map[*data.DocumentationTarget]*data.DocumentationResult, error) {
	store.mutex.RLock()
	defer store.mutex.RUnlock()

	results := make(map[*data.DocumentationTarget]*data.DocumentationResult)
	for _, target := range targets {
		entity, err := store.storage.getResultEntity(target)
		switch err {
		case nil:
			results[target] = &entity.result
		case data.ErrNoSuchTarget:
			continue
		default:
			return nil, err
		}
	}

	return results, nil
}

// PutDocument invokes Delegate.PutDocument.
func (store *DocumentationStore) PutDocument(target *data.DocumentationTarget, document *data.Document) error {
	return store.Delegate.PutDocument(target, document)
}

// defaultPutDocument saves the document information as the latest one for the documentation target.
func (store *DocumentationStore) defaultPutDocument(target *data.DocumentationTarget, document *data.Document) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	entity, err := store.storage.getResultEntity(target)
	if err != nil {
		return err
	}

	if document == nil {
		return fmt.Errorf("mock: no document specified")
	}

	if entity.result.Documents == nil {
		entity.result.Documents = data.MakeDocumentMapFromDocuments([]data.Document{*document})
	} else {
		entity.result.Documents[document.Type] = document.CommitID
	}

	return store.storage.putResultEntity(target, entity)
}

// CreateStatus invokes Delegate.CreateStatus.
func (store *DocumentationStore) CreateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus) error {
	return store.Delegate.CreateStatus(target, status)
}

// defaultCreateStatus saves a new entity for the documentation status.
func (store *DocumentationStore) defaultCreateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if status == nil {
		return fmt.Errorf("mock: no documentation status specified")
	}

	entity, err := store.storage.getStatusEntity(target, status.SequenceID)
	if err == nil {
		return data.ErrStatusAlreadyExists
	}
	if err != data.ErrNoSuchStatus {
		return err
	}

	entity = &documentationStatusEntity{status: *status}
	entity.status.RequestedAt = time.Now().UTC()
	entity.status.UpdatedAt = entity.status.RequestedAt
	return store.storage.putStatusEntity(target, entity)
}

// GetStatus invokes Delegate.GetStatus.
func (store *DocumentationStore) GetStatus(target *data.DocumentationTarget, sequenceID int64) (*data.DocumentationStatus, error) {
	return store.Delegate.GetStatus(target, sequenceID)
}

// defaultGetStatus loads the documentation status for the documentation target.
func (store *DocumentationStore) defaultGetStatus(target *data.DocumentationTarget, sequenceID int64) (*data.DocumentationStatus, error) {
	entity, err := store.storage.getStatusEntity(target, sequenceID)
	if err != nil {
		return nil, err
	}

	return &entity.status, nil
}

// UpdateStatus invokes Delegate.UpdateStatus.
func (store *DocumentationStore) UpdateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus) error {
	return store.Delegate.UpdateStatus(target, status)
}

// defaultUpdateStatus updates the existing documentation status with the given one.
func (store *DocumentationStore) defaultUpdateStatus(target *data.DocumentationTarget, status *data.DocumentationStatus) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if status == nil {
		return fmt.Errorf("mock: no documentation status specified")
	}

	entity, err := store.storage.getStatusEntity(target, status.SequenceID)
	if err != nil {
		return err
	}

	entity.status.CommitID = status.CommitID
	entity.status.TaskID = status.TaskID
	entity.status.Status = status.Status
	entity.status.UpdatedAt = time.Now().UTC()
	return store.storage.putStatusEntity(target, entity)
}

// UpdateStatusWithCondition invokes Delegate.UpdateStatusWithCondition.
func (store *DocumentationStore) UpdateStatusWithCondition(target *data.DocumentationTarget, status *data.DocumentationStatus, prevStatus string) error {
	return store.Delegate.UpdateStatusWithCondition(target, status, prevStatus)
}

// defaultUpdateStatusWithCondition updates the existing documentation status with the given one
// if the current status is equal to prevStatus, or returns an error otherwise.
func (store *DocumentationStore) defaultUpdateStatusWithCondition(target *data.DocumentationTarget, status *data.DocumentationStatus, prevStatus string) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	if status == nil {
		return fmt.Errorf("mock: no documentation status specified")
	}

	entity, err := store.storage.getStatusEntity(target, status.SequenceID)
	if err != nil {
		return err
	}
	if entity.status.Status != prevStatus {
		return data.ErrUnsatisfiedStatusCondition
	}

	entity.status.CommitID = status.CommitID
	entity.status.TaskID = status.TaskID
	entity.status.Status = status.Status
	entity.status.UpdatedAt = time.Now().UTC()
	return store.storage.putStatusEntity(target, entity)
}

// QueryStatus invokes Delegate.QueryStatus.
func (store *DocumentationStore) QueryStatus(target *data.DocumentationTarget, status string) ([]*data.DocumentationStatus, error) {
	return store.Delegate.QueryStatus(target, status)
}

// defaultQueryStatus returns a list of documentation status as the result of querying on the documentation target by the status.
func (store *DocumentationStore) defaultQueryStatus(target *data.DocumentationTarget, status string) ([]*data.DocumentationStatus, error) {
	entities, err := store.storage.listStatusEntitiesInTarget(target)
	if err != nil {
		return nil, err
	}

	var statusList []*data.DocumentationStatus
	for _, entity := range entities {
		if entity.status.Status == status {
			statusList = append(statusList, &entity.status)
		}
	}

	return statusList, nil
}
