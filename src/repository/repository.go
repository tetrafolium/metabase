package repository

import (
	"fmt"
	"hash/fnv"
)

var (
	// ErrNotFound is returned when a resource was not found.
	ErrNotFound = fmt.Errorf("repository: not found")
)

// Slug represents a URL-friendly version of a repository name.
// See https://confluence.atlassian.com/display/BITBUCKET/What+is+a+Slug for more details.
type Slug struct {
	Saas  string
	Owner string
	Name  string
}

// ValidateSlug checks if the repository slug has valid value in each field.
// It is useful if slug is not sure to be non-nil pointer.
func ValidateSlug(slug *Slug, expectedSaas string) error {
	if slug == nil {
		return fmt.Errorf("no repository slug specified")
	}
	return slug.Validate(expectedSaas)
}

// Validate checks if the repository slug has valid value in each field.
func (slug *Slug) Validate(expectedSaas string) error {
	if slug.Saas == "" {
		return fmt.Errorf("no repository saas specified")
	}
	if slug.Saas != expectedSaas {
		return fmt.Errorf("unexpected repository saas: %s", slug.Saas)
	}
	if slug.Owner == "" {
		return fmt.Errorf("no repository owner name specified")
	}
	if slug.Name == "" {
		return fmt.Errorf("no repository name specified")
	}
	return nil
}

// Hash calculates a 64-bit FNV-1a hash value for the repository slug.
func (slug *Slug) Hash() uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(slug.Saas))
	hasher.Write([]byte(slug.Owner))
	hasher.Write([]byte(slug.Name))
	return hasher.Sum64()
}

// ID represents an ID for a repository.
type ID struct {
	Saas    string
	OwnerID string
	ID      string
}

// ValidateID checks if the repository ID has valid value in each field.
// It is useful if id is not sure to be non-nil pointer.
func ValidateID(id *ID, expectedSaas string) error {
	if id == nil {
		return fmt.Errorf("no repository id specified")
	}
	return id.Validate(expectedSaas)
}

// Validate checks if the repository ID has valid value in each field.
func (id *ID) Validate(expectedSaas string) error {
	if id.Saas == "" {
		return fmt.Errorf("no repository saas specified")
	}
	if id.Saas != expectedSaas {
		return fmt.Errorf("unexpected repository saas: %s", id.Saas)
	}
	if id.OwnerID == "" {
		return fmt.Errorf("no repository owner id specified")
	}
	if id.ID == "" {
		return fmt.Errorf("no repository id specified")
	}
	return nil
}

// Hash calculates a 64-bit FNV-1a hash value for the repository ID.
func (id *ID) Hash() uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(id.Saas))
	hasher.Write([]byte(id.OwnerID))
	hasher.Write([]byte(id.ID))
	return hasher.Sum64()
}

// Permissions holds a set of permissions for a user to do operations on a repository.
type Permissions struct {
	Admin bool
}

// Repository represents detailed information for a repository.
type Repository struct {
	Slug

	ID          ID
	Permissions Permissions
}

// Constants for reference types.
const (
	ReferenceTypeBranch = "branch"
	ReferenceTypeTag    = "tag"
)

// Reference represents a reference, such as a branch or a tag, to a commit.
type Reference struct {
	Type string
	Name string
}

// ValidateReference checks if the reference has valid value in each field.
// It is useful if ref is not sure to be non-nil pointer.
func ValidateReference(ref *Reference) error {
	if ref == nil {
		return fmt.Errorf("no reference specified")
	}
	return ref.Validate()
}

// Validate checks if the reference has valid value in each field.
func (ref *Reference) Validate() error {
	if ref.Type == "" {
		return fmt.Errorf("no reference type specified")
	}
	if !ref.IsBranch() && !ref.IsTag() {
		return fmt.Errorf("unknown reference type: %s", ref.Type)
	}
	if ref.Name == "" {
		return fmt.Errorf("no reference name specified")
	}
	return nil
}

// Hash calculates a 64-bit FNV-1a hash value for the repository slug.
func (ref *Reference) Hash() uint64 {
	hasher := fnv.New64a()
	hasher.Write([]byte(ref.Type))
	hasher.Write([]byte(ref.Name))
	return hasher.Sum64()
}

// IsBranch returns true if the reference represents a branch.
// It returns false otherwise.
func (ref *Reference) IsBranch() bool {
	return ref.Type == ReferenceTypeBranch
}

// IsTag returns true if the reference represents a tag.
// It returns false otherwise.
func (ref *Reference) IsTag() bool {
	return ref.Type == ReferenceTypeTag
}

// DeployKey represents a deploy key.
type DeployKey struct {
	ID string
}

// ValidateDeployKey checks if the deploy key has valid value in each field.
// It is useful if deployKey is not sure to be non-nil pointer.
func ValidateDeployKey(deployKey *DeployKey) error {
	if deployKey == nil {
		return fmt.Errorf("no deploy key specified")
	}
	return deployKey.Validate()
}

// Validate checks if the deploy key has valid value in each field.
func (deployKey *DeployKey) Validate() error {
	if deployKey.ID == "" {
		return fmt.Errorf("no deploy key id specified")
	}
	return nil
}

// Webhook represents a webhook.
type Webhook struct {
	ID string
}

// ValidateWebhook checks if the webhook has valid value in each field.
// It is useful if webhook is not sure to be non-nil pointer.
func ValidateWebhook(webhook *Webhook) error {
	if webhook == nil {
		return fmt.Errorf("no webhook specified")
	}
	return webhook.Validate()
}

// Validate checks if the webhook has valid value in each field.
func (webhook *Webhook) Validate() error {
	if webhook.ID == "" {
		return fmt.Errorf("no webhook id specified")
	}
	return nil
}

// PushEvent represents a payload of an event triggered when a reference is pushed to a remote repository.
type PushEvent struct {
	RepositoryID int64

	ID        ID
	Slug      Slug
	Reference Reference
	CommitID  string
}

// Service is an interface for communicating with a repository hosting service.
type Service interface {
	GetServiceName() string

	GetUserID() (string, error)

	GetRepositoryBySlug(slug *Slug) (*Repository, error)
	GetRepositoryByID(id *ID) (*Repository, error)
	ListRepositories() ([]*Repository, error)

	ListReferences(slug *Slug) ([]*Reference, error)
	ListBranches(slug *Slug) ([]*Reference, error)
	ListTags(slug *Slug) ([]*Reference, error)
	GetCommitID(slug *Slug, ref *Reference) (string, error)

	RegisterDeployKey(slug *Slug, publicKey, title string) (*DeployKey, error)
	UnregisterDeployKey(slug *Slug, deployKey *DeployKey) error
	ExistsDeployKey(slug *Slug, deployKey *DeployKey) (bool, error)

	RegisterWebhook(slug *Slug, hookURL string) (*Webhook, error)
	UnregisterWebhook(slug *Slug, webhook *Webhook) error
	ExistsWebhook(slug *Slug, webhook *Webhook) (bool, error)
}
