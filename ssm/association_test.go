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
	mockDocName = "doc_341"
)

func (m *mockSSMClient) DescribeAssociationWithContext(ctx context.Context, inp *ssm.DescribeAssociationInput, _ ...request.Option) (*ssm.DescribeAssociationOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if *inp.InstanceId != mockInstanceId || *inp.Name != mockDocName {
		return nil, errors.New("mockssmclient: unknown instance id or doc name")
	}
	return &ssm.DescribeAssociationOutput{
		AssociationDescription: &ssm.AssociationDescription{
			Name:       inp.Name,
			InstanceId: inp.InstanceId,
		},
	}, nil
}

func TestSSM_DescribeAssociation(t *testing.T) {
	type fields struct {
		session *session.Session
		Service ssmiface.SSMAPI
	}
	type args struct {
		ctx        context.Context
		instanceId string
		docName    string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ssm.DescribeAssociationOutput
		wantErr bool
	}{
		{
			name:   "valid input",
			fields: fields{Service: newMockSSMClient(t, nil)},
			args:   args{ctx: context.TODO(), instanceId: mockInstanceId, docName: mockDocName},
			want: &ssm.DescribeAssociationOutput{
				AssociationDescription: &ssm.AssociationDescription{
					Name:       aws.String(mockDocName),
					InstanceId: aws.String(mockInstanceId),
				},
			},
		},
		{
			name:    "valid input, error from aws",
			fields:  fields{Service: newMockSSMClient(t, awserr.New("Bad Request", "boom.", nil))},
			args:    args{ctx: context.TODO(), instanceId: mockInstanceId, docName: mockDocName},
			wantErr: true,
		},
		{
			name:    "missing instance id",
			fields:  fields{Service: newMockSSMClient(t, nil)},
			args:    args{ctx: context.TODO(), docName: mockDocName},
			wantErr: true,
		},
		{
			name:    "unknown instance id",
			fields:  fields{Service: newMockSSMClient(t, nil)},
			args:    args{ctx: context.TODO(), instanceId: "xyz", docName: mockDocName},
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
			args:    args{ctx: context.TODO(), instanceId: mockInstanceId, docName: "xyz"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSM{
				session: tt.fields.session,
				Service: tt.fields.Service,
			}
			got, err := s.DescribeAssociation(tt.args.ctx, tt.args.instanceId, tt.args.docName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SSM.DescribeAssociation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.DescribeAssociation() = %v, want %v", got, tt.want)
			}
		})
	}
}
