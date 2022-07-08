package iam

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
)

func (m *mockIAMClient) GetInstanceProfileWithContext(ctx context.Context, name string) (*iam.GetInstanceProfileOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.GetInstanceProfileOutput{}, nil
}

func (m *mockIAMClient) ListAttachedRolePoliciesWithContext(ctx context.Context, roleName string) (*iam.ListAttachedRolePoliciesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.ListAttachedRolePoliciesOutput{}, nil
}

func (m *mockIAMClient) DetachRolePolicyWithContext(ctx context.Context, roleName, policyArn string) (*iam.DetachRolePolicyOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DetachRolePolicyOutput{}, nil
}

func (m *mockIAMClient) ListRolePoliciesWithContext(ctx context.Context, roleName string) (*iam.ListRolePoliciesOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.ListRolePoliciesOutput{}, nil
}

func (m *mockIAMClient) DeleteRolePolicyWithContext(ctx context.Context, roleName, policyName string) (*iam.DeleteRolePolicyOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DeleteRolePolicyOutput{}, nil
}

func (m *mockIAMClient) RemoveRoleFromInstanceProfileWithContext(ctx context.Context, roleName, instanceProfileName string) (*iam.RemoveRoleFromInstanceProfileOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.RemoveRoleFromInstanceProfileOutput{}, nil
}

func (m *mockIAMClient) DeleteRoleWithContext(ctx context.Context, rolenName string) (*iam.DeleteRoleOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &iam.DeleteRoleOutput{}, nil
}

func (m *mockIAMClient) DeleteInstanceProfileWithContext(ctx context.Context, input *iam.DeleteInstanceProfileInput, opt ...request.Option) (*iam.DeleteInstanceProfileOutput, error) {
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
		ctx  context.Context
		name string
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
			i := &Iam{
				Service: tt.fields.Service,
			}
			got, err := i.GetInstanceProfile(tt.args.ctx, tt.args.name)
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
		ctx      context.Context
		roleName string
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		want    []*iam.AttachedPolicy
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			got, err := i.ListAttachedRolePolicies(tt.args.ctx, tt.args.roleName)
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
		ctx       context.Context
		roleName  string
		policyArn string
	}
	tests := []struct {
		name    string
		i       *Iam
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), roleName: "abcd-12", policyArn: "e2fer34"},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), roleName: "abcd-12", policyArn: "e2fer34"},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			err := i.DetachRolePolicy(tt.args.ctx, tt.args.roleName, tt.args.policyArn)
			if (err != nil) != tt.wantErr {
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
		ctx      context.Context
		roleName string
	}
	tests := []struct {
		name    string
		i       *Iam
		args    args
		fields  fields
		want    []*string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			got, err := i.ListRolePolicies(tt.args.ctx, tt.args.roleName)
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
		ctx        context.Context
		roleName   string
		policyName string
	}
	tests := []struct {
		name    string
		i       *Iam
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), roleName: "abcd-12", policyName: "e2fer34"},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), roleName: "abcd-12", policyName: "e2fer34"},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			err := i.DeleteRolePolicy(tt.args.ctx, tt.args.roleName, tt.args.policyName)
			if (err != nil) != tt.wantErr {
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
		ctx                 context.Context
		roleName            string
		instanceProfileName string
	}
	tests := []struct {
		name    string
		i       *Iam
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), roleName: "abcd-12", instanceProfileName: "e2fer34"},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), roleName: "abcd-12", instanceProfileName: "e2fer34"},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			err := i.RemoveRoleFromInstanceProfile(tt.args.ctx, tt.args.roleName, tt.args.instanceProfileName)
			if (err != nil) != tt.wantErr {
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
		ctx      context.Context
		roleName string
	}
	tests := []struct {
		name    string
		i       *Iam
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), roleName: "abcd-12"},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), roleName: "abcd-12"},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			err := i.DeleteRole(tt.args.ctx, tt.args.roleName)
			if (err != nil) != tt.wantErr {
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
		ctx                 context.Context
		instanceProfileName string
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
			args:    args{ctx: context.TODO(), instanceProfileName: "abcd-12"},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), instanceProfileName: "abcd-12"},
			fields:  fields{Service: newmockIAMClient(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockIAMClient(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := &Iam{
				Service: tt.fields.Service,
			}
			err := i.DeleteInstanceProfile(tt.args.ctx, tt.args.instanceProfileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Iam.DeleteInstanceProfile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
