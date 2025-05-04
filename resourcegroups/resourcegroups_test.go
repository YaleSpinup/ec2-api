package resourcegroups

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resourcegroups"
	"github.com/aws/aws-sdk-go/service/resourcegroups/resourcegroupsiface"
	"testing"

	"github.com/aws/aws-sdk-go/aws/session"
)

// MockResourceGroupsClient is a mock implementation of resourcegroupsiface.ResourceGroupsAPI
type MockResourceGroupsClient struct {
	resourcegroupsiface.ResourceGroupsAPI
	mockCreateGroup             func(*resourcegroups.CreateGroupInput) (*resourcegroups.CreateGroupOutput, error)
	mockListGroupsPages         func(*resourcegroups.ListGroupsInput, func(*resourcegroups.ListGroupsOutput, bool) bool) error
	mockGetGroup                func(*resourcegroups.GetGroupInput) (*resourcegroups.GetGroupOutput, error)
	mockListGroupResourcesPages func(*resourcegroups.ListGroupResourcesInput, func(*resourcegroups.ListGroupResourcesOutput, bool) bool) error
	mockDeleteGroup             func(*resourcegroups.DeleteGroupInput) (*resourcegroups.DeleteGroupOutput, error)
}

func (m *MockResourceGroupsClient) CreateGroup(input *resourcegroups.CreateGroupInput) (*resourcegroups.CreateGroupOutput, error) {
	if m.mockCreateGroup != nil {
		return m.mockCreateGroup(input)
	}
	return nil, nil
}

func (m *MockResourceGroupsClient) ListGroupsPages(input *resourcegroups.ListGroupsInput, fn func(*resourcegroups.ListGroupsOutput, bool) bool) error {
	if m.mockListGroupsPages != nil {
		return m.mockListGroupsPages(input, fn)
	}
	return nil
}

func (m *MockResourceGroupsClient) GetGroup(input *resourcegroups.GetGroupInput) (*resourcegroups.GetGroupOutput, error) {
	if m.mockGetGroup != nil {
		return m.mockGetGroup(input)
	}
	return nil, nil
}

func (m *MockResourceGroupsClient) ListGroupResourcesPages(input *resourcegroups.ListGroupResourcesInput, fn func(*resourcegroups.ListGroupResourcesOutput, bool) bool) error {
	if m.mockListGroupResourcesPages != nil {
		return m.mockListGroupResourcesPages(input, fn)
	}
	return nil
}

func (m *MockResourceGroupsClient) DeleteGroup(input *resourcegroups.DeleteGroupInput) (*resourcegroups.DeleteGroupOutput, error) {
	if m.mockDeleteGroup != nil {
		return m.mockDeleteGroup(input)
	}
	return nil, nil
}

func TestWithSession(t *testing.T) {
	sess := session.Must(session.NewSession())
	rg := New(WithSession(sess))

	if rg.session != sess {
		t.Error("WithSession option did not set the session correctly")
	}

	if rg.Service == nil {
		t.Error("Service was not initialized with session")
	}
}

func TestWithCredentials(t *testing.T) {
	cases := []struct {
		name   string
		key    string
		secret string
		token  string
		region string
	}{
		{
			name:   "with all credentials",
			key:    "test-key",
			secret: "test-secret",
			token:  "test-token",
			region: "us-east-1",
		},
		{
			name:   "without token",
			key:    "test-key",
			secret: "test-secret",
			token:  "",
			region: "us-west-2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rg := New(WithCredentials(tc.key, tc.secret, tc.token, tc.region))

			if rg.session == nil {
				t.Error("session was not created")
			}

			if rg.Service == nil {
				t.Error("Service was not initialized")
			}
		})
	}
}

func TestNewWithMultipleOptions(t *testing.T) {
	cases := []struct {
		name    string
		opts    []Option
		wantNil bool
	}{
		{
			name:    "no options",
			opts:    []Option{},
			wantNil: true,
		},
		{
			name: "with credentials",
			opts: []Option{
				WithCredentials("key", "secret", "", "us-east-1"),
			},
			wantNil: false,
		},
		{
			name: "with session",
			opts: []Option{
				WithSession(session.Must(session.NewSession())),
			},
			wantNil: false,
		},
		{
			name: "with both options",
			opts: []Option{
				WithSession(session.Must(session.NewSession())),
				WithCredentials("key", "secret", "", "us-east-1"),
			},
			wantNil: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rg := New(tc.opts...)

			if tc.wantNil {
				if rg.Service != nil {
					t.Error("Service should be nil when no options are provided")
				}
			} else {
				if rg.Service == nil {
					t.Error("Service should not be nil when options are provided")
				}
			}
		})
	}
}

