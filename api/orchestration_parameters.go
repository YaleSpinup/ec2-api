package api

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

// createParameter creates a new SSM parameter
func (o *ssmOrchestrator) createParameter(ctx context.Context, req *SSMParameterCreateRequest) (*SSMParameterResponse, error) {
	if req == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	// AWS SSM parameters should start with a forward slash
	// Add a leading slash if not present
	paramName := aws.StringValue(req.Name)
	if len(paramName) > 0 && paramName[0] != '/' {
		paramName = "/" + paramName
		req.Name = aws.String(paramName)
	}

	log.Infof("creating SSM parameter %s", aws.StringValue(req.Name))

	// Create the SSM PutParameter input from the request
	input := &ssm.PutParameterInput{
		Name:  req.Name,
		Type:  req.Type,
		Value: req.Value,
	}

	if req.Description != nil {
		input.Description = req.Description
	}

	if req.KeyId != nil && aws.StringValue(req.Type) == "SecureString" {
		input.KeyId = req.KeyId
	}

	if req.Tier != nil {
		input.Tier = req.Tier
	}

	// Create the parameter (this will explicitly set Overwrite=false)
	output, err := o.ssmClient.CreateParameter(ctx, input)
	if err != nil {
		return nil, err
	}

	// Add tags if specified
	if req.Tags != nil && len(req.Tags) > 0 {
		// Convert tags to AWS format
		ssmTags := []*ssm.Tag{}
		for k, v := range req.Tags {
			ssmTags = append(ssmTags, &ssm.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}

		if err := o.ssmClient.AddTagsToParameter(ctx, aws.StringValue(req.Name), ssmTags); err != nil {
			log.Errorf("failed to tag parameter %s: %s", aws.StringValue(req.Name), err)
			// Continue even if tagging fails
		}
	}

	// Get the parameter details for the response
	getOutput, err := o.ssmClient.GetParameter(ctx, aws.StringValue(req.Name), false)
	if err != nil {
		log.Errorf("parameter created but couldn't retrieve its details: %s", err)

		// Return limited response with version from PutParameter output
		return &SSMParameterResponse{
			Name:    aws.StringValue(req.Name),
			Type:    aws.StringValue(req.Type),
			Version: aws.Int64Value(output.Version),
		}, nil
	}

	// Get parameter metadata including tags
	describeInput := &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Values: []*string{req.Name},
			},
		},
	}

	metadata, err := o.ssmClient.ListParameters(ctx, describeInput)
	if err != nil {
		log.Errorf("couldn't retrieve parameter metadata: %s", err)
	}

	// Build response with combined data
	response := &SSMParameterResponse{
		Name:     aws.StringValue(getOutput.Parameter.Name),
		Type:     aws.StringValue(getOutput.Parameter.Type),
		Version:  aws.Int64Value(getOutput.Parameter.Version),
		ARN:      aws.StringValue(getOutput.Parameter.ARN),
		DataType: aws.StringValue(getOutput.Parameter.DataType),
	}

	// Add metadata if available
	if metadata != nil && len(metadata.Parameters) > 0 {
		paramDetail := metadata.Parameters[0]

		if paramDetail.LastModifiedDate != nil {
			response.LastModified = timeFormat(paramDetail.LastModifiedDate)
		}

		if paramDetail.Description != nil {
			response.Description = aws.StringValue(paramDetail.Description)
		}

		if paramDetail.Tier != nil {
			response.Tier = aws.StringValue(paramDetail.Tier)
		}

		// Get tags for the parameter
		if req.Tags != nil && len(req.Tags) > 0 {
			tagList := []map[string]string{}
			for k, v := range req.Tags {
				tagList = append(tagList, map[string]string{k: v})
			}
			response.Tags = tagList
		}
	}

	return response, nil
}

