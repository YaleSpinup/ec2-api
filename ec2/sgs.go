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

	out, err := e.Service.DescribeSecurityGroupsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
		Filters: filters,
	})
	if err != nil {
		return nil, ErrCode("listing security groups", err)
	}

	log.Debugf("returning list of %d security groups", len(out.SecurityGroups))

	list := make([]map[string]*string, len(out.SecurityGroups))
	for i, s := range out.SecurityGroups {
		tags := s.Tags
		sgName := aws.StringValue(s.GroupName)

		// Loop through the tags and if Name exist, set the sgName value to it
		for _, t := range tags {
			if aws.StringValue(t.Key) == "Name" {
				sgName = aws.StringValue(t.Value)
				break
			}
		}

		list[i] = map[string]*string{
			aws.StringValue(s.GroupId): &sgName,
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

	out, err := e.Service.DescribeSecurityGroupsWithContext(ctx, &ec2.DescribeSecurityGroupsInput{
		GroupIds: aws.StringSlice(ids),
	})
	if err != nil {
		return nil, ErrCode("getting details for security groups", err)
	}

	log.Debugf("returning security groups: %+v", out.SecurityGroups)

	return out.SecurityGroups, err
}

// DeleteSecurityGroup deletes the given security group
func (e *Ec2) DeleteSecurityGroup(ctx context.Context, id string) error {
	if id == "" {
		return apierror.New(apierror.ErrBadRequest, "invalid input", nil)
	}

	log.Infof("deleting security group %s", id)

	if _, err := e.Service.DeleteSecurityGroupWithContext(ctx, &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(id),
	}); err != nil {
		return ErrCode("deleting security group", err)
	}

	return nil
}
