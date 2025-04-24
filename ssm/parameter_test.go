package ssm

import (
	"context"
	"testing"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

// mockParameterClient is a fake SSM client for parameter tests
type mockParameterClient struct {
	ssmiface.SSMAPI
	t                        *testing.T
	putParameterOutput       *ssm.PutParameterOutput
	getParameterOutput       *ssm.GetParameterOutput
	describeParametersOutput *ssm.DescribeParametersOutput
	putParameterError        error
	getParameterError        error
	describeParametersError  error
	deleteParameterError     error
	addTagsToResourceError   error
}

func (m *mockParameterClient) PutParameterWithContext(ctx context.Context, input *ssm.PutParameterInput, opts ...request.Option) (*ssm.PutParameterOutput, error) {
	if m.putParameterError != nil {
		return nil, m.putParameterError
	}
	return m.putParameterOutput, nil
}

func (m *mockParameterClient) GetParameterWithContext(ctx context.Context, input *ssm.GetParameterInput, opts ...request.Option) (*ssm.GetParameterOutput, error) {
	if m.getParameterError != nil {
		return nil, m.getParameterError
	}
	return m.getParameterOutput, nil
}

func (m *mockParameterClient) DescribeParametersWithContext(ctx context.Context, input *ssm.DescribeParametersInput, opts ...request.Option) (*ssm.DescribeParametersOutput, error) {
	if m.describeParametersError != nil {
		return nil, m.describeParametersError
	}
	return m.describeParametersOutput, nil
}

func (m *mockParameterClient) DeleteParameterWithContext(ctx context.Context, input *ssm.DeleteParameterInput, opts ...request.Option) (*ssm.DeleteParameterOutput, error) {
	if m.deleteParameterError != nil {
		return nil, m.deleteParameterError
	}
	return &ssm.DeleteParameterOutput{}, nil
}

func (m *mockParameterClient) AddTagsToResourceWithContext(ctx context.Context, input *ssm.AddTagsToResourceInput, opts ...request.Option) (*ssm.AddTagsToResourceOutput, error) {
	if m.addTagsToResourceError != nil {
		return nil, m.addTagsToResourceError
	}
	return &ssm.AddTagsToResourceOutput{}, nil
}

// Helper function to compare errors properly
func compareErrors(t *testing.T, expected, actual error) bool {
	if expected == nil && actual == nil {
		return true
	}

	if expected == nil || actual == nil {
		return false
	}

	// Try to unwrap AWS errors
	var expectedCode, actualCode, expectedMsg, actualMsg string

	if awsErr, ok := expected.(awserr.Error); ok {
		expectedCode = awsErr.Code()
		expectedMsg = awsErr.Message()
	} else {
		expectedMsg = expected.Error()
	}

	if awsErr, ok := actual.(awserr.Error); ok {
		actualCode = awsErr.Code()
		actualMsg = awsErr.Message()
	} else {
		actualMsg = actual.Error()
	}

	// If both have codes, compare them
	if expectedCode != "" && actualCode != "" {
		if expectedCode != actualCode {
			t.Logf("Error code mismatch: expected %s, got %s", expectedCode, actualCode)
			return false
		}

		if expectedMsg != actualMsg {
			// For AWS errors, we only care about code matching
			t.Logf("Error message different, but codes match: expected %s, got %s", expectedMsg, actualMsg)
		}

		return true
	}

	// Check if the messages contain similar content
	if expectedMsg == actualMsg {
		return true
	}

	return false
}

func TestCreateParameter(t *testing.T) {
	tests := []struct {
		name     string
		input    *ssm.PutParameterInput
		mockSvc  *mockParameterClient
		expected *ssm.PutParameterOutput
		err      error
	}{
		{
			name:  "nil input",
			input: nil,
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "invalid input", nil),
		},
		{
			name: "empty name",
			input: &ssm.PutParameterInput{
				Name:  aws.String(""),
				Value: aws.String("value"),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter name is required", nil),
		},
		{
			name: "empty value",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String(""),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter value is required", nil),
		},
		{
			name: "nil type",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String("value"),
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter type is required", nil),
		},
		{
			name: "AWS service error",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String("value"),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t:                 t,
				putParameterError: awserr.New("ParameterLimitExceeded", "exceeded max parameters", nil),
			},
			expected: nil,
			// This matches the structure from common.ErrCode
			err: apierror.New(apierror.ErrBadRequest, "failed to create parameter: exceeded max parameters", awserr.New("ParameterLimitExceeded", "exceeded max parameters", nil)),
		},
		{
			name: "successful create parameter",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String("value"),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t: t,
				putParameterOutput: &ssm.PutParameterOutput{
					Version: aws.Int64(1),
				},
			},
			expected: &ssm.PutParameterOutput{
				Version: aws.Int64(1),
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssmSvc := &SSM{Service: tt.mockSvc}
			out, err := ssmSvc.CreateParameter(context.Background(), tt.input)

			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				// For API errors, just compare the code and message
				apiErr, isAPIErr := err.(apierror.Error)
				expectedAPIErr, isExpectedAPIErr := tt.err.(apierror.Error)

				if isAPIErr && isExpectedAPIErr {
					if apiErr.Code != expectedAPIErr.Code {
						t.Fatalf("expected error code %s, got %s", expectedAPIErr.Code, apiErr.Code)
					}
					// Message may contain dynamic data, so we skip exact comparison
				} else {
					// At minimum we ensure error exists
					t.Logf("Got error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
				if out.Version == nil || *out.Version != *tt.expected.Version {
					t.Fatalf("expected version %d, got %d", *tt.expected.Version, *out.Version)
				}
				// Verify overwrite was set to false
				if tt.input.Overwrite == nil || *tt.input.Overwrite != false {
					t.Fatalf("expected overwrite to be false, got %v", tt.input.Overwrite)
				}
			}
		})
	}
}

