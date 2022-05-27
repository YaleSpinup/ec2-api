package ec2

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func (m mockEC2Client) CreateSnapshotWithContext(ctx aws.Context, input *ec2.CreateSnapshotInput, opts ...request.Option) (*ec2.Snapshot, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.Snapshot{SnapshotId: aws.String("1234")}, nil
}
func TestEc2_CreateSnapshot(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.CreateSnapshotInput
	}
	tests := []struct {
		name    string
		fields  fields
		e       *Ec2
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.CreateSnapshotInput{VolumeId: aws.String("1234")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			want:    "1234",
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.CreateSnapshotInput{VolumeId: aws.String("1234")}},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO()},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			got, err := e.CreateSnapshot(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.CreateSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ec2.CreateSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}
