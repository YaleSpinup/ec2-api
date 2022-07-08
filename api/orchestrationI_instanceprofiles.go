package api

import (
	"context"

	"github.com/YaleSpinup/ec2-api/common"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
)

// deleteInstanceProfile deletes the specified instance profile and associated role, if they exist
// any policies attached to the role will be detached and left intact
func (o *iamOrchestrator) deleteInstanceProfile(ctx context.Context, name string) error {
	ip, err := o.iamClient.GetInstanceProfile(ctx, &iam.GetInstanceProfileInput{InstanceProfileName: aws.String(name)})
	if err != nil {
		return common.ErrCode("failed to get instance profiles", err)
	}
	// detach policies from role(s) and delete the role(s)
	for _, managedRole := range ip.Roles {
		roleName := aws.StringValue(managedRole.RoleName)
		// detach all managed policies
		rps, err := o.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: aws.String(roleName)})
		if err != nil {
			return common.ErrCode("failed to list managed roles", err)
		}
		for _, rp := range rps {
			input := &iam.DetachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: rp.PolicyArn,
			}
			if err := o.iamClient.DetachRolePolicy(ctx, input); err != nil {
				return common.ErrCode("failed to detach managed roles", err)
			}
		}
		// delete all inline policies
		pns, err := o.iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{RoleName: aws.String(roleName)})
		if err != nil {
			return common.ErrCode("failed to list role policies", err)
		}
		for _, p := range pns {
			input := &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: p,
			}
			if err := o.iamClient.DeleteRolePolicy(ctx, input); err != nil {
				return common.ErrCode("failed to delete role policy", err)
			}
		}
		// remove the role from the instance profile
		input := &iam.RemoveRoleFromInstanceProfileInput{
			RoleName:            aws.String(roleName),
			InstanceProfileName: ip.InstanceProfileName,
		}
		if err := o.iamClient.RemoveRoleFromInstanceProfile(ctx, input); err != nil {
			return common.ErrCode("failed to remove role from instance profile", err)
		}
		// delete the role
		if err := o.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: aws.String(roleName)}); err != nil {
			return common.ErrCode("failed to delete role", err)
		}
	}
	// delete the instance profile
	if err := o.iamClient.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String(aws.StringValue(ip.InstanceProfileName))}); err != nil {
		return common.ErrCode("failed to delete instance profile", err)
	}
	return nil
}