// updateParameter updates an existing SSM parameter
func (o *ssmOrchestrator) updateParameter(ctx context.Context, req *SSMParameterCreateRequest) (*SSMParameterResponse, error) {
	if req == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	// AWS SSM parameters should start with a forward slash
	// Add a leading slash if not present
	paramName := aws.StringValue(req.Name)
	if len(paramName) > 0 && paramName[0] != '/' {
		paramName = "/" + paramName
		req.Name = aws.String(paramName)
	}

	log.Infof("updating SSM parameter %s", aws.StringValue(req.Name))

	// Create the SSM PutParameter input from the request
	input := &ssm.PutParameterInput{
		Name:  req.Name,
		Type:  req.Type,
		Value: req.Value,
	}

	if req.Description != nil {
		input.Description = req.Description
	}

	if req.KeyId != nil && aws.StringValue(req.Type) == "SecureString" {
		input.KeyId = req.KeyId
	}

	if req.Tier != nil {
		input.Tier = req.Tier
	}

	// Update the parameter (this will explicitly set Overwrite=true)
	output, err := o.ssmClient.UpdateParameter(ctx, input)
	if err != nil {
		return nil, err
	}

	// Add tags if specified
	if req.Tags != nil && len(req.Tags) > 0 {
		// Convert tags to AWS format
		ssmTags := []*ssm.Tag{}
		for k, v := range req.Tags {
			ssmTags = append(ssmTags, &ssm.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}

		if err := o.ssmClient.AddTagsToParameter(ctx, aws.StringValue(req.Name), ssmTags); err != nil {
			log.Errorf("failed to tag parameter %s: %s", aws.StringValue(req.Name), err)
			// Continue even if tagging fails
		}
	}

	// Get the parameter details for the response
	getOutput, err := o.ssmClient.GetParameter(ctx, aws.StringValue(req.Name), false)
	if err != nil {
		log.Errorf("parameter updated but couldn't retrieve its details: %s", err)

		// Return limited response with version from PutParameter output
		return &SSMParameterResponse{
			Name:    aws.StringValue(req.Name),
			Type:    aws.StringValue(req.Type),
			Version: aws.Int64Value(output.Version),
		}, nil
	}

	// Get parameter metadata including tags
	describeInput := &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Values: []*string{req.Name},
			},
		},
	}

	metadata, err := o.ssmClient.ListParameters(ctx, describeInput)
	if err != nil {
		log.Errorf("couldn't retrieve parameter metadata: %s", err)
	}

	// Build response with combined data
	response := &SSMParameterResponse{
		Name:     aws.StringValue(getOutput.Parameter.Name),
		Type:     aws.StringValue(getOutput.Parameter.Type),
		Version:  aws.Int64Value(getOutput.Parameter.Version),
		ARN:      aws.StringValue(getOutput.Parameter.ARN),
		DataType: aws.StringValue(getOutput.Parameter.DataType),
	}

	// Add metadata if available
	if metadata != nil && len(metadata.Parameters) > 0 {
		paramDetail := metadata.Parameters[0]

		if paramDetail.LastModifiedDate != nil {
			response.LastModified = timeFormat(paramDetail.LastModifiedDate)
		}

		if paramDetail.Description != nil {
			response.Description = aws.StringValue(paramDetail.Description)
		}

		if paramDetail.Tier != nil {
			response.Tier = aws.StringValue(paramDetail.Tier)
		}

		// Get tags for the parameter
		if req.Tags != nil && len(req.Tags) > 0 {
			tagList := []map[string]string{}
			for k, v := range req.Tags {
				tagList = append(tagList, map[string]string{k: v})
			}
			response.Tags = tagList
		}
	}

	return response, nil
}