func TestCreateGroup(t *testing.T) {
	cases := []struct {
		name          string
		input         CreateGroupInput
		mockResponse  *resourcegroups.CreateGroupOutput
		mockErr       error
		expectedError bool
	}{
		{
			name: "successful creation",
			input: CreateGroupInput{
				Name:        "test-group",
				Description: "test description",
				ResourceQuery: &resourcegroups.ResourceQuery{
					Type: aws.String("TAG_FILTERS_1_0"),
					Query: aws.String(`{
                        "ResourceTypeFilters": ["AWS::EC2::Instance"],
                        "TagFilters": [
                            {
                                "Key": "Environment",
                                "Values": ["Test"]
                            }
                        ]
                    }`),
				},
			},
			mockResponse: &resourcegroups.CreateGroupOutput{
				Group: &resourcegroups.Group{
					GroupArn:    aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group"),
					Name:        aws.String("test-group"),
					Description: aws.String("test description"),
				},
			},
			expectedError: false,
		},
		{
			name: "aws error",
			input: CreateGroupInput{
				Name: "test-group",
				ResourceQuery: &resourcegroups.ResourceQuery{
					Type:  aws.String("TAG_FILTERS_1_0"),
					Query: aws.String("{}"),
				},
			},
			mockErr:       fmt.Errorf("AWS error"),
			expectedError: true,
		},
		{
			name: "nil service",
			input: CreateGroupInput{
				Name: "test-group",
				ResourceQuery: &resourcegroups.ResourceQuery{
					Type:  aws.String("TAG_FILTERS_1_0"),
					Query: aws.String("{}"),
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockResourceGroupsClient{
				mockCreateGroup: func(input *resourcegroups.CreateGroupInput) (*resourcegroups.CreateGroupOutput, error) {
					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return tc.mockResponse, nil
				},
			}

			rg := &ResourceGroups{
				Service: mock,
			}

			if tc.name == "nil service" {
				rg.Service = nil
			}

			group, err := rg.CreateGroup(tc.input)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if group == nil {
					t.Error("expected group but got nil")
				}
			}
		})
	}
}

func TestListGroups(t *testing.T) {
	cases := []struct {
		name          string
		mockResponse  []*resourcegroups.Group
		mockErr       error
		expectedError bool
		nilService    bool
	}{
		{
			name: "successful listing - single page",
			mockResponse: []*resourcegroups.Group{
				{
					GroupArn:    aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group-1"),
					Name:        aws.String("test-group-1"),
					Description: aws.String("test description 1"),
				},
				{
					GroupArn:    aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group-2"),
					Name:        aws.String("test-group-2"),
					Description: aws.String("test description 2"),
				},
			},
			expectedError: false,
		},
		{
			name: "successful listing - multiple pages",
			mockResponse: []*resourcegroups.Group{
				{
					GroupArn: aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group-1"),
					Name:     aws.String("test-group-1"),
				},
				{
					GroupArn: aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group-2"),
					Name:     aws.String("test-group-2"),
				},
				{
					GroupArn: aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group-3"),
					Name:     aws.String("test-group-3"),
				},
			},
			expectedError: false,
		},
		{
			name:          "aws error",
			mockErr:       fmt.Errorf("AWS error"),
			expectedError: true,
		},
		{
			name:          "nil service",
			nilService:    true,
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockResourceGroupsClient{
				mockListGroupsPages: func(input *resourcegroups.ListGroupsInput, fn func(*resourcegroups.ListGroupsOutput, bool) bool) error {
					if tc.mockErr != nil {
						return tc.mockErr
					}

					if len(tc.mockResponse) > 2 {
						// Simulate pagination for responses with more than 2 items
						fn(&resourcegroups.ListGroupsOutput{
							Groups: tc.mockResponse[:2],
						}, false)
						fn(&resourcegroups.ListGroupsOutput{
							Groups: tc.mockResponse[2:],
						}, true)
					} else {
						fn(&resourcegroups.ListGroupsOutput{
							Groups: tc.mockResponse,
						}, true)
					}
					return nil
				},
			}

			rg := &ResourceGroups{
				Service: mock,
			}

			if tc.nilService {
				rg.Service = nil
			}

			groups, err := rg.ListGroups()

			if tc.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if groups == nil {
					t.Error("expected groups but got nil")
				}
				if len(groups) != len(tc.mockResponse) {
					t.Errorf("expected %d groups but got %d", len(tc.mockResponse), len(groups))
				}
				// Verify each group matches the mock response
				for i, group := range groups {
					if *group.GroupArn != *tc.mockResponse[i].GroupArn {
						t.Errorf("expected group ARN %s but got %s", *tc.mockResponse[i].GroupArn, *group.GroupArn)
					}
					if *group.Name != *tc.mockResponse[i].Name {
						t.Errorf("expected group name %s but got %s", *tc.mockResponse[i].Name, *group.Name)
					}
				}
			}
		})
	}
}

