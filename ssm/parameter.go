package ssm

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

// CreateParameter creates a new SSM parameter
func (s *SSM) CreateParameter(ctx context.Context, input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	if input.Name == nil || aws.StringValue(input.Name) == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	if input.Value == nil || aws.StringValue(input.Value) == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter value is required", nil)
	}

	if input.Type == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter type is required", nil)
	}

	// Ensure overwrite is set to false for creation
	input.Overwrite = aws.Bool(false)

	log.Infof("creating parameter: %s", aws.StringValue(input.Name))

	out, err := s.Service.PutParameterWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to create parameter", err)
	}

	log.Debugf("create parameter output: %+v", out)

	return out, nil
}

// UpdateParameter updates an existing SSM parameter
func (s *SSM) UpdateParameter(ctx context.Context, input *ssm.PutParameterInput) (*ssm.PutParameterOutput, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	if input.Name == nil || aws.StringValue(input.Name) == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	if input.Value == nil || aws.StringValue(input.Value) == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter value is required", nil)
	}

	if input.Type == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter type is required", nil)
	}

	// Force overwrite to true for update
	input.Overwrite = aws.Bool(true)

	log.Infof("updating parameter: %s", aws.StringValue(input.Name))

	out, err := s.Service.PutParameterWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to update parameter", err)
	}

	log.Debugf("update parameter output: %+v", out)

	return out, nil
}

// GetParameter retrieves a parameter from SSM parameter store
func (s *SSM) GetParameter(ctx context.Context, name string, withDecryption bool) (*ssm.GetParameterOutput, error) {
	if name == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	log.Infof("getting parameter: %s", name)

	out, err := s.Service.GetParameterWithContext(ctx, &ssm.GetParameterInput{
		Name:           aws.String(name),
		WithDecryption: aws.Bool(withDecryption),
	})
	if err != nil {
		return nil, common.ErrCode("failed to get parameter", err)
	}

	return out, nil
}

// DeleteParameter deletes a parameter from SSM parameter store
func (s *SSM) DeleteParameter(ctx context.Context, name string) error {
	if name == "" {
		return apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	log.Infof("deleting parameter: %s", name)

	_, err := s.Service.DeleteParameterWithContext(ctx, &ssm.DeleteParameterInput{
		Name: aws.String(name),
	})
	if err != nil {
		return common.ErrCode("failed to delete parameter", err)
	}

	return nil
}

// ListParameters lists parameters with optional filters
func (s *SSM) ListParameters(ctx context.Context, input *ssm.DescribeParametersInput) (*ssm.DescribeParametersOutput, error) {
	log.Infof("listing parameters with input: %+v", input)

	if input == nil {
		input = &ssm.DescribeParametersInput{}
	}

	output, err := s.Service.DescribeParametersWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to list parameters", err)
	}

	return output, nil
}

// AddTagsToParameter adds tags to a parameter
func (s *SSM) AddTagsToParameter(ctx context.Context, name string, tags []*ssm.Tag) error {
	if name == "" {
		return apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	if len(tags) == 0 {
		return nil
	}

	log.Infof("adding tags to parameter: %s", name)

	input := &ssm.AddTagsToResourceInput{
		ResourceType: aws.String("Parameter"),
		ResourceId:   aws.String(name),
		Tags:         tags,
	}

	_, err := s.Service.AddTagsToResourceWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to add tags to parameter", err)
	}

	return nil
}
