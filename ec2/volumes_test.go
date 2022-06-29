package ec2

import (
	"context"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

func (m mockEC2Client) CreateVolumeWithContext(ctx context.Context, input *ec2.CreateVolumeInput, opts ...request.Option) (*ec2.Volume, error) {
	if m.err != nil {
		return nil, m.err
	}

	// return nil Volume (unexpected)
	if aws.StringValue(input.VolumeType) == "weird" {
		return nil, nil
	}

	return &ec2.Volume{
		VolumeId: aws.String("vol-0123456789abcdef0"),
	}, nil
}

func (m mockEC2Client) DeleteVolumeWithContext(ctx context.Context, input *ec2.DeleteVolumeInput, opts ...request.Option) (*ec2.DeleteVolumeOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.DeleteVolumeOutput{}, nil
}

func (m mockEC2Client) ModifyVolumeWithContext(ctx context.Context, input *ec2.ModifyVolumeInput, opts ...request.Option) (*ec2.ModifyVolumeOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &ec2.ModifyVolumeOutput{VolumeModification: &ec2.VolumeModification{StatusMessage: aws.String("completed")}}, nil
}

func (m mockEC2Client) DetachVolumeWithContext(ctx context.Context, input *ec2.DetachVolumeInput, opts ...request.Option) (*ec2.VolumeAttachment, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.VolumeAttachment{
			VolumeId: aws.String("Volume-123")},
		nil
}

func (m mockEC2Client) AttachVolumeWithContext(ctx context.Context, input *ec2.AttachVolumeInput, opts ...request.Option) (*ec2.VolumeAttachment, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &ec2.VolumeAttachment{
			VolumeId: aws.String("Volume-123")},
		nil
}

func TestEc2_CreateVolume(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx   context.Context
		input *ec2.CreateVolumeInput
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ec2.Volume
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "unexpected output",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				input: &ec2.CreateVolumeInput{
					AvailabilityZone: aws.String("us-east-1"),
					VolumeType:       aws.String("weird"),
				},
			},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				input: &ec2.CreateVolumeInput{
					AvailabilityZone: aws.String("us-east-1"),
					VolumeType:       aws.String("gp3"),
				},
			},
			want: &ec2.Volume{
				VolumeId: aws.String("vol-0123456789abcdef0"),
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			got, err := e.CreateVolume(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.CreateVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.CreateVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_DeleteVolume(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx   context.Context
		input string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "nil input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name: "good input",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx:   context.TODO(),
				input: "vol-0123456789abcdef0",
			},
		},
		{
			name: "aws err",
			fields: fields{
				Service: newmockEC2Client(t, awserr.New("BadRequest", "boom", nil)),
			},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session:         tt.fields.session,
				Service:         tt.fields.Service,
				DefaultKMSKeyId: tt.fields.DefaultKMSKeyId,
				DefaultSgs:      tt.fields.DefaultSgs,
				DefaultSubnets:  tt.fields.DefaultSubnets,
				org:             tt.fields.org,
			}
			err := e.DeleteVolume(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.DeleteVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestEc2_ModifyVolume(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.ModifyVolumeInput
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ec2.VolumeModification
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.ModifyVolumeInput{Iops: aws.Int64(1234), VolumeType: aws.String("v-123"), Size: aws.Int64(456), VolumeId: aws.String("id-123")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			want:    &ec2.VolumeModification{StatusMessage: aws.String("completed")},
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.ModifyVolumeInput{Iops: aws.Int64(1234), VolumeType: aws.String("v-123"), Size: aws.Int64(456), VolumeId: aws.String("id-123")}},
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
			got, err := e.ModifyVolume(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ModifyVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ModifyVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_DetachVolume(t *testing.T) {
	type fields struct {
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.DetachVolumeInput
	}
	tests := []struct {
		name    string
		args    args
		fields  fields
		want    string
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.DetachVolumeInput{InstanceId: aws.String("v-123"), Force: aws.Bool(true), VolumeId: aws.String("id-123")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			want:    "Volume-123",
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.DetachVolumeInput{InstanceId: aws.String("v-123"), Force: aws.Bool(true), VolumeId: aws.String("id-123")}},
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
			got, err := e.DetachVolume(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.DetachVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.DetachVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_AttachVolume(t *testing.T) {
	type fields struct {
		//session *session.Session
		Service ec2iface.EC2API
	}
	type args struct {
		ctx   context.Context
		input *ec2.AttachVolumeInput
	}
	tests := []struct {
		name    string
		e       *Ec2
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "success case",
			args:    args{ctx: context.TODO(), input: &ec2.AttachVolumeInput{Device: aws.String("1234ad"), InstanceId: aws.String("51454"), VolumeId: aws.String("534")}},
			fields:  fields{Service: newmockEC2Client(t, nil)},
			want:    "Volume-123",
			wantErr: false,
		},
		{
			name:    "aws error",
			args:    args{ctx: context.TODO(), input: &ec2.AttachVolumeInput{Device: aws.String("1234ad"), InstanceId: aws.String("51454"), VolumeId: aws.String("534")}},
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
			got, err := e.AttachVolume(tt.args.ctx, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.AttachVolume() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Ec2.AttachVolume() = %v, want %v", got, tt.want)
			}
		})
	}
}