func TestGetGroup(t *testing.T) {
	cases := []struct {
		name          string
		groupName     string
		mockResponse  *resourcegroups.GetGroupOutput
		mockErr       error
		expectedError bool
		nilService    bool
	}{
		{
			name:      "successful retrieval",
			groupName: "test-group",
			mockResponse: &resourcegroups.GetGroupOutput{
				Group: &resourcegroups.Group{
					GroupArn:    aws.String("arn:aws:resource-groups:us-east-1:123456789012:group/test-group"),
					Name:        aws.String("test-group"),
					Description: aws.String("test description"),
				},
			},
			expectedError: false,
		},
		{
			name:          "group not found",
			groupName:     "non-existent-group",
			mockErr:       fmt.Errorf("ResourceNotFoundException: Group not found"),
			expectedError: true,
		},
		{
			name:          "aws error",
			groupName:     "test-group",
			mockErr:       fmt.Errorf("AWS error"),
			expectedError: true,
		},
		{
			name:          "nil service",
			groupName:     "test-group",
			nilService:    true,
			expectedError: true,
		},
		{
			name:          "empty group name",
			groupName:     "",
			mockErr:       fmt.Errorf("ValidationException: Group name is required"),
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockResourceGroupsClient{
				mockGetGroup: func(input *resourcegroups.GetGroupInput) (*resourcegroups.GetGroupOutput, error) {
					// Verify input
					if input.GroupName == nil {
						t.Error("GroupName should not be nil")
					} else if *input.GroupName != tc.groupName {
						t.Errorf("expected group name %s but got %s", tc.groupName, *input.GroupName)
					}

					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return tc.mockResponse, nil
				},
			}

			rg := &ResourceGroups{
				Service: mock,
			}

			if tc.nilService {
				rg.Service = nil
			}

			group, err := rg.GetGroup(tc.groupName)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if group == nil {
					t.Error("expected group but got nil")
				} else {
					// Verify group details
					expectedGroup := tc.mockResponse.Group
					if *group.GroupArn != *expectedGroup.GroupArn {
						t.Errorf("expected group ARN %s but got %s", *expectedGroup.GroupArn, *group.GroupArn)
					}
					if *group.Name != *expectedGroup.Name {
						t.Errorf("expected group name %s but got %s", *expectedGroup.Name, *group.Name)
					}
					if *group.Description != *expectedGroup.Description {
						t.Errorf("expected group description %s but got %s", *expectedGroup.Description, *group.Description)
					}
				}
			}
		})
	}
}

