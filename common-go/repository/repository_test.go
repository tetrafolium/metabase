package repository

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type SlugTestSuite struct {
	suite.Suite
}

func Test_SlugTestSuite(t *testing.T) {
	suite.Run(t, new(SlugTestSuite))
}

func (suite *SlugTestSuite) Test_ValidateSlug_WithValidSlug() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "docstand",
	}
	err := ValidateSlug(slug, slug.Saas)
	suite.Assert().NoError(err)
}

func (suite *SlugTestSuite) Test_ValidateSlug_WithInvalidSlug() {
	slug := &Slug{}
	err := ValidateSlug(slug, slug.Saas)
	suite.Assert().Error(err)
}

func (suite *SlugTestSuite) Test_ValidateSlug_WithNil() {
	err := ValidateSlug(nil, "")
	suite.Assert().Error(err)
}

func (suite *SlugTestSuite) Test_Validate_WithValidFields() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "docstand",
	}
	err := slug.Validate(slug.Saas)
	suite.Assert().NoError(err)
}

func (suite *SlugTestSuite) Test_Validate_WithEmptySaas() {
	slug := &Slug{
		Saas:  "",
		Owner: "tractrix",
		Name:  "docstand",
	}
	err := slug.Validate(slug.Saas)
	suite.Assert().Error(err)
}

func (suite *SlugTestSuite) Test_Validate_WithUnexpectedSaas() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "docstand",
	}
	err := slug.Validate("bitbucket.org")
	suite.Assert().Error(err)
}

func (suite *SlugTestSuite) Test_Validate_WithEmptyOwner() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "",
		Name:  "docstand",
	}
	err := slug.Validate(slug.Saas)
	suite.Assert().Error(err)
}

func (suite *SlugTestSuite) Test_Validate_WithEmptyName() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "",
	}
	err := slug.Validate(slug.Saas)
	suite.Assert().Error(err)
}

func (suite *SlugTestSuite) Test_Hash_ForSameSlug() {
	slug := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "docstand",
	}
	expected := slug.Hash()
	actual := slug.Hash()
	suite.Assert().Equal(expected, actual, "hash value should be always same")
}

func (suite *SlugTestSuite) Test_Hash_ForDifferentSlugs() {
	slug1 := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "docstand",
	}
	slug2 := &Slug{
		Saas:  "github.com",
		Owner: "tractrix",
		Name:  "common-go",
	}
	actual1 := slug1.Hash()
	actual2 := slug2.Hash()
	suite.Assert().NotEqual(actual1, actual2, "hash values calculated for different slugs should differ")
}

type IDTestSuite struct {
	suite.Suite
}

func Test_IDTestSuite(t *testing.T) {
	suite.Run(t, new(IDTestSuite))
}

func (suite *IDTestSuite) Test_ValidateID_WithValidID() {
	id := &ID{
		Saas:    "github.com",
		OwnerID: "100",
		ID:      "200",
	}
	err := ValidateID(id, id.Saas)
	suite.Assert().NoError(err)
}

func (suite *IDTestSuite) Test_ValidateID_WithInvalidID() {
	id := &ID{}
	err := ValidateID(id, id.Saas)
	suite.Assert().Error(err)
}

func (suite *IDTestSuite) Test_ValidateID_WithNil() {
	err := ValidateID(nil, "")
	suite.Assert().Error(err)
}

func (suite *IDTestSuite) Test_Validate_WithValidFields() {
	id := &ID{
		Saas:    "github.com",
		OwnerID: "100",
		ID:      "200",
	}
	err := id.Validate(id.Saas)
	suite.Assert().NoError(err)
}

func (suite *IDTestSuite) Test_Validate_WithEmptySaas() {
	id := &ID{
		Saas:    "",
		OwnerID: "100",
		ID:      "200",
	}
	err := id.Validate(id.Saas)
	suite.Assert().Error(err)
}

func (suite *IDTestSuite) Test_Validate_WithUnexpectedSaas() {
	id := &ID{
		Saas:    "github.com",
		OwnerID: "100",
		ID:      "200",
	}
	err := id.Validate("bitbucket.org")
	suite.Assert().Error(err)
}

