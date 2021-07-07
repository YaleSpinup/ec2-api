package api

import (
	"reflect"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Test_toEc2InstanceResponse(t *testing.T) {
	testTime, err := time.Parse("2006-01-02 15:04:05 -0700 MST", "2020-10-09 23:55:57 +0000 UTC")
	if err != nil {
		t.Errorf("failed to parse time: %s", err)
	}

	type args struct {
		instance *ec2.Instance
	}
	tests := []struct {
		name string
		args args
		want *Ec2InstanceResponse
	}{
		{
			name: "nil input",
			args: args{instance: nil},
		},
		{
			name: "full example",
			args: args{
				instance: &ec2.Instance{
					AmiLaunchIndex: aws.Int64(0),
					Architecture:   aws.String("x86_64"),
					BlockDeviceMappings: []*ec2.InstanceBlockDeviceMapping{
						{
							DeviceName: aws.String("/dev/sda1"),
							Ebs: &ec2.EbsInstanceBlockDevice{
								AttachTime:          aws.Time(testTime),
								DeleteOnTermination: aws.Bool(true),
								Status:              aws.String("attached"),
								VolumeId:            aws.String("vol-0000000000000"),
							},
						},
						{
							DeviceName: aws.String("xvdf"),
							Ebs: &ec2.EbsInstanceBlockDevice{
								AttachTime:          aws.Time(testTime),
								DeleteOnTermination: aws.Bool(true),
								Status:              aws.String("attached"),
								VolumeId:            aws.String("vol-0f0f0f0f0f0f0f"),
							},
						},
					},
					CapacityReservationSpecification: &ec2.CapacityReservationSpecificationResponse{
						CapacityReservationPreference: aws.String("open"),
					},
					ClientToken: aws.String("9a2549e2-3547-48ce-a65b-5f6045dfd38a"),
					CpuOptions: &ec2.CpuOptions{
						CoreCount:      aws.Int64(16),
						ThreadsPerCore: aws.Int64(2),
					},
					EbsOptimized: aws.Bool(false),
					EnaSupport:   aws.Bool(true),
					EnclaveOptions: &ec2.EnclaveOptions{
						Enabled: aws.Bool(false),
					},
					HibernationOptions: &ec2.HibernationOptions{
						Configured: aws.Bool(false),
					},
					Hypervisor: aws.String("xen"),
					IamInstanceProfile: &ec2.IamInstanceProfile{
						Arn: aws.String("arn:aws:iam::888888888888:instance-profile/windowsInstance_default"),
						Id:  aws.String("1117777777666"),
					},
					ImageId:      aws.String("ami-a1a1a1a1a1a1a1"),
					InstanceId:   aws.String("i-pppppppoooooooo"),
					InstanceType: aws.String("g3.8xlarge"),
					KeyName:      aws.String("yaleits_rd_ec2_key"),
					LaunchTime:   aws.Time(testTime),
					MetadataOptions: &ec2.InstanceMetadataOptionsResponse{
						HttpEndpoint:            aws.String("enabled"),
						HttpPutResponseHopLimit: aws.Int64(1),
						HttpTokens:              aws.String("optional"),
						State:                   aws.String("applied"),
					},
					Monitoring: &ec2.Monitoring{
						State: aws.String("disabled"),
					},
					NetworkInterfaces: []*ec2.InstanceNetworkInterface{
						{
							Attachment: &ec2.InstanceNetworkInterfaceAttachment{
								AttachTime:          aws.Time(testTime),
								AttachmentId:        aws.String("eni-attach-aaaaaaaaaaavvvvvvvvv"),
								DeleteOnTermination: aws.Bool(true),
								DeviceIndex:         aws.Int64(0),
								NetworkCardIndex:    aws.Int64(0),
								Status:              aws.String("attached"),
							},
							Description: aws.String(""),
							Groups: []*ec2.GroupIdentifier{
								{
									GroupId:   aws.String("sg-00112233445566"),
									GroupName: aws.String("spdev-000575"),
								},
							},
							InterfaceType:      aws.String("interface"),
							MacAddress:         aws.String("11:11:11:11:11:11"),
							NetworkInterfaceId: aws.String("eni-aaaaxxxxvvvv"),
							OwnerId:            aws.String("888888888888"),
							PrivateDnsName:     aws.String("ip-10-1-2-34.ec2.internal"),
							PrivateIpAddress:   aws.String("10.1.2.34"),
							PrivateIpAddresses: []*ec2.InstancePrivateIpAddress{
								{
									Primary:          aws.Bool(true),
									PrivateDnsName:   aws.String("ip-10-1-2-34.ec2.internal"),
									PrivateIpAddress: aws.String("10.1.2.34"),
								},
							},
							SourceDestCheck: aws.Bool(true),
							Status:          aws.String("in-use"),
							SubnetId:        aws.String("subnet-aabbccddee"),
							VpcId:           aws.String("vpc-0987654321"),
						},
					},
					Placement: &ec2.Placement{
						AvailabilityZone: aws.String("us-east-1d"),
						GroupName:        aws.String(""),
						Tenancy:          aws.String("default"),
					},
					Platform:         aws.String("windows"),
					PrivateDnsName:   aws.String("ip-10-1-2-34.ec2.internal"),
					PrivateIpAddress: aws.String("10.1.2.34"),
					PublicDnsName:    aws.String(""),
					RootDeviceName:   aws.String("/dev/sda1"),
					RootDeviceType:   aws.String("ebs"),
					SecurityGroups: []*ec2.GroupIdentifier{
						{
							GroupId:   aws.String("sg-00112233445566"),
							GroupName: aws.String("spdev-000575"),
						},
					},
					SourceDestCheck: aws.Bool(true),
					State: &ec2.InstanceState{
						Code: aws.Int64(16),
						Name: aws.String("running"),
					},
					StateTransitionReason: aws.String(""),
					SubnetId:              aws.String("subnet-aabbccddee"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("Name"),
							Value: aws.String("spdev-0013ee.spdev.yale.edu"),
						},
						{
							Key:   aws.String("yale:org"),
							Value: aws.String("dev"),
						},
						{
							Key:   aws.String("yale:osfamily"),
							Value: aws.String("windows"),
						},
						{
							Key:   aws.String("spinup:spaceid"),
							Value: aws.String("spdev-000575"),
						},
						{
							Key:   aws.String("yale:size"),
							Value: aws.String("g3.8xlarge"),
						},
						{
							Key:   aws.String("yale:created_at"),
							Value: aws.String("2020/10/09 23:56:57"),
						},
						{
							Key:   aws.String("yale:subsidized"),
							Value: aws.String("false"),
						},
						{
							Key:   aws.String("yale:created_by"),
							Value: aws.String("ba123"),
						},
						{
							Key:   aws.String("CreatedBy"),
							Value: aws.String("ab123"),
						},
					},
					VirtualizationType: aws.String("hvm"),
					VpcId:              aws.String("vpc-0987654321"),
				},
			},
			want: &Ec2InstanceResponse{
				Az:        "us-east-1d",
				CreatedAt: "2020/10/09 23:55:57",
				CreatedBy: "ba123",
				ID:        "i-pppppppoooooooo",
				Image:     "ami-a1a1a1a1a1a1a1",
				Ip:        "10.1.2.34",
				Key:       "yaleits_rd_ec2_key",
				Name:      "spdev-0013ee.spdev.yale.edu",
				Platform:  "windows",
				Sgs: []map[string]string{
					{
						"sg-00112233445566": "spdev-000575",
					},
				},
				State:  "running",
				Subnet: "subnet-aabbccddee",
				Tags: []map[string]string{
					{"Name": "spdev-0013ee.spdev.yale.edu"},
					{"yale:org": "dev"},
					{"yale:osfamily": "windows"},
					{"spinup:spaceid": "spdev-000575"},
					{"yale:size": "g3.8xlarge"},
					{"yale:created_at": "2020/10/09 23:56:57"},
					{"yale:subsidized": "false"},
					{"yale:created_by": "ba123"},
					{"CreatedBy": "ab123"},
				},
				Type: "g3.8xlarge",
				Volumes: map[string]*Volume{
					"vol-0000000000000": {
						AttachTime:          "2020-10-09 23:55:57 UTC",
						DeleteOnTermination: true,
						Status:              "attached",
						DeviceName:          "/dev/sda1",
					},
					"vol-0f0f0f0f0f0f0f": {
						AttachTime:          "2020-10-09 23:55:57 UTC",
						DeleteOnTermination: true,
						Status:              "attached",
						DeviceName:          "xvdf",
					},
				},
			},
		},
		{
			name: "default to createdBy tag if yale:created_at is missing",
			args: args{
				instance: &ec2.Instance{
					Platform: aws.String("winders"),
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("CreatedBy"),
							Value: aws.String("ab123"),
						},
					},
				},
			},
			want: &Ec2InstanceResponse{
				CreatedBy: "ab123",
				Platform:  "winders",
				Sgs:       []map[string]string{},
				Tags: []map[string]string{
					{"CreatedBy": "ab123"},
				},
				Volumes: map[string]*Volume{},
			},
		},
		{
			name: "default to linux tag if platform is missing",
			args: args{
				instance: &ec2.Instance{
					Tags: []*ec2.Tag{
						{
							Key:   aws.String("yale:created_by"),
							Value: aws.String("ba321"),
						},
					},
				},
			},
			want: &Ec2InstanceResponse{
				CreatedBy: "ba321",
				Platform:  "linux",
				Sgs:       []map[string]string{},
				Tags: []map[string]string{
					{"yale:created_by": "ba321"},
				},
				Volumes: map[string]*Volume{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toEc2InstanceResponse(tt.args.instance); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toEc2InstanceResponse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_toEc2VolumeResponse(t *testing.T) {
	type args struct {
		volume *ec2.Volume
	}
	tests := []struct {
		name string
		args args
		want *Ec2VolumeResponse
	}{
		{
			name: "nil input",
			args: args{volume: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toEc2VolumeResponse(tt.args.volume); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toEc2VolumeResponse() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func Test_toEC2SnapshotResponse(t *testing.T) {
	type args struct {
		snapshot *ec2.Snapshot
	}
	tests := []struct {
		name string
		args args
		want *Ec2SnapshotResponse
	}{
		{
			name: "nil input",
			args: args{snapshot: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toEC2SnapshotResponse(tt.args.snapshot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toEC2SnapshotResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
