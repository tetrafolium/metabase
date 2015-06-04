package data

import (
	"time"

	"github.com/tractrix/common-go/repository"
)

// Document represents a generated document.
type Document struct {
	Type     string
	CommitID string
}

// DocumentMap represents a map of types and commit IDs for generated documents.
type DocumentMap map[string]string

// MakeDocumentMapFromDocuments returns a new DocumentMap having document information specified.
// If there are multiple documents having the same type with different commit IDs in the list specified,
// only the document at the largest index appears in the result.
func MakeDocumentMapFromDocuments(documents []Document) DocumentMap {
	docMap := make(DocumentMap)

	for _, document := range documents {
		docMap[document.Type] = document.CommitID
	}

	return docMap
}

// ListDocuments returns a list of document information being held in this map.
func (docMap DocumentMap) ListDocuments() []Document {
	var documents []Document

	for docType, commitID := range docMap {
		documents = append(documents, Document{
			Type:     docType,
			CommitID: commitID,
		})
	}

	return documents
}

// DocumentationTarget represents a target for documentation.
type DocumentationTarget struct {
	RepositoryID int64
	Reference    repository.Reference
}

// DocumentationResult represents documents generated for a documentation target.
type DocumentationResult struct {
	Documents DocumentMap
}

// DocumentationStatus represents documentation status for a documentation target.
type DocumentationStatus struct {
	// A sequential documentation ID which is unique in a repository including branches and tags.
	SequenceID int64
	// A commit ID to be documented.
	CommitID string
	// An ID for the documentation task.
	// This is used only for cancelling the queued task when the documentation target is unregistered.
	TaskID string
	// Status represents the current status of the documentation task.
	Status string
	// A datetime when a documentation request is accepted.
	// This should be automatically set by DocumentationStore, so that specifying this is ignored
	// when updating the status.
	RequestedAt time.Time
	// A datetime when the status is updated most recently.
	// This should be automatically set by DocumentationStore, so that specifying this is ignored
	// when updating the status.
	UpdatedAt time.Time
}

// DeletingTarget represents a documentation target being deleted.
type DeletingTarget struct {
	RepositoryID int64
	Reference    repository.Reference
}
