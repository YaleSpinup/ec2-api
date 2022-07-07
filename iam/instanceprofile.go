package iam

import (
	"context"

	"github.com/YaleSpinup/apierror"
	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	log "github.com/sirupsen/logrus"
)

func (i *Iam) GetInstanceProfile(ctx context.Context, name string) (*iam.InstanceProfile, error) {
	if name == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	inp := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(name),
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

	return out.InstanceProfile, nil
}

func (i *Iam) ListAttachedRolePolicies(ctx context.Context, roleName string) ([]*iam.AttachedPolicy, error) {
	if roleName == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	log.Infof("getting managed roles for %s", roleName)

	out, err := i.Service.ListAttachedRolePoliciesWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to list attached policies", err)
	}

	log.Debugf("got output instanceprofiles: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected list attached policies", nil)
	}

	return out.AttachedPolicies, nil
}

func (i *Iam) DetachRolePolicy(ctx context.Context, roleName, policyArn string) error {
	if roleName == "" || policyArn == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.DetachRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyArn: aws.String(policyArn),
	}
	log.Infof("detaching managed roles for %s, %s", roleName, policyArn)

	_, err := i.Service.DetachRolePolicyWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to detach managed policies", err)
	}

	return nil
}

func (i *Iam) ListRolePolicies(ctx context.Context, roleName string) ([]*string, error) {
	if roleName == "" {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.ListRolePoliciesInput{
		RoleName: aws.String(roleName),
	}
	log.Infof("getting managed roles for %s", roleName)

	out, err := i.Service.ListRolePoliciesWithContext(ctx, input)
	if err != nil {
		return nil, common.ErrCode("failed to list attached policies", err)
	}

	log.Debugf("got output instanceprofiles: %+v", out)

	if out == nil {
		return nil, apierror.New(apierror.ErrInternalError, "Unexpected list attached policies", nil)
	}

	return out.PolicyNames, nil
}

func (i *Iam) DeleteRolePolicy(ctx context.Context, roleName, policyName string) error {
	if roleName == "" || policyName == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.DeleteRolePolicyInput{
		RoleName:  aws.String(roleName),
		PolicyName: aws.String(policyName),
	}
	log.Infof("deleting attached roles for %s, %s", roleName, policyName)

	_, err := i.Service.DeleteRolePolicyWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to detach managed policies", err)
	}

	return nil
}

func (i *Iam) RemoveRoleFromInstanceProfile(ctx context.Context, roleName, instanceProfileName string) error {
	if roleName == "" || instanceProfileName  == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.RemoveRoleFromInstanceProfileInput{
		RoleName:  aws.String(roleName),
		InstanceProfileName: aws.String(instanceProfileName),
	}
	log.Infof("deleting attached roles for %s, %s", roleName, instanceProfileName)

	_, err := i.Service.RemoveRoleFromInstanceProfileWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to detach managed policies", err)
	}

	return nil
}

func (i *Iam) DeleteRole(ctx context.Context, roleName string) error {
	if roleName == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.DeleteRoleInput{
		RoleName:  aws.String(roleName),
	}
	log.Infof("deleting attached roles for %s", roleName)

	_, err := i.Service.DeleteRoleWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to detach managed policies", err)
	}

	return nil
}

func (i *Iam) DeleteInstanceProfile(ctx context.Context, instanceProfileName string) error {
	if instanceProfileName == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}
	input := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	}
	log.Infof("deleting attached roles for %s", instanceProfileName)

	_, err := i.Service.DeleteInstanceProfileWithContext(ctx, input)
	if err != nil {
		return common.ErrCode("failed to detach managed policies", err)
	}

	return nil
}