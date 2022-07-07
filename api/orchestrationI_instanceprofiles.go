package api

import (
	"context"

	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
)

func (o *iamOrchestrator) getInstanceProfile(ctx context.Context, name string) error {
	ip, err := o.iamClient.GetInstanceProfile(ctx, name)
	if err != nil {
		return common.ErrCode("failed to get instance profiles", err)
	}
	for _, managedRole := range ip.Roles {
		roleName := aws.StringValue(managedRole.RoleName)
		rps, err := o.iamClient.ListAttachedRolePolicies(ctx, roleName)
		if err != nil {
			return common.ErrCode("failed to get managed roles", err)
		}
		for _, rp := range rps {
			if err := o.iamClient.DetachRolePolicy(ctx, roleName, aws.StringValue(rp.PolicyArn)); err != nil {
				return common.ErrCode("failed to detach managed roles", err)
			}
		}
		pns, err := o.iamClient.ListRolePolicies(ctx, roleName)
		if err != nil {
			return common.ErrCode("failed to get managed roles", err)
		}
		for _, p := range pns {
			if err := o.iamClient.DeleteRolePolicy(ctx, roleName, aws.StringValue(p)); err != nil {
				return common.ErrCode("failed to detach managed roles", err)
			}
		}
		if err := o.iamClient.RemoveRoleFromInstanceProfile(ctx, roleName, aws.StringValue(ip.InstanceProfileName)); err != nil {
			return common.ErrCode("failed to detach managed roles", err)
		}
		if err := o.iamClient.DeleteRole(ctx, roleName); err != nil {
			return common.ErrCode("failed to detach managed roles", err)
		}
	}
	if err := o.iamClient.DeleteInstanceProfile(ctx, aws.StringValue(ip.InstanceProfileName)); err != nil {
		return common.ErrCode("failed to detach managed roles", err)
	}
	return nil
}