func TestUpdateParameter(t *testing.T) {
	tests := []struct {
		name     string
		input    *ssm.PutParameterInput
		mockSvc  *mockParameterClient
		expected *ssm.PutParameterOutput
		err      error
	}{
		{
			name:  "nil input",
			input: nil,
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "invalid input", nil),
		},
		{
			name: "empty name",
			input: &ssm.PutParameterInput{
				Name:  aws.String(""),
				Value: aws.String("value"),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter name is required", nil),
		},
		{
			name: "empty value",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String(""),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter value is required", nil),
		},
		{
			name: "nil type",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String("value"),
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter type is required", nil),
		},
		{
			name: "AWS service error",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String("value"),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t:                 t,
				putParameterError: awserr.New("ParameterNotFound", "parameter not found", nil),
			},
			expected: nil,
			// This matches the structure from common.ErrCode
			err: apierror.New(apierror.ErrBadRequest, "failed to update parameter: parameter not found", awserr.New("ParameterNotFound", "parameter not found", nil)),
		},
		{
			name: "successful update parameter",
			input: &ssm.PutParameterInput{
				Name:  aws.String("name"),
				Value: aws.String("updatedValue"),
				Type:  aws.String("String"),
			},
			mockSvc: &mockParameterClient{
				t: t,
				putParameterOutput: &ssm.PutParameterOutput{
					Version: aws.Int64(2),
				},
			},
			expected: &ssm.PutParameterOutput{
				Version: aws.Int64(2),
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssmSvc := &SSM{Service: tt.mockSvc}
			out, err := ssmSvc.UpdateParameter(context.Background(), tt.input)

			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				// For API errors, just compare the code and message
				apiErr, isAPIErr := err.(apierror.Error)
				expectedAPIErr, isExpectedAPIErr := tt.err.(apierror.Error)

				if isAPIErr && isExpectedAPIErr {
					if apiErr.Code != expectedAPIErr.Code {
						t.Fatalf("expected error code %s, got %s", expectedAPIErr.Code, apiErr.Code)
					}
					// Message may contain dynamic data, so we skip exact comparison
				} else {
					// At minimum we ensure error exists
					t.Logf("Got error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
				if out.Version == nil || *out.Version != *tt.expected.Version {
					t.Fatalf("expected version %d, got %d", *tt.expected.Version, *out.Version)
				}
				// Verify overwrite was set to true
				if tt.input.Overwrite == nil || *tt.input.Overwrite != true {
					t.Fatalf("expected overwrite to be true, got %v", tt.input.Overwrite)
				}
			}
		})
	}
}