// getParameter retrieves an SSM parameter
func (o *ssmOrchestrator) getParameter(ctx context.Context, name string, withDecryption bool) (*SSMParameterResponse, error) {
	if name == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	// AWS SSM parameters should start with a forward slash
	// Add a leading slash if not present
	paramName := name
	if len(paramName) > 0 && paramName[0] != '/' {
		paramName = "/" + paramName
	}

	log.Infof("getting SSM parameter %s", paramName)

	// Get the parameter
	getOutput, err := o.ssmClient.GetParameter(ctx, paramName, withDecryption)
	if err != nil {
		return nil, err
	}

	// Get parameter metadata including tags
	describeInput := &ssm.DescribeParametersInput{
		ParameterFilters: []*ssm.ParameterStringFilter{
			{
				Key:    aws.String("Name"),
				Values: []*string{aws.String(paramName)},
			},
		},
	}

	metadata, err := o.ssmClient.ListParameters(ctx, describeInput)
	if err != nil {
		log.Errorf("couldn't retrieve parameter metadata: %s", err)
	}

	// Build response
	response := &SSMParameterResponse{
		Name:     aws.StringValue(getOutput.Parameter.Name),
		Type:     aws.StringValue(getOutput.Parameter.Type),
		Version:  aws.Int64Value(getOutput.Parameter.Version),
		ARN:      aws.StringValue(getOutput.Parameter.ARN),
		DataType: aws.StringValue(getOutput.Parameter.DataType),
	}

	// Add metadata if available
	if metadata != nil && len(metadata.Parameters) > 0 {
		paramDetail := metadata.Parameters[0]

		if paramDetail.LastModifiedDate != nil {
			response.LastModified = timeFormat(paramDetail.LastModifiedDate)
		}

		if paramDetail.Description != nil {
			response.Description = aws.StringValue(paramDetail.Description)
		}

		if paramDetail.Tier != nil {
			response.Tier = aws.StringValue(paramDetail.Tier)
		}

		// Try to get tags
		tagOutput, tagErr := o.ssmClient.Service.ListTagsForResourceWithContext(ctx, &ssm.ListTagsForResourceInput{
			ResourceType: aws.String("Parameter"),
			ResourceId:   aws.String(name),
		})

		if tagErr == nil && tagOutput.TagList != nil && len(tagOutput.TagList) > 0 {
			tagList := []map[string]string{}
			for _, tag := range tagOutput.TagList {
				tagList = append(tagList, map[string]string{
					aws.StringValue(tag.Key): aws.StringValue(tag.Value),
				})
			}
			response.Tags = tagList
		}
	}

	return response, nil
}

// deleteParameter deletes an SSM parameter
func (o *ssmOrchestrator) deleteParameter(ctx context.Context, name string) error {
	if name == "" {
		return apierror.New(apierror.ErrBadRequest, "parameter name is required", nil)
	}

	// AWS SSM parameters should start with a forward slash
	// Add a leading slash if not present
	paramName := name
	if len(paramName) > 0 && paramName[0] != '/' {
		paramName = "/" + paramName
	}

	log.Infof("deleting SSM parameter %s", paramName)

	return o.ssmClient.DeleteParameter(ctx, paramName)
}

// listParameters lists SSM parameters with optional filters
func (o *ssmOrchestrator) listParameters(ctx context.Context, filters map[string]string, maxResults int64, nextToken string) ([]*SSMParameterResponse, string, error) {
	log.Infof("listing SSM parameters with filters: %+v", filters)

	input := &ssm.DescribeParametersInput{}

	if maxResults > 0 {
		input.MaxResults = aws.Int64(maxResults)
	}

	if nextToken != "" {
		input.NextToken = aws.String(nextToken)
	}

	// Set up parameter filters if provided
	if len(filters) > 0 {
		paramFilters := []*ssm.ParameterStringFilter{}

		for key, value := range filters {
			if value != "" {
				paramFilters = append(paramFilters, &ssm.ParameterStringFilter{
					Key: aws.String(key),
					Values: []*string{
						aws.String(value),
					},
				})
			}
		}

		if len(paramFilters) > 0 {
			input.ParameterFilters = paramFilters
		}
	}

	// Get parameters
	output, err := o.ssmClient.ListParameters(ctx, input)
	if err != nil {
		return nil, "", err
	}

	// Map response
	response := []*SSMParameterResponse{}
	for _, param := range output.Parameters {
		paramResponse := &SSMParameterResponse{
			Name: aws.StringValue(param.Name),
			Type: aws.StringValue(param.Type),
		}

		if param.LastModifiedDate != nil {
			paramResponse.LastModified = timeFormat(param.LastModifiedDate)
		}

		if param.Description != nil {
			paramResponse.Description = aws.StringValue(param.Description)
		}

		if param.Tier != nil {
			paramResponse.Tier = aws.StringValue(param.Tier)
		}

		response = append(response, paramResponse)
	}

	var nextTokenResponse string
	if output.NextToken != nil {
		nextTokenResponse = aws.StringValue(output.NextToken)
	}

	return response, nextTokenResponse, nil
}