func TestListGroupResources(t *testing.T) {
	cases := []struct {
		name          string
		groupName     string
		mockResponse  []*resourcegroups.ListGroupResourcesItem
		mockErr       error
		expectedError bool
		nilService    bool
	}{
		{
			name:      "successful listing - single page",
			groupName: "test-group",
			mockResponse: []*resourcegroups.ListGroupResourcesItem{
				{
					Identifier: &resourcegroups.ResourceIdentifier{
						ResourceArn:  aws.String("arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0"),
						ResourceType: aws.String("AWS::EC2::Instance"),
					},
				},
				{
					Identifier: &resourcegroups.ResourceIdentifier{
						ResourceArn:  aws.String("arn:aws:ec2:us-east-1:123456789012:instance/i-0987654321fedcba0"),
						ResourceType: aws.String("AWS::EC2::Instance"),
					},
				},
			},
			expectedError: false,
		},
		{
			name:      "successful listing - multiple pages",
			groupName: "test-group",
			mockResponse: []*resourcegroups.ListGroupResourcesItem{
				{
					Identifier: &resourcegroups.ResourceIdentifier{
						ResourceArn:  aws.String("arn:aws:ec2:us-east-1:123456789012:instance/i-1"),
						ResourceType: aws.String("AWS::EC2::Instance"),
					},
				},
				{
					Identifier: &resourcegroups.ResourceIdentifier{
						ResourceArn:  aws.String("arn:aws:ec2:us-east-1:123456789012:instance/i-2"),
						ResourceType: aws.String("AWS::EC2::Instance"),
					},
				},
				{
					Identifier: &resourcegroups.ResourceIdentifier{
						ResourceArn:  aws.String("arn:aws:s3:us-east-1:123456789012:bucket/test-bucket"),
						ResourceType: aws.String("AWS::S3::Bucket"),
					},
				},
			},
			expectedError: false,
		},
		{
			name:          "group not found",
			groupName:     "non-existent-group",
			mockErr:       fmt.Errorf("ResourceNotFoundException: Group not found"),
			expectedError: true,
		},
		{
			name:          "aws error",
			groupName:     "test-group",
			mockErr:       fmt.Errorf("AWS error"),
			expectedError: true,
		},
		{
			name:          "nil service",
			groupName:     "test-group",
			nilService:    true,
			expectedError: true,
		},
		{
			name:          "empty group name",
			groupName:     "",
			mockErr:       fmt.Errorf("ValidationException: Group name is required"),
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockResourceGroupsClient{
				mockListGroupResourcesPages: func(input *resourcegroups.ListGroupResourcesInput, fn func(*resourcegroups.ListGroupResourcesOutput, bool) bool) error {
					// Verify input
					if input.GroupName == nil {
						t.Error("GroupName should not be nil")
					} else if *input.GroupName != tc.groupName {
						t.Errorf("expected group name %s but got %s", tc.groupName, *input.GroupName)
					}

					if tc.mockErr != nil {
						return tc.mockErr
					}

					if len(tc.mockResponse) > 2 {
						// Simulate pagination for responses with more than 2 items
						fn(&resourcegroups.ListGroupResourcesOutput{
							Resources: tc.mockResponse[:2],
						}, false)
						fn(&resourcegroups.ListGroupResourcesOutput{
							Resources: tc.mockResponse[2:],
						}, true)
					} else {
						fn(&resourcegroups.ListGroupResourcesOutput{
							Resources: tc.mockResponse,
						}, true)
					}
					return nil
				},
			}

			rg := &ResourceGroups{
				Service: mock,
			}

			if tc.nilService {
				rg.Service = nil
			}

			resources, err := rg.ListGroupResources(tc.groupName)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resources == nil {
					t.Error("expected resources but got nil")
				}
				if len(resources) != len(tc.mockResponse) {
					t.Errorf("expected %d resources but got %d", len(tc.mockResponse), len(resources))
				}
				// Verify each resource matches the mock response
				for i, resource := range resources {
					if *resource.Identifier.ResourceArn != *tc.mockResponse[i].Identifier.ResourceArn {
						t.Errorf("expected resource ARN %s but got %s",
							*tc.mockResponse[i].Identifier.ResourceArn,
							*resource.Identifier.ResourceArn)
					}
					if *resource.Identifier.ResourceType != *tc.mockResponse[i].Identifier.ResourceType {
						t.Errorf("expected resource type %s but got %s",
							*tc.mockResponse[i].Identifier.ResourceType,
							*resource.Identifier.ResourceType)
					}
				}
			}
		})
	}
}

func TestDeleteGroup(t *testing.T) {
	cases := []struct {
		name          string
		groupName     string
		mockResponse  *resourcegroups.DeleteGroupOutput
		mockErr       error
		expectedError bool
		nilService    bool
	}{
		{
			name:          "successful deletion",
			groupName:     "test-group",
			mockResponse:  &resourcegroups.DeleteGroupOutput{},
			expectedError: false,
		},
		{
			name:          "group not found",
			groupName:     "non-existent-group",
			mockErr:       fmt.Errorf("ResourceNotFoundException: Group not found"),
			expectedError: true,
		},
		{
			name:          "aws error",
			groupName:     "test-group",
			mockErr:       fmt.Errorf("AWS error"),
			expectedError: true,
		},
		{
			name:          "nil service",
			groupName:     "test-group",
			nilService:    true,
			expectedError: true,
		},
		{
			name:          "empty group name",
			groupName:     "",
			mockErr:       fmt.Errorf("ValidationException: Group name is required"),
			expectedError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mock := &MockResourceGroupsClient{
				mockDeleteGroup: func(input *resourcegroups.DeleteGroupInput) (*resourcegroups.DeleteGroupOutput, error) {
					// Verify input
					if input.GroupName == nil {
						t.Error("GroupName should not be nil")
					} else if *input.GroupName != tc.groupName {
						t.Errorf("expected group name %s but got %s", tc.groupName, *input.GroupName)
					}

					if tc.mockErr != nil {
						return nil, tc.mockErr
					}
					return tc.mockResponse, nil
				},
			}

			rg := &ResourceGroups{
				Service: mock,
			}

			if tc.nilService {
				rg.Service = nil
			}

			err := rg.DeleteGroup(tc.groupName)

			if tc.expectedError {
				if err == nil {
					t.Error("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}
