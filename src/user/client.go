package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

var (
	// ErrNoSuchUser is returned when a given tractrix user did not exist.
	ErrNoSuchUser = fmt.Errorf("no such user")
)

type userAccount struct {
	ID      string `json:"id"`
	Service string `json:"service"`
}

type userManagerCreateRequest struct {
	Account userAccount `json:"account"`
}

type userManagerResponse struct {
	ID      string      `json:"id"`
	Account userAccount `json:"account"`
}

// ManagerClient is a client that communicates with user-manager servers.
// It implements UserManager interface.
type ManagerClient struct {
	client *http.Client
	apiURL *url.URL
}

// NewManagerClient returns a new client that hits the apiURL provided by user-manager servers.
func NewManagerClient(client *http.Client, apiURL *url.URL) (*ManagerClient, error) {
	if apiURL == nil {
		return nil, fmt.Errorf("no apiURL specified")
	}

	ManagerClient := &ManagerClient{
		apiURL: apiURL,
	}

	if client == nil {
		ManagerClient.client = http.DefaultClient
	} else {
		ManagerClient.client = client
	}

	return ManagerClient, nil
}

// GetID requests an user-manager server to return TUID of specified service user ID
func (client *ManagerClient) GetID(service string, id string) (uint64, error) {
	if service == "" || id == "" {
		return 0, fmt.Errorf("service and id must not be empty")
	}

	getIDEndpoint := fmt.Sprintf("%susers/%s/%s", client.apiURL.String(), service, id)
	resp, err := client.client.Get(getIDEndpoint)
	if err != nil {
		return 0, err
	}

	response := userManagerResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return 0, ErrNoSuchUser
		}
		return 0, fmt.Errorf("error received from user-manager: %d", resp.StatusCode)
	}

	tuid, err := strconv.ParseUint(response.ID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid response from user manager: %v", err)
	}

	return tuid, nil
}

// CreateID requests an user-manager server to create TUID for specified service user ID
func (client *ManagerClient) CreateID(service string, id string) (uint64, error) {
	if service == "" || id == "" {
		return 0, fmt.Errorf("service and id must not be empty")
	}

	body := bytes.NewBuffer(nil)
	err := json.NewEncoder(body).Encode(userManagerCreateRequest{
		Account: userAccount{
			ID:      id,
			Service: service,
		}})
	if err != nil {
		return 0, err
	}

	createIDEndpoint := fmt.Sprintf("%susers", client.apiURL.String())
	resp, err := client.client.Post(createIDEndpoint, "application/json", body)
	if err != nil {
		return 0, err
	}

	response := userManagerResponse{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("error received from user-manager: %d", resp.StatusCode)
	}

	tuid, err := strconv.ParseUint(response.ID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid response from user manager: %v", err)
	}

	return tuid, nil
}
