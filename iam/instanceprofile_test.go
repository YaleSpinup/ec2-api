package iam

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

// func (m *mockIAMClient) GetInstanceProfileWithContext(ctx context.Context, input *iam.GetInstanceProfileInput, opt ...request.Option) (*iam.GetInstanceProfileOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.GetInstanceProfileOutput{InstanceProfile: &iam.InstanceProfile{Roles: []*iam.Role{{RoleName: aws.String("role123")}}}}, nil
// }

// func (m *mockIAMClient) ListAttachedRolePoliciesWithContext(ctx context.Context, input *iam.ListAttachedRolePoliciesInput, opt ...request.Option) (*iam.ListAttachedRolePoliciesOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{{PolicyArn: aws.String("arn123")}}}, nil
// }

// func (m *mockIAMClient) DetachRolePolicyWithContext(ctx context.Context, input *iam.DetachRolePolicyInput, opt ...request.Option) (*iam.DetachRolePolicyOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.DetachRolePolicyOutput{}, nil
// }

// func (m *mockIAMClient) ListRolePoliciesWithContext(ctx context.Context, input *iam.ListRolePoliciesInput, opt ...request.Option) (*iam.ListRolePoliciesOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.ListRolePoliciesOutput{PolicyNames: aws.StringSlice([]string{"abcd-1"})}, nil
// }

// func (m *mockIAMClient) DeleteRolePolicyWithContext(ctx context.Context, input *iam.DeleteRolePolicyInput, opt ...request.Option) (*iam.DeleteRolePolicyOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.DeleteRolePolicyOutput{}, nil
// }

// func (m *mockIAMClient) RemoveRoleFromInstanceProfileWithContext(ctx context.Context, input *iam.RemoveRoleFromInstanceProfileInput, opt ...request.Option) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.RemoveRoleFromInstanceProfileOutput{}, nil
// }

// func (m *mockIAMClient) DeleteRoleWithContext(ctx context.Context, input *iam.DeleteRoleInput, opt ...request.Option) (*iam.DeleteRoleOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.DeleteRoleOutput{}, nil
// }

// func (m *mockIAMClient) DeleteInstanceProfileWithContext(ctx context.Context, input *iam.DeleteInstanceProfileInput, opt ...request.Option) (*iam.DeleteInstanceProfileOutput, error) {
// 	if m.err != nil {
// 		return nil, m.err
// 	}
// 	return &iam.DeleteInstanceProfileOutput{}, nil
// }

func TestIam_GetInstanceProfile(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.GetInstanceProfileInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		want    *iam.InstanceProfile
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.i.GetInstanceProfile(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Iam.GetInstanceProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Iam.GetInstanceProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}
