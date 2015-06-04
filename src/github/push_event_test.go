package github

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tractrix/common-go/repository"
)

type ServicePushEventTestSuite struct {
	ServiceTestSuiteBase
}

func Test_ServicePushEventTestSuite(t *testing.T) {
	suite.Run(t, new(ServicePushEventTestSuite))
}

func (suite *ServicePushEventTestSuite) Test_ToRepositoryPushEvent_WithValidPushEvent() {
	payload := &PushEvent{}
	err := json.NewDecoder(bytes.NewBufferString(testPushEventPayloadJSON)).Decode(payload)
	suite.Assert().NoError(err)

	repoPushEvent := payload.ToRepositoryPushEvent()
	suite.Assert().NotNil(repoPushEvent)
	suite.Assert().Equal(testID, repoPushEvent.ID)
	suite.Assert().Equal(testSlug, repoPushEvent.Slug)
	suite.Assert().Equal(testReference, repoPushEvent.Reference)
	suite.Assert().Equal(testCommitID, repoPushEvent.CommitID)
}

func (suite *ServicePushEventTestSuite) Test_ToRepositoryPushEvent_WithValidTagPushEvent() {
	payload := &PushEvent{}
	err := json.NewDecoder(bytes.NewBufferString(testPushEventPayloadJSON)).Decode(payload)
	suite.Assert().NoError(err)

	refString := "refs/tag/tagname"
	payload.Ref = &refString
	expectedReference := repository.Reference{
		Type: repository.ReferenceTypeTag,
		Name: "tagname",
	}

	repoPushEvent := payload.ToRepositoryPushEvent()
	suite.Assert().NotNil(repoPushEvent)
	suite.Assert().Equal(testID, repoPushEvent.ID)
	suite.Assert().Equal(testSlug, repoPushEvent.Slug)
	suite.Assert().Equal(expectedReference, repoPushEvent.Reference)
	suite.Assert().Equal(testCommitID, repoPushEvent.CommitID)
}

func (suite *ServicePushEventTestSuite) Test_ToRepositoryPushEvent_WithEmptyPushEvent() {
	payload := &PushEvent{}

	repoPushEvent := payload.ToRepositoryPushEvent()
	suite.Assert().NotNil(repoPushEvent)
	suite.Assert().Equal(repository.ID{Saas: ServiceName}, repoPushEvent.ID)
	suite.Assert().Equal(repository.Slug{Saas: ServiceName}, repoPushEvent.Slug)
	suite.Assert().Equal(repository.Reference{}, repoPushEvent.Reference)
	suite.Assert().Equal("", repoPushEvent.CommitID)
}

func (suite *ServicePushEventTestSuite) Test_IsDeleteEvent_WithDeletePushEvent() {
	deleted := true
	payload := &PushEvent{
		Deleted: &deleted,
	}

	actual := payload.IsDeleteEvent()
	suite.Assert().Equal(deleted, actual)
}

func (suite *ServicePushEventTestSuite) Test_IsDeleteEvent_WithNonDeletePushEvent() {
	deleted := false
	payload := &PushEvent{
		Deleted: &deleted,
	}

	actual := payload.IsDeleteEvent()
	suite.Assert().Equal(deleted, actual)
}

func (suite *ServicePushEventTestSuite) Test_IsDeleteEvent_WithEmptyPushEvent() {
	payload := &PushEvent{}

	actual := payload.IsDeleteEvent()
	suite.Assert().Equal(false, actual)
}

var testID = repository.ID{
	Saas:    ServiceName,
	OwnerID: "100",
	ID:      "1000",
}

var testSlug = repository.Slug{
	Saas:  ServiceName,
	Owner: "testorg",
	Name:  "testrepo",
}

var testReference = repository.Reference{
	Type: repository.ReferenceTypeBranch,
	Name: "master",
}

var testCommitID = "2222222222222222222222222222222222222222"

var testPushEventPayloadJSON = `
{
  "ref": "refs/heads/master",
  "before": "1111111111111111111111111111111111111111",
  "after": "2222222222222222222222222222222222222222",
  "created": false,
  "deleted": false,
  "forced": false,
  "base_ref": null,
  "compare": "https://github.com/testorg/testrepo/compare/111111111111...222222222222",
  "commits": [{
      "id": "2222222222222222222222222222222222222222",
      "distinct": true,
      "message": "Update README.md",
      "timestamp": "2015-05-07T16:36:13+09:00",
      "url": "https://github.com/testorg/testrepo/commit/2222222222222222222222222222222222222222",
      "author": {
        "name": "Test User",
        "email": "testuser@users.noreply.github.com",
        "username": "testuser"
      },
      "committer": {
        "name": "Test User",
        "email": "testuser@users.noreply.github.com",
        "username": "testuser"
      },
      "added": [],
      "removed": [],
      "modified": ["README.md"]
    }],
  "head_commit": {
    "id": "2222222222222222222222222222222222222222",
    "distinct": true,
    "message": "Update README.md",
    "timestamp": "2015-05-07T16:36:13+09:00",
    "url": "https://github.com/testorg/testrepo/commit/2222222222222222222222222222222222222222",
    "author": {
      "name": "Test User",
      "email": "testuser@users.noreply.github.com",
      "username": "testuser"
    },
    "committer": {
      "name": "Test User",
      "email": "testuser@users.noreply.github.com",
      "username": "testuser"
    },
    "added": [],
    "removed": [],
    "modified": ["README.md"]
  },
  "repository": {
  	"id": 1000,
    "name": "testrepo",
    "full_name": "testorg/testrepo",
    "owner": {
      "name": "testorg",
      "email": ""
    },
    "private": true,
    "html_url": "https://github.com/testorg/testrepo",
    "description": "",
    "fork": false,
    "url": "https://github.com/testorg/testrepo",
    "created_at": 1429159784,
    "updated_at": "2015-05-07T07:36:13Z",
    "pushed_at": 1430984173,
    "git_url": "git://github.com/testorg/testrepo.git",
    "ssh_url": "git@github.com:testorg/testrepo.git",
    "clone_url": "https://github.com/testorg/testrepo.git",
    "language": "Go",
    "default_branch": "master",
    "master_branch": "master",
    "organization": "testorg"
  },
  "pusher": {
    "name": "testuser",
    "email": "testuser@users.noreply.github.com"
  },
  "organization": {
    "login": "testorg",
    "id": 100,
    "url": "https://api.github.com/orgs/testorg",
    "description": ""
  }
}
`
