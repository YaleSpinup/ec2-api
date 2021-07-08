package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

func notTerminated() *ec2.Filter {
	return &ec2.Filter{
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
	}
}

func inOrg(org string) *ec2.Filter {
	return &ec2.Filter{
		// TODO: switch to "tag:spinup:org"
		// Name: aws.String("tag:spinup:org"),
		Name: aws.String("tag:yale:org"),
		Values: aws.StringSlice(
			[]string{org},
		),
	}
}

func withInstanceId(id string) *ec2.Filter {
	return &ec2.Filter{
		Name: aws.String("tag:spinup:instanceid"),
		Values: aws.StringSlice(
			[]string{id},
		),
	}
}
