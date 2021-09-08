package ec2

import (
	"context"
	"github.com/YaleSpinup/apierror"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	log "github.com/sirupsen/logrus"
)

// ListSecurityGroups List all security groups in an aws account
func (e *Ec2) ListSecurityGroups(ctx context.Context, org string) ([]map[string]*string, error) {
	log.Infof("listing ec2 security groups (org: '%s')", org)

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
		tags := s.Tags
		var sgName *string

		// Loop through the tags and if Name exist, set the sgName value to it
		for _, t := range tags {
			if *t.Key == "Name" {
				sgName = t.Value
			}
		}

		// If sgName is nil, use the GroupName on the security group as a fallback
		if sgName == nil {
			sgName = s.GroupName
		}

		list[i] = map[string]*string{
			*s.GroupId: sgName,
		}
	}

	return list, err
}

// GetSecurityGroup Get the given security groups by a list of ids
func (e *Ec2) GetSecurityGroup(ctx context.Context, ids ...string) ([]*ec2.SecurityGroup, error) {
	if len(ids) == 0 {
		return nil, apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("getting details about security group ids %+v", ids)

	input := ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice(ids),
	}

	out, err := e.Service.DescribeSecurityGroupsWithContext(ctx, &input)
	if err != nil {
		return nil, ErrCode("getting details for security groups", err)
	}

	log.Debugf("returning security groups: %+v", out.SecurityGroups)

	return out.SecurityGroups, err
}