func (suite *IDTestSuite) Test_Validate_WithEmptyOwnerID() {
	id := &ID{
		Saas:    "github.com",
		OwnerID: "",
		ID:      "200",
	}
	err := id.Validate(id.Saas)
	suite.Assert().Error(err)
}

func (suite *IDTestSuite) Test_Validate_WithEmptyID() {
	id := &ID{
		Saas:    "github.com",
		OwnerID: "100",
		ID:      "",
	}
	err := id.Validate(id.Saas)
	suite.Assert().Error(err)
}

func (suite *IDTestSuite) Test_Hash_ForSameID() {
	id := &ID{
		Saas:    "github.com",
		OwnerID: "100",
		ID:      "200",
	}
	expected := id.Hash()
	actual := id.Hash()
	suite.Assert().Equal(expected, actual, "hash value should be always same")
}

func (suite *IDTestSuite) Test_Hash_ForDifferentIDs() {
	id1 := &ID{
		Saas:    "github.com",
		OwnerID: "100",
		ID:      "200",
	}
	id2 := &ID{
		Saas:    "github.com",
		OwnerID: "101",
		ID:      "201",
	}
	actual1 := id1.Hash()
	actual2 := id2.Hash()
	suite.Assert().NotEqual(actual1, actual2, "hash values calculated for different ids should differ")
}

type ReferenceTestSuite struct {
	suite.Suite
}

func Test_ReferenceTestSuite(t *testing.T) {
	suite.Run(t, new(ReferenceTestSuite))
}

func (suite *ReferenceTestSuite) Test_ReferenceTypeBranch() {
	// Pay much attention if ReferenceTypeBranch needs to be changed.
	suite.Assert().Equal("branch", ReferenceTypeBranch)
}

func (suite *ReferenceTestSuite) Test_ReferenceTypeTag() {
	// Pay much attention if ReferenceTypeBranch needs to be changed.
	suite.Assert().Equal("tag", ReferenceTypeTag)
}

func (suite *ReferenceTestSuite) Test_ValidateReference_WithValidReference() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	err := ValidateReference(ref)
	suite.Assert().NoError(err)
}

func (suite *ReferenceTestSuite) Test_ValidateReference_WithInvalidReference() {
	ref := &Reference{}
	err := ValidateReference(ref)
	suite.Assert().Error(err)
}

func (suite *ReferenceTestSuite) Test_ValidateReference_WithNil() {
	err := ValidateReference(nil)
	suite.Assert().Error(err)
}

func (suite *ReferenceTestSuite) Test_Validate_WithValidBranch() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	err := ref.Validate()
	suite.Assert().NoError(err)
}

func (suite *ReferenceTestSuite) Test_Validate_WithValidTag() {
	ref := &Reference{
		Type: ReferenceTypeTag,
		Name: "0.0.1",
	}
	err := ref.Validate()
	suite.Assert().NoError(err)
}

func (suite *ReferenceTestSuite) Test_Validate_WithUnknownType() {
	ref := &Reference{
		Type: "unknown",
		Name: "master",
	}
	err := ref.Validate()
	suite.Assert().Error(err)
}

func (suite *ReferenceTestSuite) Test_Validate_WithEmptyType() {
	ref := &Reference{
		Type: "",
		Name: "master",
	}
	err := ref.Validate()
	suite.Assert().Error(err)
}

func (suite *ReferenceTestSuite) Test_Validate_WithEmptyName() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "",
	}
	err := ref.Validate()
	suite.Assert().Error(err)
}

func (suite *ReferenceTestSuite) Test_Hash_ForSameID() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	expected := ref.Hash()
	actual := ref.Hash()
	suite.Assert().Equal(expected, actual, "hash value should be always same")
}

func (suite *ReferenceTestSuite) Test_Hash_ForDifferentIDs() {
	ref1 := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	ref2 := &Reference{
		Type: ReferenceTypeTag,
		Name: "0.0.1",
	}
	actual1 := ref1.Hash()
	actual2 := ref2.Hash()
	suite.Assert().NotEqual(actual1, actual2, "hash values calculated for different references should differ")
}

