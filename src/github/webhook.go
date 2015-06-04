package github

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/go-github/github"

	"github.com/tractrix/common-go/repository"
)

func (service *Service) obtainWebhookID(webhook *repository.Webhook) (int, error) {
	if err := repository.ValidateWebhook(webhook); err != nil {
		return 0, err
	}

	hookID, err := strconv.Atoi(webhook.ID)
	if err != nil {
		return 0, fmt.Errorf("invalid webhook id: %s", err.Error())
	}

	return hookID, nil
}

// RegisterWebhook registers the webhook URL with the repository.
func (service *Service) RegisterWebhook(slug *repository.Slug, hookURL string) (*repository.Webhook, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return nil, err
	}

	hook := &github.Hook{
		Name:   github.String("web"),
		Events: []string{"push"},
		Active: github.Bool(true),
		Config: map[string]interface{}{
			"url":          hookURL,
			"content_type": "json",
		},
	}
	hook, resp, err := service.client.Repositories.CreateHook(slug.Owner, slug.Name, hook)
	if err != nil {
		return nil, service.translateErrorResponse(resp, err)
	}
	if hook.ID == nil {
		return nil, fmt.Errorf("no webhook id obtained")
	}

	webhook := &repository.Webhook{
		ID: strconv.Itoa(*hook.ID),
	}

	return webhook, nil
}

// UnregisterWebhook unregisters the webhook from the repository.
func (service *Service) UnregisterWebhook(slug *repository.Slug, webhook *repository.Webhook) error {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return err
	}

	hookID, err := service.obtainWebhookID(webhook)
	if err != nil {
		return err
	}

	resp, err := service.client.Repositories.DeleteHook(slug.Owner, slug.Name, hookID)
	return service.translateErrorResponse(resp, err)
}

// ExistsWebhook returns true if the webhook exists on the repository.
// It returns false otherwise.
func (service *Service) ExistsWebhook(slug *repository.Slug, webhook *repository.Webhook) (bool, error) {
	if err := repository.ValidateSlug(slug, ServiceName); err != nil {
		return false, err
	}

	hookID, err := service.obtainWebhookID(webhook)
	if err != nil {
		return false, err
	}

	_, resp, err := service.client.Repositories.GetHook(slug.Owner, slug.Name, hookID)
	if err != nil {
		if resp.Response.StatusCode == http.StatusNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
