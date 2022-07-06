package iam

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
)

func (i *Iam) GetInstanceProfile(ctx context.Context, inp *iam.GetInstanceProfileInput) ([]*Role, error) {
	if inp == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("getting instanceprofiles %s", aws.StringValue(inp.InstanceProfileName))

	out, err := i.Service.GetInstanceProfileWithContext(ctx, inp)
	if err != nil {
		return nil, common.ErrCode("failed to get instanceprofiles", err)
	}

	log.Debugf("got output instanceprofiles: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected get instanceprofiles", nil)
	}

	return out.InstanceProfile.Roles, nil
}
