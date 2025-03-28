package ssm

import (
	"context"
	"reflect"
	"testing"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// mockSSMAPI is a mock implementation of ssmiface.SSMAPI
type mockSSMAPI struct {
	ssmiface.SSMAPI
	// DescribeInstanceInformationWithContextOutput is the output that will be returned by DescribeInstanceInformationWithContext
	DescribeInstanceInformationWithContextOutput *ssm.DescribeInstanceInformationOutput
	// DescribeInstanceInformationWithContextError is the error that will be returned by DescribeInstanceInformationWithContext
	DescribeInstanceInformationWithContextError error
}

// DescribeInstanceInformationWithContext mocks the DescribeInstanceInformationWithContext method
func (m *mockSSMAPI) DescribeInstanceInformationWithContext(ctx aws.Context, input *ssm.DescribeInstanceInformationInput, opts ...request.Option) (*ssm.DescribeInstanceInformationOutput, error) {
	if m.DescribeInstanceInformationWithContextError != nil {
		return nil, m.DescribeInstanceInformationWithContextError
	}
	return m.DescribeInstanceInformationWithContextOutput, nil
}

// TestGetInstanceInformationWithFilters tests the GetInstanceInformationWithFilters function
func TestGetInstanceInformationWithFilters(t *testing.T) {
	// Test case 1: Basic instance information retrieval
	mockOutput := &ssm.DescribeInstanceInformationOutput{
		InstanceInformationList: []*ssm.InstanceInformation{
			{
				InstanceId:   aws.String("i-1234567890abcdef0"),
				PingStatus:   aws.String("Online"),
				AgentVersion: aws.String("3.0.0.0"),
			},
		},
	}

	s := SSM{
		Service: &mockSSMAPI{
			DescribeInstanceInformationWithContextOutput: mockOutput,
		},
	}

	instances, err := s.GetInstanceInformationWithFilters(context.Background(), map[string]string{
		"InstanceIds": "i-1234567890abcdef0",
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if !reflect.DeepEqual(instances, mockOutput.InstanceInformationList) {
		t.Errorf("Got %+v, expected %+v", instances, mockOutput.InstanceInformationList)
	}

	// Test case 2: Error from SSM service
	s = SSM{
		Service: &mockSSMAPI{
			DescribeInstanceInformationWithContextError: apierror.New(apierror.ErrInternalError, "SSM service error", nil),
		},
	}

	_, err = s.GetInstanceInformationWithFilters(context.Background(), map[string]string{
		"InstanceIds": "i-1234567890abcdef0",
	})

	if err == nil {
		t.Errorf("Expected error, but got none")
	}
}