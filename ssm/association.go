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
