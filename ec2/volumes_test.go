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
