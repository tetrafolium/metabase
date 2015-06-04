package github

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"

	"github.com/tractrix/common-go/repository"
)

func (service *Service) obtainDeployKeyID(deployKey *repository.DeployKey) (int, error) {
	if err := repository.ValidateDeployKey(deployKey); err != nil {
		return 0, err
	}

	keyID, err := strconv.Atoi(deployKey.ID)
	if err != nil {
		return 0, fmt.Errorf("invalid deploy key id: %s", err.Error())
	}

	return keyID, nil
}

// RegisterDeployKey registers the deploy key with the repository.
func (service *Service) RegisterDeployKey(slug *repository.Slug, publicKey, title string) (*repository.DeployKey, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return nil, err
	}

	key := &github.Key{
		Key:   github.String(publicKey),
		Title: github.String(title),
	}
	key, resp, err := service.client.Repositories.CreateKey(slug.Owner, slug.Name, key)
	if err != nil {
		return nil, service.translateErrorResponse(resp, err)
	}
	if key.ID == nil {
		return nil, fmt.Errorf("no deploy key id obtained")
	}

	deployKey := &repository.DeployKey{
		ID: strconv.Itoa(*key.ID),
	}

	return deployKey, nil
}

// UnregisterDeployKey unregisters the deploy key from the repository.
func (service *Service) UnregisterDeployKey(slug *repository.Slug, deployKey *repository.DeployKey) error {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return err
	}

	keyID, err := service.obtainDeployKeyID(deployKey)
	if err != nil {
		return err
	}

	resp, err := service.client.Repositories.DeleteKey(slug.Owner, slug.Name, keyID)
	return service.translateErrorResponse(resp, err)
}

// ExistsDeployKey returns true if the deploy key exists on the repository.
// It returns false otherwise.
func (service *Service) ExistsDeployKey(slug *repository.Slug, deployKey *repository.DeployKey) (bool, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return false, err
	}

	keyID, err := service.obtainDeployKeyID(deployKey)
	if err != nil {
		return false, err
	}

	_, resp, err := service.client.Repositories.GetKey(slug.Owner, slug.Name, keyID)
	if err != nil {
		if resp.Response.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
