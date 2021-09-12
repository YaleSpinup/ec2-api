package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"reflect"
	"testing"
)

var securityGroups = []*ec2.SecurityGroup{
	{
		Description: aws.String("This is a test description - SecurityGroup - sg-0000000001"),
		GroupId:     aws.String("sg-0000000001"),
		GroupName:   aws.String("foo1"),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(6000),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000001"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000001"),
						PrefixListId: aws.String("pre"),
					},
				},
				ToPort: aws.Int64(6001),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
						GroupId:                aws.String("sg-000000020"),
						GroupName:              aws.String("some-name1"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000020"),
						VpcId:                  aws.String("vpc-0000000020"),
						VpcPeeringConnectionId: aws.String("vpc-0000000021"),
					},
				},
			},
		},
		IpPermissionsEgress: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(0),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-f8iskag235fs2100"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-f8iskag235fs2100"),
						PrefixListId: aws.String("epre"),
					},
				},
				ToPort: aws.Int64(0),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
						GroupId:                aws.String("sg-000000020"),
						GroupName:              aws.String("some-name1"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000020"),
						VpcId:                  aws.String("vpc-0000000020"),
						VpcPeeringConnectionId: aws.String("vpc-0000000021"),
					},
				},
			},
		},
		OwnerId: aws.String("0000000001"),
		Tags:    nil,
		VpcId:   aws.String("vpc-0000000001"),
	},
	{
		Description: aws.String("This is a test description - SecurityGroup - sg-0000000002"),
		GroupId:     aws.String("sg-0000000002"),
		GroupName:   aws.String("foo2"),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(6002),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000002"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000002"),
						PrefixListId: aws.String("pre"),
					},
				},
				ToPort: aws.Int64(6003),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000002"),
						GroupId:                aws.String("sg-000000022"),
						GroupName:              aws.String("some-name2"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000021"),
						VpcId:                  aws.String("vpc-0000000022"),
						VpcPeeringConnectionId: aws.String("vpc-0000000023"),
					},
				},
			},
		},
		IpPermissionsEgress: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(0),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000002"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000002"),
						PrefixListId: aws.String("epre"),
					},
				},
				ToPort: aws.Int64(0),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000002"),
						GroupId:                aws.String("sg-000000022"),
						GroupName:              aws.String("some-name2"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000021"),
						VpcId:                  aws.String("vpc-0000000022"),
						VpcPeeringConnectionId: aws.String("vpc-0000000023"),
					},
				},
			},
		},
		OwnerId: aws.String("0000000002"),
		Tags:    nil,
		VpcId:   aws.String("vpc-0000000002"),
	},
	{
		Description: aws.String("This is a test description - SecurityGroup - sg-0000000003"),
		GroupId:     aws.String("sg-0000000003"),
		GroupName:   aws.String("foo3"),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(6004),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000003"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000003"),
						PrefixListId: aws.String("pre"),
					},
				},
				ToPort: aws.Int64(6005),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000003"),
						GroupId:                aws.String("sg-000000023"),
						GroupName:              aws.String("some-name3"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000022"),
						VpcId:                  aws.String("vpc-0000000024"),
						VpcPeeringConnectionId: aws.String("vpc-0000000025"),
					},
				},
			},
		},
		IpPermissionsEgress: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(0),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000003"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000003"),
						PrefixListId: aws.String("epre"),
					},
				},
				ToPort: aws.Int64(0),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000003"),
						GroupId:                aws.String("sg-000000023"),
						GroupName:              aws.String("some-name3"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000022"),
						VpcId:                  aws.String("vpc-0000000024"),
						VpcPeeringConnectionId: aws.String("vpc-0000000025"),
					},
				},
			},
		},
		OwnerId: aws.String("0000000003"),
		Tags:    nil,
		VpcId:   aws.String("vpc-0000000003"),
	},
	{
		Description: aws.String("This is a test description - SecurityGroup - sg-0000000004"),
		GroupId:     aws.String("sg-0000000004"),
		GroupName:   aws.String("foo4"),
		IpPermissions: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(6005),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000004"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000004"),
						PrefixListId: aws.String("pre"),
					},
				},
				ToPort: aws.Int64(6006),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000004"),
						GroupId:                aws.String("sg-000000023"),
						GroupName:              aws.String("some-name4"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000023"),
						VpcId:                  aws.String("vpc-0000000026"),
						VpcPeeringConnectionId: aws.String("vpc-0000000027"),
					},
				},
			},
		},
		IpPermissionsEgress: []*ec2.IpPermission{
			{
				FromPort:   aws.Int64(0),
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp:      aws.String("0.0.0.0/0"),
						Description: aws.String("This is a test description - IpRange - sg-0000000004"),
					},
				},
				Ipv6Ranges: nil,
				PrefixListIds: []*ec2.PrefixListId{
					{
						Description:  aws.String("This is a test description - PrefixListId - sg-0000000004"),
						PrefixListId: aws.String("epre"),
					},
				},
				ToPort: aws.Int64(0),
				UserIdGroupPairs: []*ec2.UserIdGroupPair{
					{
						Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000004"),
						GroupId:                aws.String("sg-000000023"),
						GroupName:              aws.String("some-name4"),
						PeeringStatus:          aws.String("active"),
						UserId:                 aws.String("0000000023"),
						VpcId:                  aws.String("vpc-0000000026"),
						VpcPeeringConnectionId: aws.String("vpc-0000000027"),
					},
				},
			},
		},
		OwnerId: aws.String("0000000004"),
		Tags:    nil,
		VpcId:   aws.String("vpc-0000000004"),
	},
}

