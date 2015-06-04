package github

import (
	"fmt"
	"strings"

	"github.com/google/go-github/github"

	"github.com/tractrix/common-go/repository"
)

const (
	refPrefixBranch = "refs/heads/"
	refPrefixTag    = "refs/tags/"
)

func (service *Service) listRefs(slug *repository.Slug, refType string) ([]github.Reference, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return nil, err
	}

	var totalRefs []github.Reference
	var opt github.ReferenceListOptions
	opt.Type = refType
	opt.PerPage = maxPageSize
	for {
		gitRefs, resp, err := service.client.Git.ListRefs(slug.Owner, slug.Name, &opt)
		if err != nil {
			return nil, service.translateErrorResponse(resp, err)
		}
		totalRefs = append(totalRefs, gitRefs...)

		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return totalRefs, nil
}

func (service *Service) branchFromGitReference(gitRef *github.Reference) *repository.Reference {
	if gitRef.Ref != nil && strings.HasPrefix(*gitRef.Ref, refPrefixBranch) {
		return &repository.Reference{
			Type: repository.ReferenceTypeBranch,
			Name: strings.TrimPrefix(*gitRef.Ref, refPrefixBranch),
		}
	}
	return nil
}

func (service *Service) tagFromGitReference(gitRef *github.Reference) *repository.Reference {
	if gitRef.Ref != nil && strings.HasPrefix(*gitRef.Ref, refPrefixTag) {
		return &repository.Reference{
			Type: repository.ReferenceTypeTag,
			Name: strings.TrimPrefix(*gitRef.Ref, refPrefixTag),
		}
	}
	return nil
}

// ListReferences retrieves all branches and tags in the repository.
func (service *Service) ListReferences(slug *repository.Slug) ([]*repository.Reference, error) {
	gitRefs, err := service.listRefs(slug, "")
	if err != nil {
		return nil, err
	}

	var totalRefs []*repository.Reference
	for _, gitRef := range gitRefs {
		if branch := service.branchFromGitReference(&gitRef); branch != nil {
			totalRefs = append(totalRefs, branch)
		} else if tag := service.tagFromGitReference(&gitRef); tag != nil {
			totalRefs = append(totalRefs, tag)
		}
	}

	return totalRefs, nil
}

// ListBranches retrieves all branches in the repository.
func (service *Service) ListBranches(slug *repository.Slug) ([]*repository.Reference, error) {
	gitRefs, err := service.listRefs(slug, "heads")
	if err != nil {
		return nil, err
	}

	var totalRefs []*repository.Reference
	for _, gitRef := range gitRefs {
		if branch := service.branchFromGitReference(&gitRef); branch != nil {
			totalRefs = append(totalRefs, branch)
		}
	}

	return totalRefs, nil
}

// ListTags retrieves all tags in the repository.
func (service *Service) ListTags(slug *repository.Slug) ([]*repository.Reference, error) {
	gitRefs, err := service.listRefs(slug, "tags")
	if err != nil {
		return nil, err
	}

	var totalRefs []*repository.Reference
	for _, gitRef := range gitRefs {
		if tag := service.tagFromGitReference(&gitRef); tag != nil {
			totalRefs = append(totalRefs, tag)
		}
	}

	return totalRefs, nil
}

// GetCommitID returns the commit ID for the repository and reference.
// If the reference points at a tag, it returns the commit ID of the tag itself.
// If the reference points at a branch, it returns the ID of HEAD commit on the branch.
func (service *Service) GetCommitID(slug *repository.Slug, ref *repository.Reference) (string, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return "", err
	}

	if err := repository.ValidateReference(ref); err != nil {
		return "", err
	}

	var refPath string
	if ref.IsTag() {
		refPath = fmt.Sprintf("tags/%s", ref.Name)
	} else {
		refPath = fmt.Sprintf("heads/%s", ref.Name)
	}

	gitRef, resp, err := service.client.Git.GetRef(slug.Owner, slug.Name, refPath)
	if err != nil {
		return "", service.translateErrorResponse(resp, err)
	}
	if gitRef.Object == nil || gitRef.Object.SHA == nil {
		return "", fmt.Errorf("no commit id obtained")
	}

	return *gitRef.Object.SHA, nil
}
