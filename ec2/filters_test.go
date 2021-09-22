package ec2

import (
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func Test_notTerminated(t *testing.T) {
	tests := []struct {
		name string
		want *ec2.Filter
	}{
		{
			name: "not terminated filter",
			want: &ec2.Filter{
				Name: aws.String("instance-state-name"),
				Values: aws.StringSlice(
					[]string{
						"pending",
						"running",
						"shutting-down",
						"stopping",
						"stopped",
					},
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := notTerminated(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("notTerminated() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inOrg(t *testing.T) {
	type args struct {
		org string
	}
	tests := []struct {
		name string
		args args
		want *ec2.Filter
	}{
		{
			name: "in foo org",
			args: args{org: "foo"},
			want: &ec2.Filter{
				Name: aws.String("tag:yale:org"),
				Values: aws.StringSlice(
					[]string{"foo"},
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inOrg(tt.args.org); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inOrg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_withInstanceId(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name string
		args args
		want *ec2.Filter
	}{
		{
			name: "with instance id",
			args: args{id: "i-abcdefg0123"},
			want: &ec2.Filter{
				Name: aws.String("tag:spinup:instanceid"),
				Values: aws.StringSlice(
					[]string{"i-abcdefg0123"},
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := withInstanceId(tt.args.id); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("withInstanceId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_inVpc(t *testing.T) {
	type args struct {
		vpc string
	}
	tests := []struct {
		name string
		args args
		want *ec2.Filter
	}{
		{
			name: "with vpc id",
			args: args{vpc: "vpc-abcdefg0123"},
			want: &ec2.Filter{
				Name: aws.String("vpc-id"),
				Values: aws.StringSlice(
					[]string{"vpc-abcdefg0123"},
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inVpc(tt.args.vpc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("inVpc() = %v, want %v", got, tt.want)
			}
		})
	}
}
