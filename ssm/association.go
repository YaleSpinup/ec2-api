package ssm

import (
	"context"
	"strconv"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	log "github.com/sirupsen/logrus"
)

func (s *SSM) DescribeAssociation(ctx context.Context, instanceId, docName string) (*ssm.DescribeAssociationOutput, error) {
	if instanceId == "" || docName == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "both instanceId and docName should be present", nil)
	}
	out, err := s.Service.DescribeAssociationWithContext(ctx, &ssm.DescribeAssociationInput{
		Name:       aws.String(docName),
		InstanceId: aws.String(instanceId),
	})
	if err != nil {
		return nil, common.ErrCode("failed to describe association", err)
	}
	log.Debugf("got output describing SSM Association: %+v", out)
	return out, nil
}

func (s *SSM) CreateAssociation(ctx context.Context, instanceId, docName string) (string, error) {
	if instanceId == "" || docName == "" {
		return "", apierror.New(apierror.ErrBadRequest, "both instanceId and docName should be present", nil)
	}
	inp := &ssm.CreateAssociationInput{
		Name:       aws.String(docName),
		InstanceId: aws.String(instanceId),
	}
	out, err := s.Service.CreateAssociationWithContext(ctx, inp)
	if err != nil {
		return "", common.ErrCode("failed to create association", err)
	}
	log.Debugf("got output creating SSM Association: %+v", out)
	return aws.StringValue(out.AssociationDescription.AssociationId), nil
}

// CreateAssociationByTag create a ssm association by tag targets
// 1. Create the targets structure
// 2. Create required association input
// 3. Handle optional association name
// 4. Handle optional document version
// 5. Handle optional schedule
// 6. Handle optional parameters
// 7. Create the ssm association with the created input
// 8. Return the association id of the created association
func (s *SSM) CreateAssociationByTag(ctx context.Context, associationName string, docName string, docVersion int, scheduleExpression string, scheduleOffset int, tagFilters map[string][]string, parameters map[string][]string) (string, error) {
	// Create the Targets structure
	var targets []*ssm.Target
	for tagKey, tagValues := range tagFilters {
		if tagKey == "" {
			return "", apierror.New(apierror.ErrBadRequest, "tag key cannot be empty", nil)
		}
		targets = append(targets,
			&ssm.Target{
				Key:    aws.String("tag:" + tagKey),
				Values: aws.StringSlice(tagValues),
			},
		)
	}

	// Create the input
	input := &ssm.CreateAssociationInput{
		Name:    aws.String(docName),
		Targets: targets,
	}

	// Handle optional association name
	if associationName != "" {
		input.AssociationName = aws.String(associationName)
	}

	// Handle optional document version
	if docVersion != 0 {
		input.DocumentVersion = aws.String(strconv.Itoa(docVersion))
	} else {
		input.DocumentVersion = aws.String("$DEFAULT")
	}

	// Handle optional schedule
	if scheduleExpression != "" {
		input.ScheduleExpression = aws.String(scheduleExpression)
		if scheduleOffset != 0 {
			input.ScheduleOffset = aws.Int64(int64(scheduleOffset))
		}
	}

	// Handle optional parameters
	if len(parameters) > 0 {
		// Convert the parameters from map[string][]string to map[string][]*string
		awsParams := make(map[string][]*string)
		for key, values := range parameters {
			// Convert each string in the slice to *string
			var awsValues []*string
			for _, val := range values {
				awsValues = append(awsValues, aws.String(val))
			}
			awsParams[key] = awsValues
		}
		input.Parameters = awsParams
	}

	// create the ssm association in aws
	out, err := s.Service.CreateAssociationWithContext(ctx, input)
	if err != nil {
		return "", common.ErrCode("failed to create association", err)
	}

	// return the association id
	log.Debugf("got output creating SSM Association: %+v", out)
	log.Info("created association with id: ", aws.StringValue(out.AssociationDescription.AssociationId))
	return aws.StringValue(out.AssociationDescription.AssociationId), nil
}
