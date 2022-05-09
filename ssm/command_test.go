package ssm

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
)

const (
	mockInstanceId = "instance_id_341"
	mockCommandId  = "command_id_567"
)

func (m *mockSSMClient) GetCommandInvocationWithContext(ctx context.Context, inp *ssm.GetCommandInvocationInput, _ ...request.Option) (*ssm.GetCommandInvocationOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if *inp.InstanceId != mockInstanceId || *inp.CommandId != mockCommandId {
		return nil, errors.New("mockssmclient: unknown ids")
	}
	return &ssm.GetCommandInvocationOutput{
		CommandId:  inp.CommandId,
		InstanceId: inp.InstanceId,
	}, nil
}

func (m *mockSSMClient) SendCommandWithContext(ctx aws.Context, inp *ssm.SendCommandInput, opt ...request.Option) (*ssm.SendCommandOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ssm.SendCommandOutput{
		Command: &ssm.Command{CommandId: aws.String("Command-123")},
	}, nil

}

func TestSSM_GetCommandInvocation(t *testing.T) {
	type fields struct {
		session *session.Session
		Service ssmiface.SSMAPI
	}
	type args struct {
		ctx        context.Context
		instanceId string
		commandId  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ssm.GetCommandInvocationOutput
		wantErr bool
	}{
		{
			name:   "valid input",
			fields: fields{Service: newMockSSMClient(t, nil)},
			args:   args{ctx: context.TODO(), instanceId: mockInstanceId, commandId: mockCommandId},
			want: &ssm.GetCommandInvocationOutput{
				CommandId:  aws.String(mockCommandId),
				InstanceId: aws.String(mockInstanceId),
			},
		},
		{
			name:    "valid input, error from aws",
			fields:  fields{Service: newMockSSMClient(t, awserr.New("Bad Request", "boom.", nil))},
			args:    args{ctx: context.TODO(), instanceId: mockInstanceId, commandId: mockCommandId},
			wantErr: true,
		},
		{
			name:    "missing instance id",
			fields:  fields{Service: newMockSSMClient(t, nil)},
			args:    args{ctx: context.TODO(), commandId: mockCommandId},
			wantErr: true,
		},
		{
			name:    "unknown instance id",
			fields:  fields{Service: newMockSSMClient(t, nil)},
			args:    args{ctx: context.TODO(), instanceId: "xyz", commandId: mockCommandId},
			wantErr: true,
		},
		{
			name:    "missing command id",
			fields:  fields{Service: newMockSSMClient(t, nil)},
			args:    args{ctx: context.TODO(), instanceId: mockInstanceId},
			wantErr: true,
		},
		{
			name:    "unknown command id",
			fields:  fields{Service: newMockSSMClient(t, nil)},
			args:    args{ctx: context.TODO(), instanceId: mockInstanceId, commandId: "xyz"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSM{
				session: tt.fields.session,
				Service: tt.fields.Service,
			}
			got, err := s.GetCommandInvocation(tt.args.ctx, tt.args.instanceId, tt.args.commandId)
			if (err != nil) != tt.wantErr {
				t.Errorf("SSM.GetCommandInvocation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.GetCommandInvocation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSSM_SendCommand(t *testing.T) {
	type fields struct {
		session *session.Session
		Service ssmiface.SSMAPI
	}
	type args struct {
		ctx   context.Context
		input *ssm.SendCommandInput
	}
	tests := []struct {
		name    string
		fields  fields
		s       *SSM
		args    args
		want    *ssm.Command
		wantErr bool
	}{
		{
			name:   "valid input",
			fields: fields{Service: newMockSSMClient(t, nil)},
			args:   args{ctx: context.TODO(), input: &ssm.SendCommandInput{}},
			want:   &ssm.Command{CommandId: aws.String("Command-123")},
		},
		{
			name:    "valid input, aws error",
			fields:  fields{Service: newMockSSMClient(t, errors.New("some error"))},
			args:    args{ctx: context.TODO(), input: &ssm.SendCommandInput{}},
			wantErr: true,
		},
		{
			name:    "invalid input",
			fields:  fields{Service: newMockSSMClient(t, errors.New("some error"))},
			args:    args{ctx: context.TODO(), input: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSM{
				session: tt.fields.session,
				Service: tt.fields.Service,
			}
			got, err := s.SendCommand(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SSM.SendCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SSM.SendCommand() = %v, want %v", got, tt.want)
			}
		})
	}
}
