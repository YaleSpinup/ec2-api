package ec2

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

func (e *Ec2) ListSecurityGroups(ctx context.Context, org string, name string) ([]map[string]*string, error) {
	log.Infof("listing ec2 security groups (name: '%s', org: '%s')", name, org)

	var filters []*ec2.Filter
	if org != "" {
		filters = []*ec2.Filter{inOrg(org)}
	}

	input := ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice([]string{}),
		Filters:  filters,
	}

	out, err := e.Service.DescribeSecurityGroupsWithContext(ctx, &input)
	if err != nil {
		return nil, ErrCode("listing security groups", err)
	}

	log.Debugf("returning list of %d snapshots", len(out.SecurityGroups))

	list := make([]map[string]*string, len(out.SecurityGroups))
	for i, s := range out.SecurityGroups {
		list[i] = map[string]*string{
			*s.GroupId: s.GroupName,
		}
	}

	return list, err
}
