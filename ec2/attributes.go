package ec2

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) UpdateAttributes(ctx context.Context, instanceType, instanceId string) error {
	if len(instanceId) == 0 || len(instanceType) == 0 {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("updating attributes: %v with instance type %+v", instanceId, instanceType)

	input := ec2.ModifyInstanceAttributeInput{
		InstanceType: &ec2.AttributeValue{Value: aws.String(instanceType)},
		InstanceId:   aws.String(instanceId),
	}

	if _, err := e.Service.ModifyInstanceAttributeWithContext(ctx, &input); err != nil {
		return common.ErrCode("updating attributes", err)
	}

	return nil
}
