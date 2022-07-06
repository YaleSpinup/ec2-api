package api

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

func (o *iamOrchestrator) getInstanceProfile(ctx context.Context, name string) (string, error) {
	if name == "" {
		return "", apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	inp := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
	}
	out, err := o.iamClient.GetInstanceProfile(ctx, inp)
	if err != nil {
		return "", common.ErrCode("failed to update instance type attributes", err)
	}
	for _, s := range out {



	}

	return aws.StringValue(out.InstanceProfileId), nil
}