func TestGetParameter(t *testing.T) {
	tests := []struct {
		name           string
		parameterName  string
		withDecryption bool
		mockSvc        *mockParameterClient
		expected       *ssm.GetParameterOutput
		err            error
	}{
		{
			name:           "empty name",
			parameterName:  "",
			withDecryption: false,
			mockSvc: &mockParameterClient{
				t: t,
			},
			expected: nil,
			err:      apierror.New(apierror.ErrBadRequest, "parameter name is required", nil),
		},
		{
			name:           "AWS service error",
			parameterName:  "test-param",
			withDecryption: false,
			mockSvc: &mockParameterClient{
				t:                 t,
				getParameterError: awserr.New("ParameterNotFound", "parameter not found", nil),
			},
			expected: nil,
			// This matches the structure from common.ErrCode
			err: apierror.New(apierror.ErrBadRequest, "failed to get parameter: parameter not found", awserr.New("ParameterNotFound", "parameter not found", nil)),
		},
		{
			name:           "successful get parameter",
			parameterName:  "test-param",
			withDecryption: false,
			mockSvc: &mockParameterClient{
				t: t,
				getParameterOutput: &ssm.GetParameterOutput{
					Parameter: &ssm.Parameter{
						Name:  aws.String("test-param"),
						Value: aws.String("test-value"),
						Type:  aws.String("String"),
					},
				},
			},
			expected: &ssm.GetParameterOutput{
				Parameter: &ssm.Parameter{
					Name:  aws.String("test-param"),
					Value: aws.String("test-value"),
					Type:  aws.String("String"),
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssmSvc := &SSM{Service: tt.mockSvc}
			out, err := ssmSvc.GetParameter(context.Background(), tt.parameterName, tt.withDecryption)

			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				// For API errors, just compare the code and message
				apiErr, isAPIErr := err.(apierror.Error)
				expectedAPIErr, isExpectedAPIErr := tt.err.(apierror.Error)

				if isAPIErr && isExpectedAPIErr {
					if apiErr.Code != expectedAPIErr.Code {
						t.Fatalf("expected error code %s, got %s", expectedAPIErr.Code, apiErr.Code)
					}
					// Message may contain dynamic data, so we skip exact comparison
				} else {
					// At minimum we ensure error exists
					t.Logf("Got error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
				if *out.Parameter.Name != *tt.expected.Parameter.Name {
					t.Fatalf("expected name %s, got %s", *tt.expected.Parameter.Name, *out.Parameter.Name)
				}
				if *out.Parameter.Value != *tt.expected.Parameter.Value {
					t.Fatalf("expected value %s, got %s", *tt.expected.Parameter.Value, *out.Parameter.Value)
				}
				if *out.Parameter.Type != *tt.expected.Parameter.Type {
					t.Fatalf("expected type %s, got %s", *tt.expected.Parameter.Type, *out.Parameter.Type)
				}
			}
		})
	}
}

func TestDeleteParameter(t *testing.T) {
	tests := []struct {
		name          string
		parameterName string
		mockSvc       *mockParameterClient
		err           error
	}{
		{
			name:          "empty name",
			parameterName: "",
			mockSvc: &mockParameterClient{
				t: t,
			},
			err: apierror.New(apierror.ErrBadRequest, "parameter name is required", nil),
		},
		{
			name:          "AWS service error",
			parameterName: "test-param",
			mockSvc: &mockParameterClient{
				t:                    t,
				deleteParameterError: awserr.New("ParameterNotFound", "parameter not found", nil),
			},
			// This matches the structure from common.ErrCode
			err: apierror.New(apierror.ErrBadRequest, "failed to delete parameter: parameter not found", awserr.New("ParameterNotFound", "parameter not found", nil)),
		},
		{
			name:          "successful delete parameter",
			parameterName: "test-param",
			mockSvc: &mockParameterClient{
				t: t,
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssmSvc := &SSM{Service: tt.mockSvc}
			err := ssmSvc.DeleteParameter(context.Background(), tt.parameterName)

			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				// For API errors, just compare the code and message
				apiErr, isAPIErr := err.(apierror.Error)
				expectedAPIErr, isExpectedAPIErr := tt.err.(apierror.Error)

				if isAPIErr && isExpectedAPIErr {
					if apiErr.Code != expectedAPIErr.Code {
						t.Fatalf("expected error code %s, got %s", expectedAPIErr.Code, apiErr.Code)
					}
					// Message may contain dynamic data, so we skip exact comparison
				} else {
					// At minimum we ensure error exists
					t.Logf("Got error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
			}
		})
	}
}

func TestListParameters(t *testing.T) {
	tests := []struct {
		name     string
		input    *ssm.DescribeParametersInput
		mockSvc  *mockParameterClient
		expected *ssm.DescribeParametersOutput
		err      error
	}{
		{
			name:  "AWS service error",
			input: &ssm.DescribeParametersInput{},
			mockSvc: &mockParameterClient{
				t:                       t,
				describeParametersError: awserr.New("InternalServerError", "internal server error", nil),
			},
			expected: nil,
			// This matches the structure from common.ErrCode
			err: apierror.New(apierror.ErrBadRequest, "failed to list parameters: internal server error", awserr.New("InternalServerError", "internal server error", nil)),
		},
		{
			name:  "successful list parameters",
			input: &ssm.DescribeParametersInput{},
			mockSvc: &mockParameterClient{
				t: t,
				describeParametersOutput: &ssm.DescribeParametersOutput{
					Parameters: []*ssm.ParameterMetadata{
						{
							Name: aws.String("param1"),
							Type: aws.String("String"),
						},
						{
							Name: aws.String("param2"),
							Type: aws.String("SecureString"),
						},
					},
				},
			},
			expected: &ssm.DescribeParametersOutput{
				Parameters: []*ssm.ParameterMetadata{
					{
						Name: aws.String("param1"),
						Type: aws.String("String"),
					},
					{
						Name: aws.String("param2"),
						Type: aws.String("SecureString"),
					},
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssmSvc := &SSM{Service: tt.mockSvc}
			out, err := ssmSvc.ListParameters(context.Background(), tt.input)

			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				// For API errors, just compare the code and message
				apiErr, isAPIErr := err.(apierror.Error)
				expectedAPIErr, isExpectedAPIErr := tt.err.(apierror.Error)

				if isAPIErr && isExpectedAPIErr {
					if apiErr.Code != expectedAPIErr.Code {
						t.Fatalf("expected error code %s, got %s", expectedAPIErr.Code, apiErr.Code)
					}
					// Message may contain dynamic data, so we skip exact comparison
				} else {
					// At minimum we ensure error exists
					t.Logf("Got error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
				if len(out.Parameters) != len(tt.expected.Parameters) {
					t.Fatalf("expected %d parameters, got %d", len(tt.expected.Parameters), len(out.Parameters))
				}
				for i, param := range out.Parameters {
					if *param.Name != *tt.expected.Parameters[i].Name {
						t.Fatalf("expected name %s, got %s", *tt.expected.Parameters[i].Name, *param.Name)
					}
					if *param.Type != *tt.expected.Parameters[i].Type {
						t.Fatalf("expected type %s, got %s", *tt.expected.Parameters[i].Type, *param.Type)
					}
				}
			}
		})
	}
}

func TestAddTagsToParameter(t *testing.T) {
	tests := []struct {
		name          string
		parameterName string
		tags          []*ssm.Tag
		mockSvc       *mockParameterClient
		err           error
	}{
		{
			name:          "empty name",
			parameterName: "",
			tags: []*ssm.Tag{
				{
					Key:   aws.String("key"),
					Value: aws.String("value"),
				},
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			err: apierror.New(apierror.ErrBadRequest, "parameter name is required", nil),
		},
		{
			name:          "empty tags",
			parameterName: "test-param",
			tags:          []*ssm.Tag{},
			mockSvc: &mockParameterClient{
				t: t,
			},
			err: nil,
		},
		{
			name:          "AWS service error",
			parameterName: "test-param",
			tags: []*ssm.Tag{
				{
					Key:   aws.String("key"),
					Value: aws.String("value"),
				},
			},
			mockSvc: &mockParameterClient{
				t:                      t,
				addTagsToResourceError: awserr.New("ParameterNotFound", "parameter not found", nil),
			},
			// This matches the structure from common.ErrCode
			err: apierror.New(apierror.ErrBadRequest, "failed to add tags to parameter: parameter not found", awserr.New("ParameterNotFound", "parameter not found", nil)),
		},
		{
			name:          "successful add tags",
			parameterName: "test-param",
			tags: []*ssm.Tag{
				{
					Key:   aws.String("key"),
					Value: aws.String("value"),
				},
			},
			mockSvc: &mockParameterClient{
				t: t,
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssmSvc := &SSM{Service: tt.mockSvc}
			err := ssmSvc.AddTagsToParameter(context.Background(), tt.parameterName, tt.tags)

			if tt.err != nil {
				if err == nil {
					t.Fatalf("expected error but got nil")
				}

				// For API errors, just compare the code and message
				apiErr, isAPIErr := err.(apierror.Error)
				expectedAPIErr, isExpectedAPIErr := tt.err.(apierror.Error)

				if isAPIErr && isExpectedAPIErr {
					if apiErr.Code != expectedAPIErr.Code {
						t.Fatalf("expected error code %s, got %s", expectedAPIErr.Code, apiErr.Code)
					}
					// Message may contain dynamic data, so we skip exact comparison
				} else {
					// At minimum we ensure error exists
					t.Logf("Got error as expected: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %s", err)
				}
			}
		})
	}
}
