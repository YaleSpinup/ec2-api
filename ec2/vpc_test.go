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

var testVpcs = []*ec2.Vpc{
	{
		CidrBlock: aws.String("10.1.2.0/22"),
		CidrBlockAssociationSet: []*ec2.VpcCidrBlockAssociation{
			{
				AssociationId: aws.String("vpc-cidr-assoc-aabbccdd"),
				CidrBlock:     aws.String("10.1.2.0/22"),
				CidrBlockState: &ec2.VpcCidrBlockState{
					State: aws.String("associated"),
				},
			},
		},
		DhcpOptionsId:   aws.String("dopt-01234567890aaa"),
		InstanceTenancy: aws.String("default"),
		IsDefault:       aws.Bool(false),
		OwnerId:         aws.String("012345678910"),
		State:           aws.String("available"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("Awesome VPC"),
			},
		},
		VpcId: aws.String("vpc-a1a1qa1a1"),
	},
}

func (m *mockEC2Client) DescribeVpcsWithContext(ctx context.Context, input *ec2.DescribeVpcsInput, opts ...request.Option) (*ec2.DescribeVpcsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}
	if input.VpcIds == nil {
		return &ec2.DescribeVpcsOutput{Vpcs: testVpcs}, nil
	}

	vpcs := []*ec2.Vpc{}
	for _, v := range testVpcs {
		for _, id := range input.VpcIds {
			if aws.StringValue(id) == aws.StringValue(v.VpcId) {
				vpcs = append(vpcs, v)
			}
		}
	}

	return &ec2.DescribeVpcsOutput{
		Vpcs: vpcs,
	}, nil
}

func TestEc2_ListVPCs(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []map[string]string
		wantErr bool
	}{
		{
			name:   "list vpcs",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args:   args{ctx: context.TODO()},
			want: []map[string]string{
				{
					"id": "vpc-a1a1qa1a1",
				},
			},
		},
		{
			name: "aws error",
			args: args{
				ctx: context.TODO(),
			},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
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
			got, err := e.ListVPCs(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ListVPCs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ListVPCs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_GetVpcById(t *testing.T) {
	type fields struct {
		session         *session.Session
		Service         ec2iface.EC2API
		DefaultKMSKeyId string
		DefaultSgs      []string
		DefaultSubnets  []string
		org             string
	}
	type args struct {
		ctx context.Context
		id  string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *ec2.Vpc
		wantErr bool
	}{
		{
			name:   "valid vpc",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args:   args{ctx: context.TODO(), id: "vpc-a1a1qa1a1"},
			want:   testVpcs[0],
		},
		{
			name:    "unknown vpc",
			fields:  fields{Service: newmockEC2Client(t, nil)},
			args:    args{ctx: context.TODO(), id: "vpc-a33323334"},
			wantErr: true,
		},
		{
			name: "aws error",
			args: args{
				ctx: context.TODO(),
			},
			fields:  fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
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
			got, err := e.GetVPCByID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ListVPCs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ListVPCs() = %v, want %v", got, tt.want)
			}
		})
	}
}
