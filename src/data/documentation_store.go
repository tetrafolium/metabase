package data

import "fmt"

// Constants for documentation status.
const (
	DocumentationStatusQueuing   = "queuing"
	DocumentationStatusQueued    = "queued"
	DocumentationStatusStarted   = "started"
	DocumentationStatusSucceeded = "succeeded"
	DocumentationStatusFailed    = "failed"
	DocumentationStatusCancelled = "cancelled"
)

var (
	// ErrNoSuchTarget is returned when a given documentation target was not found.
	ErrNoSuchTarget = fmt.Errorf("no such documentation target")
	// ErrTargetAlreadyExists is returned when a given documentation target already existed
	// and then creating new documentation failed.
	ErrTargetAlreadyExists = fmt.Errorf("documentation target already exists")

	// ErrNoSuchStatus is returned when a given documentation status was not found.
	ErrNoSuchStatus = fmt.Errorf("no such documentation status")
	// ErrStatusAlreadyExists is returned when a given documentation status already existed
	// and then creating new status failed.
	ErrStatusAlreadyExists = fmt.Errorf("documentation status already exists")
	// ErrUnsatisfiedStatusCondition is returned when an attempt of updating documentation status
	// fails due to the unsatisfied condition.
	ErrUnsatisfiedStatusCondition = fmt.Errorf("unsatisfied documentation status condition")
)

// DocumentationStore is an interface for loading/saving documentation data.
type DocumentationStore interface {
	CreateTarget(target *DocumentationTarget) error
	DeleteTarget(target *DocumentationTarget) error
	ExistsTarget(target *DocumentationTarget) (bool, error)
	CountTargetsInRepository(id int64) (int, error)

	MarkDeletingTarget(target *DeletingTarget) error
	UnmarkDeletingTarget(target *DeletingTarget) error
	IsDeletingTarget(target *DeletingTarget) (bool, error)

	GetResult(target *DocumentationTarget) (*DocumentationResult, error)
	GetResultMulti(targets []*DocumentationTarget) (map[*DocumentationTarget]*DocumentationResult, error)
	PutDocument(target *DocumentationTarget, document *Document) error

	CreateStatus(target *DocumentationTarget, status *DocumentationStatus) error
	GetStatus(target *DocumentationTarget, sequenceID int64) (*DocumentationStatus, error)
	GetStatusMulti(targets []*DocumentationTarget) (map[*DocumentationTarget]*DocumentationStatus, error)
	UpdateStatus(target *DocumentationTarget, status *DocumentationStatus) error
	UpdateStatusWithCondition(target *DocumentationTarget, status *DocumentationStatus, prevStatus string) error
	QueryStatus(target *DocumentationTarget, status string) ([]*DocumentationStatus, error)
}
