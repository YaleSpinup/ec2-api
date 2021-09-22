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

var testSubnets = []*ec2.Subnet{
	{
		AssignIpv6AddressOnCreation: aws.Bool(false),
		AvailabilityZone:            aws.String("us-east-1a"),
		AvailabilityZoneId:          aws.String("use1-az4"),
		AvailableIpAddressCount:     aws.Int64(843),
		CidrBlock:                   aws.String("10.1.32.0/22"),
		DefaultForAz:                aws.Bool(false),
		MapCustomerOwnedIpOnLaunch:  aws.Bool(false),
		MapPublicIpOnLaunch:         aws.Bool(false),
		OwnerId:                     aws.String("012345678901"),
		State:                       aws.String("available"),
		SubnetArn:                   aws.String("arn:aws:ec2:us-east-1:012345678901:subnet/subnet-aabbccdd"),
		SubnetId:                    aws.String("subnet-aabbccdd"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-private-sub1"),
			},
		},
		VpcId: aws.String("vpc-a1a1qa1a1"),
	},
	{
		AssignIpv6AddressOnCreation: aws.Bool(false),
		AvailabilityZone:            aws.String("us-east-1d"),
		AvailabilityZoneId:          aws.String("use1-az2"),
		AvailableIpAddressCount:     aws.Int64(789),
		CidrBlock:                   aws.String("10.1.36.0/22"),
		DefaultForAz:                aws.Bool(false),
		MapCustomerOwnedIpOnLaunch:  aws.Bool(false),
		MapPublicIpOnLaunch:         aws.Bool(false),
		OwnerId:                     aws.String("012345678901"),
		State:                       aws.String("available"),
		SubnetArn:                   aws.String("arn:aws:ec2:us-east-1:012345678901:subnet/subnet-eeffgghh"),
		SubnetId:                    aws.String("subnet-eeffgghh"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-private-sub2"),
			},
		},
		VpcId: aws.String("vpc-a1a1qa1a1"),
	},
	{
		AssignIpv6AddressOnCreation: aws.Bool(false),
		AvailabilityZone:            aws.String("us-east-1d"),
		AvailabilityZoneId:          aws.String("use1-az2"),
		AvailableIpAddressCount:     aws.Int64(996),
		CidrBlock:                   aws.String("10.1.44.0/22"),
		DefaultForAz:                aws.Bool(false),
		MapCustomerOwnedIpOnLaunch:  aws.Bool(false),
		MapPublicIpOnLaunch:         aws.Bool(false),
		OwnerId:                     aws.String("012345678901"),
		State:                       aws.String("available"),
		SubnetArn:                   aws.String("arn:aws:ec2:us-east-1:012345678901:subnet/subnet-iijjkkll"),
		SubnetId:                    aws.String("subnet-iijjkkll"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-DMZ-sub2"),
			},
		},
		VpcId: aws.String("vpc-a1a1qa1a1"),
	},
	{
		AssignIpv6AddressOnCreation: aws.Bool(false),
		AvailabilityZone:            aws.String("us-east-1a"),
		AvailabilityZoneId:          aws.String("use1-az4"),
		AvailableIpAddressCount:     aws.Int64(996),
		CidrBlock:                   aws.String("10.1.40.0/22"),
		DefaultForAz:                aws.Bool(false),
		MapCustomerOwnedIpOnLaunch:  aws.Bool(false),
		MapPublicIpOnLaunch:         aws.Bool(false),
		OwnerId:                     aws.String("012345678901"),
		State:                       aws.String("available"),
		SubnetArn:                   aws.String("arn:aws:ec2:us-east-1:012345678901:subnet/subnet-mmnnoopp"),
		SubnetId:                    aws.String("subnet-mmnnoopp"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("test-DMZ-sub1"),
			},
		},
		VpcId: aws.String("vpc-a1a1qa1a1"),
	},
	{
		AssignIpv6AddressOnCreation: aws.Bool(false),
		AvailabilityZone:            aws.String("us-east-1a"),
		AvailabilityZoneId:          aws.String("use1-az4"),
		AvailableIpAddressCount:     aws.Int64(124),
		CidrBlock:                   aws.String("192.168.1.0/24"),
		DefaultForAz:                aws.Bool(false),
		MapCustomerOwnedIpOnLaunch:  aws.Bool(false),
		MapPublicIpOnLaunch:         aws.Bool(false),
		OwnerId:                     aws.String("012345678901"),
		State:                       aws.String("available"),
		SubnetArn:                   aws.String("arn:aws:ec2:us-east-1:012345678901:subnet/subnet-mmnnoopp"),
		SubnetId:                    aws.String("subnet-mmnnoopp"),
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String("top-secret-1"),
			},
		},
		VpcId: aws.String("vpc-zzzzzzzz"),
	},
}

func (m *mockEC2Client) DescribeSubnetsWithContext(ctx context.Context, input *ec2.DescribeSubnetsInput, opts ...request.Option) (*ec2.DescribeSubnetsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	subnets := []*ec2.Subnet{}
	for _, s := range testSubnets {
		match := true
		for _, f := range input.Filters {
			switch aws.StringValue(f.Name) {
			case "vpc-id":
				var matchesVpc bool
				for _, v := range aws.StringValueSlice(f.Values) {
					if v == aws.StringValue(s.VpcId) {
						matchesVpc = true
						break
					}
				}

				if !matchesVpc {
					match = false
					break
				}
			case "state":
				var matchesState bool
				for _, v := range aws.StringValueSlice(f.Values) {
					if v == aws.StringValue(s.State) {
						matchesState = true
						break
					}
				}

				if !matchesState {
					match = false
					break
				}
			}
		}

		if match {
			subnets = append(subnets, s)
		}
	}

	return &ec2.DescribeSubnetsOutput{Subnets: subnets}, nil
}

func TestEc2_ListSubnets(t *testing.T) {
	allSubnets := make([]map[string]string, len(testSubnets))
	for i, s := range testSubnets {
		var name string
		for _, t := range s.Tags {
			if aws.StringValue(t.Key) == "Name" {
				name = aws.StringValue(t.Value)
				break
			}
		}

		allSubnets[i] = map[string]string{
			aws.StringValue(s.SubnetId): name,
		}
	}

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
		vpc string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []map[string]string
		wantErr bool
	}{
		{
			name:   "empty vpc",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
			},
			want: allSubnets,
		},
		{
			name:   "existing vpc",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				vpc: "vpc-zzzzzzzz",
			},
			want: []map[string]string{
				{
					"subnet-mmnnoopp": "top-secret-1",
				},
			},
		},
		{
			name:   "bad vpc",
			fields: fields{Service: newmockEC2Client(t, awserr.New("Bad Request", "boom.", nil))},
			args: args{
				ctx: context.TODO(),
				vpc: "abc123",
			},
			wantErr: true,
		},
		{
			name: "aws error",
			fields: fields{
				Service: newmockEC2Client(t, nil),
			},
			args: args{
				ctx: context.TODO(),
				vpc: "abc123",
			},
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
			got, err := e.ListSubnets(tt.args.ctx, tt.args.vpc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ListSubnets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ListSubnets() = %v, want %v", got, tt.want)
			}
		})
	}
}
