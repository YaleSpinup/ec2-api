package api

import (
	"context"
	"fmt"

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
	for _, r := range ip.Roles {
		// detach all attached policies
		rps, err := o.iamClient.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{RoleName: r.RoleName})
		if err != nil {
			return common.ErrCode(fmt.Sprintf("failed to list attached policies for the role %s", aws.StringValue(r.RoleName)), err)
		}
		for _, rp := range rps {
			input := &iam.DetachRolePolicyInput{
				RoleName:  r.RoleName,
				PolicyArn: rp.PolicyArn,
			}
			if err := o.iamClient.DetachRolePolicy(ctx, input); err != nil {
				return common.ErrCode(fmt.Sprintf("failed to detach policy for the role %s", aws.StringValue(r.RoleName)), err)
			}
		}

		// delete all inline policies
		ps, err := o.iamClient.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{RoleName: r.RoleName})
		if err != nil {
			return common.ErrCode("failed to list inline policies", err)
		}
		for _, p := range ps {
			input := &iam.DeleteRolePolicyInput{
				RoleName:   r.RoleName,
				PolicyName: p,
			}
			if err := o.iamClient.DeleteRolePolicy(ctx, input); err != nil {
				return common.ErrCode("failed to delete inline policy", err)
			}
		}

		// remove the role from the instance profile
		input := &iam.RemoveRoleFromInstanceProfileInput{
			RoleName:            r.RoleName,
			InstanceProfileName: ip.InstanceProfileName,
		}
		if err := o.iamClient.RemoveRoleFromInstanceProfile(ctx, input); err != nil {
			return common.ErrCode("failed to remove role from instance profile", err)
		}

		// delete the role
		if err := o.iamClient.DeleteRole(ctx, &iam.DeleteRoleInput{RoleName: r.RoleName}); err != nil {
			return common.ErrCode("failed to delete role", err)
		}
	}

	// delete the instance profile
	if err := o.iamClient.DeleteInstanceProfile(ctx, &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String(aws.StringValue(ip.InstanceProfileName))}); err != nil {
		return common.ErrCode("failed to delete instance profile", err)
	}
	return nil
}
