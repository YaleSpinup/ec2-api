package iam

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

func (m *mockIAMClient) GetInstanceProfileWithContext(ctx aws.Context, input *iam.GetInstanceProfileInput, opts ...request.Option) (*iam.GetInstanceProfileOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.GetInstanceProfileOutput{InstanceProfile: &iam.InstanceProfile{Roles: []*iam.Role{{RoleName: aws.String("role123")}}}}, nil
}

func (m *mockIAMClient) ListAttachedRolePoliciesWithContext(ctx aws.Context, input *iam.ListAttachedRolePoliciesInput, opts ...request.Option) (*iam.ListAttachedRolePoliciesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.ListAttachedRolePoliciesOutput{AttachedPolicies: []*iam.AttachedPolicy{{PolicyArn: aws.String("arn123")}}}, nil
}

func (m *mockIAMClient) DetachRolePolicyWithContext(ctx aws.Context, input *iam.DetachRolePolicyInput, opts ...request.Option) (*iam.DetachRolePolicyOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DetachRolePolicyOutput{}, nil
}

func (m *mockIAMClient) ListRolePoliciesWithContext(ctx aws.Context, input *iam.ListRolePoliciesInput, opts ...request.Option) (*iam.ListRolePoliciesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.ListRolePoliciesOutput{PolicyNames: aws.StringSlice([]string{"abcd-1"})}, nil
}

func (m *mockIAMClient) DeleteRolePolicyWithContext(ctx aws.Context, input *iam.DeleteRolePolicyInput, opts ...request.Option) (*iam.DeleteRolePolicyOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DeleteRolePolicyOutput{}, nil
}

func (m *mockIAMClient) RemoveRoleFromInstanceProfileWithContext(ctx aws.Context, input *iam.RemoveRoleFromInstanceProfileInput, opts ...request.Option) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.RemoveRoleFromInstanceProfileOutput{}, nil
}

func (m *mockIAMClient) DeleteRoleWithContext(ctx aws.Context, input *iam.DeleteRoleInput, opts ...request.Option) (*iam.DeleteRoleOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DeleteRoleOutput{}, nil
}

func (m *mockIAMClient) DeleteInstanceProfileWithContext(ctx aws.Context, input *iam.DeleteInstanceProfileInput, opts ...request.Option) (*iam.DeleteInstanceProfileOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DeleteInstanceProfileOutput{}, nil
}

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
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &iam.GetInstanceProfileInput{InstanceProfileName: aws.String("profile-456")}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			want:    &iam.InstanceProfile{Roles: []*iam.Role{{RoleName: aws.String("role123")}}},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &iam.GetInstanceProfileInput{InstanceProfileName: aws.String("profile-456")}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			got, err := i.GetInstanceProfile(tt.args.ctx, tt.args.input)
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

func TestIam_ListAttachedRolePolicies(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.ListAttachedRolePoliciesInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		want    []*iam.AttachedPolicy
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &iam.ListAttachedRolePoliciesInput{RoleName: aws.String("role-123")}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			want:    []*iam.AttachedPolicy{{PolicyArn: aws.String("arn123")}},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &iam.ListAttachedRolePoliciesInput{RoleName: aws.String("role-123")}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			got, err := i.ListAttachedRolePolicies(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Iam.ListAttachedRolePolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Iam.ListAttachedRolePolicies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIam_DetachRolePolicy(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.DetachRolePolicyInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(), input: &iam.DetachRolePolicyInput{
				RoleName:  aws.String("role-123"),
				PolicyArn: aws.String("arn-798"),
			}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name: "aws error",
			args: args{ctx: context.TODO(), input: &iam.DetachRolePolicyInput{
				RoleName:  aws.String("role-123"),
				PolicyArn: aws.String("arn-798"),
			}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			if err := i.DetachRolePolicy(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Iam.DetachRolePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIam_ListRolePolicies(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.ListRolePoliciesInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		want    []*string
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &iam.ListRolePoliciesInput{RoleName: aws.String("role-657")}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			want:    aws.StringSlice([]string{"abcd-1"}),
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &iam.ListRolePoliciesInput{RoleName: aws.String("role-657")}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			got, err := i.ListRolePolicies(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Iam.ListRolePolicies() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Iam.ListRolePolicies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIam_DeleteRolePolicy(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.DeleteRolePolicyInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(), input: &iam.DeleteRolePolicyInput{
				RoleName:   aws.String("role-123"),
				PolicyName: aws.String("acbd"),
			}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name: "aws error",
			args: args{ctx: context.TODO(), input: &iam.DeleteRolePolicyInput{
				RoleName:   aws.String("role-123"),
				PolicyName: aws.String("acbd"),
			}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			if err := i.DeleteRolePolicy(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Iam.DeleteRolePolicy() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIam_RemoveRoleFromInstanceProfile(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.RemoveRoleFromInstanceProfileInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(), input: &iam.RemoveRoleFromInstanceProfileInput{
				RoleName:            aws.String("role-123"),
				InstanceProfileName: aws.String("profile-256")}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name: "aws error",
			args: args{ctx: context.TODO(), input: &iam.RemoveRoleFromInstanceProfileInput{
				RoleName:            aws.String("role-123"),
				InstanceProfileName: aws.String("profile-256")}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			if err := i.RemoveRoleFromInstanceProfile(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Iam.RemoveRoleFromInstanceProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIam_DeleteRole(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.DeleteRoleInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &iam.DeleteRoleInput{RoleName: aws.String("role-123")}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &iam.DeleteRoleInput{RoleName: aws.String("role-123")}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			if err := i.DeleteRole(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Iam.DeleteRole() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIam_DeleteInstanceProfile(t *testing.T) {
	type fields struct {
		Service iamiface.IAMAPI
	}
	type args struct {
		ctx   context.Context
		input *iam.DeleteInstanceProfileInput
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String("profilename-123")}},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &iam.DeleteInstanceProfileInput{InstanceProfileName: aws.String("profilename-123")}},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			if err := i.DeleteInstanceProfile(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Iam.DeleteInstanceProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
