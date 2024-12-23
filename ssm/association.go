package ssm

import (
	"context"

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

func (s *SSM) CreateAssociationByTag(ctx context.Context, tagKey string, tagValues []string, docName string) (string, error) {
	// Check for missing values
	if tagKey == "" || tagValues == nil {
		return "", apierror.New(apierror.ErrBadRequest, "both tagKey and tagValues should be present", nil)
	}
	if docName == "" {
		return "", apierror.New(apierror.ErrBadRequest, "docName should be present", nil)
	}

	// Create the Targets structure
	targets := []*ssm.Target{
		{
			Key:    aws.String("tag:" + tagKey),
			Values: aws.StringSlice(tagValues),
		},
	}

	inp := &ssm.CreateAssociationInput{
		Name:    aws.String(docName),
		Targets: targets,
	}

	out, err := s.Service.CreateAssociationWithContext(ctx, inp)
	if err != nil {
		return "", common.ErrCode("failed to create association", err)
	}
	log.Debugf("got output creating SSM Association: %+v", out)
	return aws.StringValue(out.AssociationDescription.AssociationId), nil
}
