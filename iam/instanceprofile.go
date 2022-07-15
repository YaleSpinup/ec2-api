package iam

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
)

func (i *Iam) GetInstanceProfile(ctx context.Context, input *iam.GetInstanceProfileInput) (*iam.InstanceProfile, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("getting instanceprofiles %s", aws.StringValue(input.InstanceProfileName))

	out, err := i.Service.GetInstanceProfileWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to get instanceprofiles", err)
	}
	log.Debugf("got output instanceprofiles: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected output in gettting instanceprofiles", nil)
	}

	return out.InstanceProfile, nil
}

func (i *Iam) ListAttachedRolePolicies(ctx context.Context, input *iam.ListAttachedRolePoliciesInput) ([]*iam.AttachedPolicy, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("list attached role policies  %s", aws.StringValue(input.RoleName))

	out, err := i.Service.ListAttachedRolePoliciesWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to list attached role policies", err)
	}
	log.Debugf("got output attached role policies: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected list attached role policies", nil)
	}

	return out.AttachedPolicies, nil
}

func (i *Iam) DetachRolePolicy(ctx context.Context, input *iam.DetachRolePolicyInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("detaching role policy for %s, %s", aws.StringValue(input.RoleName), aws.StringValue(input.PolicyArn))

	_, err := i.Service.DetachRolePolicyWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to detach role policy", err)
	}

	return nil
}

func (i *Iam) ListRolePolicies(ctx context.Context, input *iam.ListRolePoliciesInput) ([]*string, error) {
	if input == nil {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("listing role policies for %s", *input.RoleName)

	out, err := i.Service.ListRolePoliciesWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to list role policies", err)
	}
	log.Debugf("got output list of role policies: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected list of role policies", nil)
	}

	return out.PolicyNames, nil
}

func (i *Iam) DeleteRolePolicy(ctx context.Context, input *iam.DeleteRolePolicyInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("deleting role policy for %s, %s", aws.StringValue(input.RoleName), aws.StringValue(input.PolicyName))

	if _, err := i.Service.DeleteRolePolicyWithContext(ctx, input); err != nil {
		return common.ErrCode("failed to delete role policy", err)
	}

	return nil
}

func (i *Iam) RemoveRoleFromInstanceProfile(ctx context.Context, input *iam.RemoveRoleFromInstanceProfileInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("removing role from instanceprofile %s, %s", aws.StringValue(input.RoleName), aws.StringValue(input.InstanceProfileName))

	if _, err := i.Service.RemoveRoleFromInstanceProfileWithContext(ctx, input); err != nil {
		return common.ErrCode("failed to remove role from instanceprofile", err)
	}

	return nil
}

func (i *Iam) DeleteRole(ctx context.Context, input *iam.DeleteRoleInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("deleting role for %s", aws.StringValue(input.RoleName))

	if _, err := i.Service.DeleteRoleWithContext(ctx, input); err != nil {
		return common.ErrCode("failed to delete role", err)
	}

	return nil
}

func (i *Iam) DeleteInstanceProfile(ctx context.Context, input *iam.DeleteInstanceProfileInput) error {
	if input == nil {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	log.Infof("deleting instanceprofile for %s", aws.StringValue(input.InstanceProfileName))

	if _, err := i.Service.DeleteInstanceProfileWithContext(ctx, input); err != nil {
		return common.ErrCode("failed to delete instanceprofile", err)
	}

	return nil
}
