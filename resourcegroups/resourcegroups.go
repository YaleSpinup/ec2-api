package resourcegroups

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroups/resourcegroupsiface"
	log "github.com/sirupsen/logrus"
)

// ResourceGroups is a wrapper around the aws Resource Groups service
type ResourceGroups struct {
	session *session.Session
	Service resourcegroupsiface.ResourceGroupsAPI
}

// CreateGroupInput represents the input parameters for creating a resource group
type CreateGroupInput struct {
	Name          string
	Description   string
	ResourceQuery *resourcegroups.ResourceQuery
}

// Option A resource group option
type Option func(*ResourceGroups)

// New creates a new ResourceGroups
func New(opts ...Option) *ResourceGroups {
	rg := ResourceGroups{}

	for _, opt := range opts {
		opt(&rg)
	}

	if rg.session != nil {
		rg.Service = resourcegroups.New(rg.session)
	}

	return &rg
}

// WithSession make sure the connection is using aws session functionality
func WithSession(sess *session.Session) Option {
	return func(rg *ResourceGroups) {
		log.Debug("using aws session")
		rg.session = sess
	}
}

// WithCredentials configures the resources group aws client with the given credentials
func WithCredentials(key, secret, token, region string) Option {
	return func(rg *ResourceGroups) {
		log.Debugf("creating new session with key id %s in region %s", key, region)
		sess := session.Must(session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(key, secret, token),
			Region:      aws.String(region),
		}))
		rg.session = sess
	}
}

// CreateGroup creates a new AWS resource group with the specified configuration
func (rg *ResourceGroups) CreateGroup(input CreateGroupInput) (*resourcegroups.Group, error) {
	if rg.Service == nil {
		return nil, fmt.Errorf("resource groups service not initialized")
	}

	params := &resourcegroups.CreateGroupInput{
		Name:          aws.String(input.Name),
		ResourceQuery: input.ResourceQuery,
	}

	if input.Description != "" {
		params.Description = aws.String(input.Description)
	}

	result, err := rg.Service.CreateGroup(params)
	if err != nil {
		return nil, err
	}

	return result.Group, nil
}

// ListGroups returns all resource groups in the account
func (rg *ResourceGroups) ListGroups() ([]*resourcegroups.Group, error) {
	if rg.Service == nil {
		return nil, fmt.Errorf("resource groups service not initialized")
	}

	input := &resourcegroups.ListGroupsInput{}
	var groups []*resourcegroups.Group

	// Handle pagination
	err := rg.Service.ListGroupsPages(input, func(page *resourcegroups.ListGroupsOutput, lastPage bool) bool {
		groups = append(groups, page.Groups...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return groups, nil
}

// GetGroup retrieves details of a specific resource group
func (rg *ResourceGroups) GetGroup(name string) (*resourcegroups.Group, error) {
	if rg.Service == nil {
		return nil, fmt.Errorf("resource groups service not initialized")
	}

	input := &resourcegroups.GetGroupInput{
		GroupName: aws.String(name),
	}

	result, err := rg.Service.GetGroup(input)
	if err != nil {
		return nil, err
	}

	return result.Group, nil
}

// ListGroupResources lists all resources in a specific group
func (rg *ResourceGroups) ListGroupResources(name string) ([]*resourcegroups.ListGroupResourcesItem, error) {
	if rg.Service == nil {
		return nil, fmt.Errorf("resource groups service not initialized")
	}

	input := &resourcegroups.ListGroupResourcesInput{
		GroupName: aws.String(name),
	}

	var resources []*resourcegroups.ListGroupResourcesItem

	// Handle pagination
	err := rg.Service.ListGroupResourcesPages(input, func(page *resourcegroups.ListGroupResourcesOutput, lastPage bool) bool {
		resources = append(resources, page.Resources...)
		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return resources, nil
}

// DeleteGroup deletes a resource group
func (rg *ResourceGroups) DeleteGroup(name string) error {
	if rg.Service == nil {
		return fmt.Errorf("resource groups service not initialized")
	}

	input := &resourcegroups.DeleteGroupInput{
		GroupName: aws.String(name),
	}

	_, err := rg.Service.DeleteGroup(input)
	return err
}
