package github

import (
	"strconv"
	"strings"

	"github.com/google/go-github/github"

	"github.com/tractrix/common-go/repository"
)

// PushEvent represents own github push event payload struct.
// This is defined here due to a bug in go-github package
// https://github.com/google/go-github/issues/131
type PushEvent struct {
	HeadCommit   *PushEventCommit     `json:"head_commit,omitempty"`
	Forced       *bool                `json:"forced,omitempty"`
	Created      *bool                `json:"created,omitempty"`
	Deleted      *bool                `json:"deleted,omitempty"`
	Ref          *string              `json:"ref,omitempty"`
	Before       *string              `json:"before,omitempty"`
	After        *string              `json:"after,omitempty"`
	Compare      *string              `json:"compare,omitempty"`
	Size         *int                 `json:"size,omitempty"`
	Commits      []PushEventCommit    `json:"commits,omitempty"`
	Owner        *github.User         `json:"owner,omitempty"`
	Organization *github.Organization `json:"organization,omitempty"`
	Repo         *struct {
		ID       *int    `json:"id,omitempty"`
		FullName *string `json:"full_name,omitempty"`
	} `json:"repository,omitempty"`
}

// PushEventCommit represents information about commits included in push event
type PushEventCommit struct {
	ID       *string              `json:"id,omitempty"`
	Message  *string              `json:"message,omitempty"`
	Author   *github.CommitAuthor `json:"author,omitempty"`
	URL      *string              `json:"url,omitempty"`
	Distinct *bool                `json:"distinct,omitempty"`
	Added    []string             `json:"added,omitempty"`
	Removed  []string             `json:"removed,omitempty"`
	Modified []string             `json:"modified,omitempty"`
}

// IsDeleteEvent chceks if the event is branch/tag delete event
func (event *PushEvent) IsDeleteEvent() bool {
	return event.Deleted != nil && *event.Deleted
}

// ToRepositoryPushEvent converts GitHub push event payloads into repository push event
func (event *PushEvent) ToRepositoryPushEvent() *repository.PushEvent {
	var repoEvent repository.PushEvent

	repoEvent.ID = *event.toRepositoryID()
	repoEvent.Slug = *event.toRepositorySlug()
	repoEvent.Reference = *event.toRepositoryReference()

	if event.After != nil {
		repoEvent.CommitID = *event.After
	}

	return &repoEvent
}

func (event *PushEvent) toRepositoryID() *repository.ID {
	var id repository.ID

	id.Saas = ServiceName
	if event.Repo != nil && event.Repo.ID != nil {
		id.ID = strconv.Itoa(*event.Repo.ID)
	}
	if event.Organization != nil && event.Organization.ID != nil {
		id.OwnerID = strconv.Itoa(*event.Organization.ID)
	}

	return &id
}

func (event *PushEvent) toRepositorySlug() *repository.Slug {
	var slug repository.Slug

	slug.Saas = ServiceName
	if event.Repo != nil && event.Repo.FullName != nil {
		ownerRepo := strings.Split(*event.Repo.FullName, "/")
		if len(ownerRepo) == 2 {
			slug.Owner = ownerRepo[0]
			slug.Name = ownerRepo[1]
		}
	}

	return &slug
}

func (event *PushEvent) toRepositoryReference() *repository.Reference {
	var reference repository.Reference

	if event.Ref != nil {
		ref := strings.Split(*event.Ref, "/") // refs/heads|tags/<ref-name>
		if len(ref) == 3 && ref[0] == "refs" {
			if ref[1] == "heads" {
				reference.Type = repository.ReferenceTypeBranch
			} else {
				reference.Type = repository.ReferenceTypeTag
			}
			reference.Name = ref[2]
		}
	}

	return &reference
}