func (m mockEC2Client) DescribeSecurityGroupsWithContext(ctx context.Context, input *ec2.DescribeSecurityGroupsInput, opts ...request.Option) (*ec2.DescribeSecurityGroupsOutput, error) {
	if m.err != nil {
		return nil, m.err
	}

	securityGroupList := []*ec2.SecurityGroup{}
	for _, securityGroup := range securityGroups {
		m.t.Logf("testing security group %s against filters", aws.StringValue(securityGroup.GroupId))

		if input.GroupIds != nil {
			m.t.Logf("checking pass security group %+v", input.GroupIds)

			var match bool
			for _, id := range input.GroupIds {
				m.t.Logf("id: %s | securityGroup.GroupId: %s", aws.StringValue(id), aws.StringValue(securityGroup.GroupId))
				if aws.StringValue(id) == aws.StringValue(securityGroup.GroupId) {
					match = true
					break
				}
			}

			if !match {
				continue
			}
		}

		m.t.Logf("security group %s matches filters", aws.StringValue(securityGroup.GroupId))

		securityGroupList = append(securityGroupList, securityGroup)
	}

	return &ec2.DescribeSecurityGroupsOutput{SecurityGroups: securityGroupList}, nil
}

func TestEc2_ListSecurityGroups(t *testing.T) {
	type fields struct {
		session *session.Session
		Service ec2iface.EC2API
		org     string
	}
	type args struct {
		ctx context.Context
		org string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []map[string]*string
		wantErr bool
	}{
		{
			name:   "matches list",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				org: "",
			},
			want: []map[string]*string{
				{
					"sg-0000000001": aws.String("foo1"),
				},
				{
					"sg-0000000002": aws.String("foo2"),
				},
				{
					"sg-0000000003": aws.String("foo3"),
				},
				{
					"sg-0000000004": aws.String("foo4"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session: tt.fields.session,
				Service: tt.fields.Service,
				org:     tt.fields.org,
			}
			got, err := e.ListSecurityGroups(tt.args.ctx, tt.args.org)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.ListSecurityGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.ListSecurityGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEc2_GetSecurityGroup(t *testing.T) {
	type fields struct {
		session *session.Session
		Service ec2iface.EC2API
		org     string
	}
	type args struct {
		ctx context.Context
		ids []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*ec2.SecurityGroup
		wantErr bool
	}{
		{
			name:    "nil ids",
			fields:  fields{Service: newmockEC2Client(t, nil)},
			args:    args{ctx: context.TODO()},
			wantErr: true,
		},
		{
			name:   "empty ids",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{},
			},
			wantErr: true,
		},
		{
			name:   "one id",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{"sg-0000000001"},
			},
			want: []*ec2.SecurityGroup{
				{
					Description: aws.String("This is a test description - SecurityGroup - sg-0000000001"),
					GroupId:     aws.String("sg-0000000001"),
					GroupName:   aws.String("foo1"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(6000),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000001"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000001"),
									PrefixListId: aws.String("pre"),
								},
							},
							ToPort: aws.Int64(6001),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
									GroupId:                aws.String("sg-000000020"),
									GroupName:              aws.String("some-name1"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000020"),
									VpcId:                  aws.String("vpc-0000000020"),
									VpcPeeringConnectionId: aws.String("vpc-0000000021"),
								},
							},
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(0),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-f8iskag235fs2100"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-f8iskag235fs2100"),
									PrefixListId: aws.String("epre"),
								},
							},
							ToPort: aws.Int64(0),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
									GroupId:                aws.String("sg-000000020"),
									GroupName:              aws.String("some-name1"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000020"),
									VpcId:                  aws.String("vpc-0000000020"),
									VpcPeeringConnectionId: aws.String("vpc-0000000021"),
								},
							},
						},
					},
					OwnerId: aws.String("0000000001"),
					Tags:    nil,
					VpcId:   aws.String("vpc-0000000001"),
				},
			},
		},
		{
			name:   "one id not owned by self",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{"sg-0000000005"},
			},
			want: []*ec2.SecurityGroup{},
		},
		{
			name:   "multiple ids",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{
					"sg-0000000001",
					"sg-0000000003",
					"sg-0000000004",
				},
			},
			want: []*ec2.SecurityGroup{
				{
					Description: aws.String("This is a test description - SecurityGroup - sg-0000000001"),
					GroupId:     aws.String("sg-0000000001"),
					GroupName:   aws.String("foo1"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(6000),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000001"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000001"),
									PrefixListId: aws.String("pre"),
								},
							},
							ToPort: aws.Int64(6001),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
									GroupId:                aws.String("sg-000000020"),
									GroupName:              aws.String("some-name1"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000020"),
									VpcId:                  aws.String("vpc-0000000020"),
									VpcPeeringConnectionId: aws.String("vpc-0000000021"),
								},
							},
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(0),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-f8iskag235fs2100"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-f8iskag235fs2100"),
									PrefixListId: aws.String("epre"),
								},
							},
							ToPort: aws.Int64(0),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
									GroupId:                aws.String("sg-000000020"),
									GroupName:              aws.String("some-name1"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000020"),
									VpcId:                  aws.String("vpc-0000000020"),
									VpcPeeringConnectionId: aws.String("vpc-0000000021"),
								},
							},
						},
					},
					OwnerId: aws.String("0000000001"),
					Tags:    nil,
					VpcId:   aws.String("vpc-0000000001"),
				},
				{
					Description: aws.String("This is a test description - SecurityGroup - sg-0000000003"),
					GroupId:     aws.String("sg-0000000003"),
					GroupName:   aws.String("foo3"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(6004),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000003"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000003"),
									PrefixListId: aws.String("pre"),
								},
							},
							ToPort: aws.Int64(6005),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000003"),
									GroupId:                aws.String("sg-000000023"),
									GroupName:              aws.String("some-name3"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000022"),
									VpcId:                  aws.String("vpc-0000000024"),
									VpcPeeringConnectionId: aws.String("vpc-0000000025"),
								},
							},
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(0),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000003"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000003"),
									PrefixListId: aws.String("epre"),
								},
							},
							ToPort: aws.Int64(0),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000003"),
									GroupId:                aws.String("sg-000000023"),
									GroupName:              aws.String("some-name3"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000022"),
									VpcId:                  aws.String("vpc-0000000024"),
									VpcPeeringConnectionId: aws.String("vpc-0000000025"),
								},
							},
						},
					},
					OwnerId: aws.String("0000000003"),
					Tags:    nil,
					VpcId:   aws.String("vpc-0000000003"),
				},
				{
					Description: aws.String("This is a test description - SecurityGroup - sg-0000000004"),
					GroupId:     aws.String("sg-0000000004"),
					GroupName:   aws.String("foo4"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(6005),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000004"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000004"),
									PrefixListId: aws.String("pre"),
								},
							},
							ToPort: aws.Int64(6006),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000004"),
									GroupId:                aws.String("sg-000000023"),
									GroupName:              aws.String("some-name4"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000023"),
									VpcId:                  aws.String("vpc-0000000026"),
									VpcPeeringConnectionId: aws.String("vpc-0000000027"),
								},
							},
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(0),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000004"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000004"),
									PrefixListId: aws.String("epre"),
								},
							},
							ToPort: aws.Int64(0),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000004"),
									GroupId:                aws.String("sg-000000023"),
									GroupName:              aws.String("some-name4"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000023"),
									VpcId:                  aws.String("vpc-0000000026"),
									VpcPeeringConnectionId: aws.String("vpc-0000000027"),
								},
							},
						},
					},
					OwnerId: aws.String("0000000004"),
					Tags:    nil,
					VpcId:   aws.String("vpc-0000000004"),
				},
			},
		},
		{
			name:   "multiple ids with some not owned by self",
			fields: fields{Service: newmockEC2Client(t, nil)},
			args: args{
				ctx: context.TODO(),
				ids: []string{
					"sg-0000000001",
					"sg-0000000004",
					"sg-0000000005",
				},
			},
			want: []*ec2.SecurityGroup{
				{
					Description: aws.String("This is a test description - SecurityGroup - sg-0000000001"),
					GroupId:     aws.String("sg-0000000001"),
					GroupName:   aws.String("foo1"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(6000),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000001"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000001"),
									PrefixListId: aws.String("pre"),
								},
							},
							ToPort: aws.Int64(6001),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
									GroupId:                aws.String("sg-000000020"),
									GroupName:              aws.String("some-name1"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000020"),
									VpcId:                  aws.String("vpc-0000000020"),
									VpcPeeringConnectionId: aws.String("vpc-0000000021"),
								},
							},
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(0),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-f8iskag235fs2100"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-f8iskag235fs2100"),
									PrefixListId: aws.String("epre"),
								},
							},
							ToPort: aws.Int64(0),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000001"),
									GroupId:                aws.String("sg-000000020"),
									GroupName:              aws.String("some-name1"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000020"),
									VpcId:                  aws.String("vpc-0000000020"),
									VpcPeeringConnectionId: aws.String("vpc-0000000021"),
								},
							},
						},
					},
					OwnerId: aws.String("0000000001"),
					Tags:    nil,
					VpcId:   aws.String("vpc-0000000001"),
				},
				{
					Description: aws.String("This is a test description - SecurityGroup - sg-0000000004"),
					GroupId:     aws.String("sg-0000000004"),
					GroupName:   aws.String("foo4"),
					IpPermissions: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(6005),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000004"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000004"),
									PrefixListId: aws.String("pre"),
								},
							},
							ToPort: aws.Int64(6006),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000004"),
									GroupId:                aws.String("sg-000000023"),
									GroupName:              aws.String("some-name4"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000023"),
									VpcId:                  aws.String("vpc-0000000026"),
									VpcPeeringConnectionId: aws.String("vpc-0000000027"),
								},
							},
						},
					},
					IpPermissionsEgress: []*ec2.IpPermission{
						{
							FromPort:   aws.Int64(0),
							IpProtocol: aws.String("-1"),
							IpRanges: []*ec2.IpRange{
								{
									CidrIp:      aws.String("0.0.0.0/0"),
									Description: aws.String("This is a test description - IpRange - sg-0000000004"),
								},
							},
							Ipv6Ranges: nil,
							PrefixListIds: []*ec2.PrefixListId{
								{
									Description:  aws.String("This is a test description - PrefixListId - sg-0000000004"),
									PrefixListId: aws.String("epre"),
								},
							},
							ToPort: aws.Int64(0),
							UserIdGroupPairs: []*ec2.UserIdGroupPair{
								{
									Description:            aws.String("This is a test description - UserIdGroupPair - sg-0000000004"),
									GroupId:                aws.String("sg-000000023"),
									GroupName:              aws.String("some-name4"),
									PeeringStatus:          aws.String("active"),
									UserId:                 aws.String("0000000023"),
									VpcId:                  aws.String("vpc-0000000026"),
									VpcPeeringConnectionId: aws.String("vpc-0000000027"),
								},
							},
						},
					},
					OwnerId: aws.String("0000000004"),
					Tags:    nil,
					VpcId:   aws.String("vpc-0000000004"),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Ec2{
				session: tt.fields.session,
				Service: tt.fields.Service,
			}
			got, err := e.GetSecurityGroup(tt.args.ctx, tt.args.ids...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Ec2.GetSecurityGroup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Ec2.GetSecurityGroup() = %v, want %v", got, tt.want)
			}
		})
	}
}
