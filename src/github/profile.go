package github

import "strconv"

// GetUserID gets github user ID
func (service *Service) GetUserID() (string, error) {
	user, resp, err := service.client.Users.Get("")
	if err != nil {
		return "", service.translateErrorResponse(resp, err)
	}

	return strconv.Itoa(*user.ID), nil
}
