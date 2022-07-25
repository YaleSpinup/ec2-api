package ec2

import (
	"context"
	"fmt"
	"reflect"
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

func (m mockEC2Client) DeleteSnapshotWithContext(aws.Context, *ec2.DeleteSnapshotInput, ...request.Option) (*ec2.DeleteSnapshotOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DeleteSnapshotOutput{}, nil
}

func (m mockEC2Client) DescribeSnapshotsWithContext(aws.Context, *ec2.DescribeSnapshotsInput, ...request.Option) (*ec2.DescribeSnapshotsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.DescribeSnapshotsOutput{Snapshots: []*ec2.Snapshot{{SnapshotId: aws.String("snap-123")}, {SnapshotId: aws.String("snap-456")}}}, nil
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

func TestEc2_DeleteSnapshot(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.DeleteSnapshotInput
	}
	tests := []struct {
		name    string
		fields  fields
		e       *Ec2
		args    args
		wantErr bool
	}{
		{
			name: "success case",
			args: args{ctx: context.TODO(), input: &ec2.DeleteSnapshotInput{
				SnapshotId: aws.String("3514")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.DeleteSnapshotInput{SnapshotId: aws.String("1234")}},
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
			if err := e.DeleteSnapshot(tt.args.ctx, tt.args.input); (err != nil) != tt.wantErr {
				t.Errorf("Ec2.DeleteSnapshot() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEc2_ListSnapshots(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.DescribeSnapshotsInput
	}
	tests := []struct {
		name    string
		e       *Ec2
		args    args
		fields  fields
		want    *ec2.DescribeSnapshotsOutput
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.DescribeSnapshotsInput{Filters: []*ec2.Filter{{Name: aws.String("description"), Values: aws.StringSlice([]string{fmt.Sprintf("*for %s from vol*", "i-123")})}}}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			want:    &ec2.DescribeSnapshotsOutput{Snapshots: []*ec2.Snapshot{{SnapshotId: aws.String("snap-123")}, {SnapshotId: aws.String("snap-456")}}},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.DescribeSnapshotsInput{Filters: []*ec2.Filter{{Name: aws.String("description"), Values: aws.StringSlice([]string{fmt.Sprintf("*for %s from vol*", "i-123")})}}}},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			wantErr: true,
		},
		{
			name:    "nil input",
			args:    args{ctx: context.TODO(), input: nil},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				Service: tt.fields.Service,
			}
			got, err := e.ListSnapshots(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ListSnapshots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ListSnapshots() = %v, want %v", got, tt.want)
			}
		})
	}
}