func (suite *ReferenceTestSuite) Test_IsBranch_WithBranchType() {
	ref := &Reference{
		Type: ReferenceTypeBranch,
		Name: "master",
	}
	suite.Assert().True(ref.IsBranch(), "true should be returned for %s", ReferenceTypeBranch)
}

func (suite *ReferenceTestSuite) Test_IsBranch_WithNonBranchType() {
	nonBranchTypes := []string{ReferenceTypeTag, "unknown-type", ""}
	for _, nonBranchType := range nonBranchTypes {
		ref := &Reference{
			Type: nonBranchType,
			Name: "master",
		}
		suite.Assert().False(ref.IsBranch(), "false should be returned for non branch type: %s", nonBranchType)
	}
}

func (suite *ReferenceTestSuite) Test_IsTag_WithTagType() {
	ref := &Reference{
		Type: ReferenceTypeTag,
		Name: "master",
	}
	suite.Assert().True(ref.IsTag(), "true should be returned for %s", ReferenceTypeTag)
}

func (suite *ReferenceTestSuite) Test_IsTag_WithNonTagType() {
	nonTagTypes := []string{ReferenceTypeBranch, "unknown-type", ""}
	for _, nonTagType := range nonTagTypes {
		ref := &Reference{
			Type: nonTagType,
			Name: "master",
		}
		suite.Assert().False(ref.IsTag(), "false should be returned for non tag type: %s", nonTagType)
	}
}

type DeployKeyTestSuite struct {
	suite.Suite
}

func Test_DeployKeyTestSuite(t *testing.T) {
	suite.Run(t, new(DeployKeyTestSuite))
}

func (suite *DeployKeyTestSuite) Test_ValidateDeployKey_WithValidDeployKey() {
	deployKey := &DeployKey{
		ID: "200",
	}
	err := ValidateDeployKey(deployKey)
	suite.Assert().NoError(err)
}

func (suite *DeployKeyTestSuite) Test_ValidateDeployKey_WithInvalidDeployKey() {
	deployKey := &DeployKey{}
	err := ValidateDeployKey(deployKey)
	suite.Assert().Error(err)
}

func (suite *DeployKeyTestSuite) Test_ValidateDeployKey_WithNil() {
	err := ValidateDeployKey(nil)
	suite.Assert().Error(err)
}

func (suite *DeployKeyTestSuite) Test_Validate_WithValidFields() {
	deployKey := &DeployKey{
		ID: "200",
	}
	err := deployKey.Validate()
	suite.Assert().NoError(err)
}

func (suite *DeployKeyTestSuite) Test_Validate_WithEmptyID() {
	deployKey := &DeployKey{
		ID: "",
	}
	err := deployKey.Validate()
	suite.Assert().Error(err)
}

type WebhookTestSuite struct {
	suite.Suite
}

func Test_WebhookTestSuite(t *testing.T) {
	suite.Run(t, new(WebhookTestSuite))
}

func (suite *WebhookTestSuite) Test_ValidateWebhook_WithValidWebhook() {
	webhook := &Webhook{
		ID: "200",
	}
	err := ValidateWebhook(webhook)
	suite.Assert().NoError(err)
}

func (suite *WebhookTestSuite) Test_ValidateWebhook_WithInvalidWebhook() {
	webhook := &Webhook{}
	err := ValidateWebhook(webhook)
	suite.Assert().Error(err)
}

func (suite *WebhookTestSuite) Test_ValidateWebhook_WithNil() {
	err := ValidateWebhook(nil)
	suite.Assert().Error(err)
}

func (suite *WebhookTestSuite) Test_Validate_WithValidFields() {
	webhook := &Webhook{
		ID: "200",
	}
	err := webhook.Validate()
	suite.Assert().NoError(err)
}

func (suite *WebhookTestSuite) Test_Validate_WithEmptyID() {
	webhook := &Webhook{
		ID: "",
	}
	err := webhook.Validate()
	suite.Assert().Error(err)
}
